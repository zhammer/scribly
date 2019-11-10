from contextlib import asynccontextmanager
from typing import AsyncIterator, Dict, Optional, Sequence

import asyncpg
from asyncpg import Record

from scribly.definitions import DatabaseGateway, Story, Turn, User
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

    async def start_story(self, user: User, title: str, body: str) -> Story:
        story_record = await self.connection.fetchrow(
            """
            INSERT INTO stories (title, state, created_by)
            VALUES ($1, 'draft', $2)
            RETURNING *;
            """,
            title,
            user.id,
        )
        story_id = story_record["id"]
        turn_record = await self.connection.fetchrow(
            """
            INSERT INTO turns (story_id, taken_by, action, text_written)
            VALUES ($1, $2, 'write', $3)
            RETURNING *;
            """,
            story_id,
            user.id,
            body,
        )
        return _pluck_story(story_record, [turn_record], {user.id: user})


def _pluck_story(
    story_record: Record, turn_records: Sequence[Record], user_by_id: Dict[int, User]
) -> Story:
    turns = [_pluck_turn(turn_record, user_by_id) for turn_record in turn_records]
    return Story(
        id=story_record["id"],
        title=story_record["title"],
        state=story_record["state"],
        created_by=user_by_id[story_record["created_by"]],
        cowriters=None,
        turns=turns,
    )


def _pluck_turn(turn_record: Record, user_by_id: Dict[int, User]) -> Turn:
    return Turn(
        taken_by=user_by_id[turn_record["taken_by"]],
        action=turn_record["action"],
        text_written=turn_record["text_written"],
    )
