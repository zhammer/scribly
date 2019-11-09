from contextlib import asynccontextmanager
from typing import AsyncIterator, Optional

import asyncpg

from scribly.definitions import DatabaseGateway, User
from scribly.exceptions import AuthError


class Database(DatabaseGateway):
    def __init__(self, connection: asyncpg.Connection) -> None:
        self.connection = connection

    async def fetch_user(self, username: str, password: str) -> User:
        row = await self.connection.fetchrow(
            "SELECT id, username FROM users WHERE username = $1 AND password = $2",
            username,
            password,
        )

        if not row:
            raise AuthError()

        return User(row["id"], row["username"])
