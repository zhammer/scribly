from dataclasses import dataclass
from typing import Sequence

from scribly import policies
from scribly.definitions import Context, Story, TurnAction, User
from scribly.util import shuffle


@dataclass
class Scribly:
    context: Context

    async def log_in(self, username: str, password: str) -> User:
        return await self.context.database.fetch_user(username, password)

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

            full_cowriters = [user] + list(shuffle(cowriters))

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

