from dataclasses import dataclass

from scribly.definitions import Context, Story, User
from scribly import policies


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
