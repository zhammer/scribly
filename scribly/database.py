from contextlib import asynccontextmanager
from typing import AsyncIterator, Dict, Optional, Sequence

import asyncpg
from asyncpg import Record

from scribly.definitions import DatabaseGateway, Story, Turn, User
from scribly.exceptions import AuthError, StoryNotFound


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

    async def fetch_story(self, story_id: int) -> Story:
        story_record = await self.connection.fetchrow(
            """
            SELECT * FROM stories WHERE id = $1;
            """,
            story_id,
        )

        if not story_record:
            raise StoryNotFound(f"Could not find story {story_id}")

        turn_records = await self.connection.fetch(
            """
            SELECT * FROM turns WHERE story_id = $1 ORDER BY id;
            """,
            story_id,
        )

        cowriter_records = await self.connection.fetch(
            """
            SELECT u.* FROM story_cowriters sc
            JOIN users u on sc.user_id = u.id
            WHERE sc.story_id = $1
            ORDER BY sc.turn_index;
            """,
            story_id,
        )

        created_by = await self.connection.fetchrow(
            "SELECT * FROM users WHERE id = $1", story_record["created_by"]
        )

        return _pluck_story_existing(
            story_record, turn_records, created_by, cowriter_records
        )

    @asynccontextmanager
    async def transaction(self) -> AsyncIterator[None]:
        async with self.connection.transaction():
            yield


def _pluck_story_existing(
    story_record: Record,
    turn_records: Sequence[Record],
    created_by_record: Record,
    cowriter_records: Optional[Sequence[Record]] = None,
) -> Story:
    created_by = _pluck_user(created_by_record)
    if not cowriter_records:
        cowriters = None
        user_by_id = {created_by.id: created_by}
    else:
        cowriters = [
            _pluck_user(cowriter_record) for cowriter_record in cowriter_records
        ]
        user_by_id = {user.id: user for user in cowriters}
    turns = [_pluck_turn(turn_record, user_by_id) for turn_record in turn_records]
    return Story(
        id=story_record["id"],
        title=story_record["title"],
        state=story_record["state"],
        created_by=user_by_id[story_record["created_by"]],
        cowriters=cowriters,
        turns=turns,
    )


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


def _pluck_user(user_record: Record) -> User:
    return User(id=user_record["id"], username=user_record["username"])


def _pluck_turn(turn_record: Record, user_by_id: Dict[int, User]) -> Turn:
    return Turn(
        taken_by=user_by_id[turn_record["taken_by"]],
        action=turn_record["action"],
        text_written=turn_record["text_written"],
    )
