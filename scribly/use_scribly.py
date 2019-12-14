from dataclasses import dataclass
from typing import Sequence

from scribly import auth, emails, exceptions, policies
from scribly.definitions import Context, Me, Story, TurnAction, User
from scribly.util import shuffle


@dataclass
class Scribly:
    context: Context

    async def log_in(self, username: str, password: str) -> User:
        user, hash = await self.context.database.fetch_user_with_password_hash(username)

        if not auth.verify_password_hash(hash, password):
            raise exceptions.AuthError()

        if auth.check_needs_rehash(hash):
            rehashed_password = auth.hash_password(password)
            await self.context.database.update_password(user, rehashed_password)

        return user

    async def sign_up(self, username: str, password: str, email: str) -> User:
        policies.require_valid_signup_info(username, password, email)

        hash = auth.hash_password(password)
        return await self.context.database.add_user(username, hash, email)

    async def start_story(self, user: User, title: str, body: str) -> Story:
        async with self.context.database.transaction():
            return await self.context.database.start_story(user, title, body)

    async def get_story(self, user: User, story_id: int) -> Story:
        story = await self.context.database.fetch_story(story_id)

        policies.require_user_can_access_story(user, story)

        return story

    async def add_cowriters(
        self, user: User, story_id: int, cowriter_usernames: Sequence[str]
    ) -> Story:
        async with self.context.database.transaction():
            story = await self.context.database.fetch_story(story_id, for_update=True)

            policies.require_user_can_add_cowriters(user, story, cowriter_usernames)

            cowriters = await self.context.database.fetch_users(
                usernames=cowriter_usernames
            )

            policies.require_valid_cowriters(cowriters, cowriter_usernames)

            full_cowriters = [user] + list(cowriters)

            return await self.context.database.add_cowriters(story, full_cowriters)

    async def take_turn_pass(self, user: User, story_id: int) -> Story:
        async with self.context.database.transaction():
            story = await self.context.database.fetch_story(story_id, for_update=True)

            policies.require_user_can_take_turn_pass(user, story)

            return await self.context.database.add_turn_pass(user, story)

    async def take_turn_write(
        self, user: User, story_id: int, text_written: str
    ) -> Story:
        async with self.context.database.transaction():
            story = await self.context.database.fetch_story(story_id, for_update=True)

            policies.require_user_can_take_turn_write(user, story, text_written)

            return await self.context.database.add_turn_write(user, story, text_written)

    async def take_turn_finish(self, user: User, story_id: int) -> Story:
        async with self.context.database.transaction():
            story = await self.context.database.fetch_story(story_id, for_update=True)

            policies.require_user_can_take_turn_finish(user, story)

            return await self.context.database.add_turn_finish(user, story)

    async def take_turn_write_and_finish(
        self, user: User, story_id: int, text_written: str
    ) -> Story:
        async with self.context.database.transaction():
            story = await self.context.database.fetch_story(story_id, for_update=True)

            policies.require_user_can_take_turn_write_and_finish(
                user, story, text_written
            )

            return await self.context.database.add_turn_write_and_finish(
                user, story, text_written
            )

    async def get_me(self, user: User) -> Me:
        return await self.context.database.fetch_me(user)

    async def send_verification_email(self, user: User) -> None:
        policies.require_can_send_verification_email(user)
        verification_token = auth.build_email_verification_token(user)
        email = emails.build_email_verification_email(user, verification_token)
        await self.context.emailer.send_email(email)

    async def verify_email(self, email_verification_token: str) -> None:
        async with self.context.database.transaction():
            verification_token_payload = auth.parse_email_verification_token(
                email_verification_token
            )
            user = await self.context.database.fetch_user(
                verification_token_payload.user_id, for_update=True
            )
            policies.require_valid_email_verification(user, verification_token_payload)
            await self.context.database.update_email_verification_status(
                user, "verified"
            )
