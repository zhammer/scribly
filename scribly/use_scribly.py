from dataclasses import dataclass

from scribly.definitions import Context, Story, User


@dataclass
class Scribly:
    context: Context

    async def log_in(self, username: str, password: str) -> User:
        return await self.context.database.fetch_user(username, password)

    async def start_story(self, user: User, title: str, body: str) -> Story:
        return await self.context.database.start_story(user, title, body)
