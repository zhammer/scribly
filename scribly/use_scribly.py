from dataclasses import dataclass

from scribly.definitions import Context, User


@dataclass
class Scribly:
    context: Context

    async def log_in(self, username: str, password: str) -> User:
        return await self.context.database.fetch_user(username, password)
