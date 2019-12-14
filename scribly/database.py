from contextlib import asynccontextmanager
from typing import AsyncIterator, Dict, Optional, Sequence, Tuple

import asyncpg
from asyncpg import Record

from scribly.definitions import (
    DatabaseGateway,
    EmailVerificationState,
    Me,
    Story,
    Turn,
    User,
)
from scribly.exceptions import AuthError, ScriblyException, StoryNotFound


class Database(DatabaseGateway):
    QUERY_INSERT_TURN = """
        INSERT INTO turns (story_id, taken_by, action, text_written)
        VALUES ($1, $2, $3, $4)
        RETURNING *
        """

    def __init__(self, connection: asyncpg.Connection) -> None:
        self.connection = connection

    async def add_user(self, username: str, password: str, email: str) -> User:
        user = await self.connection.fetchrow(
            """
            INSERT INTO users (username, password, email)
            VALUES ($1, $2, $3)
            RETURNING *
            """,
            username,
            password,
            email,
        )
        return _pluck_user(user)

    async def update_password(self, user: User, password: str) -> None:
        await self.connection.execute(
            """
            UPDATE users SET password = $1, updated_at = NOW()
            WHERE id = $2;
            """,
            password,
            user.id,
        )

    async def fetch_user_with_password_hash(self, username: str) -> Tuple[User, str]:
        row = await self.connection.fetchrow(
            "SELECT * FROM users WHERE username = $1", username,
        )

        if not row:
            raise AuthError()

        return _pluck_user(row), row["password"]

    async def fetch_users(self, *, usernames: Sequence[str]) -> Sequence[User]:
        rows = await self.connection.fetch(
            "SELECT * FROM users WHERE username = any($1::text[])", usernames
        )

        return [_pluck_user(row) for row in rows]

    async def add_cowriters(self, story: Story, cowriters: Sequence[User]) -> Story:
        await self.connection.executemany(
            """
            INSERT INTO story_cowriters (story_id, user_id, turn_index)
            VALUES ($1, $2, $3);
            """,
            [
                (story.id, cowriter.id, index)
                for index, cowriter in enumerate(cowriters)
            ],
        )

        await self.connection.execute(
            """
            UPDATE stories SET state = 'in_progress', updated_at = NOW()
            WHERE id = $1;
            """,
            story.id,
        )

        return Story(
            id=story.id,
            title=story.title,
            state="in_progress",
            created_by=story.created_by,
            cowriters=cowriters,
            turns=story.turns,
        )

    async def update_email_verification_status(
        self, user: User, status: EmailVerificationState
    ) -> User:
        await self.connection.execute(
            """
            UPDATE users SET email_verification_status = $1, updated_at = NOW()
            WHERE id = $1;
            """,
            user.id,
        )
        return User(
            id=user.id,
            username=user.username,
            email=user.email_verification_status,
            email_verification_status=status,
        )

    async def fetch_user(self, user_id: int, for_update: bool = False) -> User:
        user_record = await self.connection.fetchrow(
            f"""
            SELECT * FROM users WHERE id = $1 {"FOR UPDATE" if for_update else ""}
            """,
            user_id,
        )
        if not user_record:
            raise ScriblyException(f"User {user_id} doesn't exist.")

        return _pluck_user(user_record)

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
            self.QUERY_INSERT_TURN, story_id, user.id, "write", body,
        )
        return _pluck_story(story_record, [turn_record], {user.id: user})

    async def fetch_me(self, user: User) -> Me:
        """
        Slow quick implementation of this at the moment using self.fetch_story.
        """
        story_records = await self.connection.fetch(
            """
            SELECT DISTINCT s.id FROM stories s
            LEFT JOIN story_cowriters sc on sc.story_id = s.id
            WHERE s.created_by = $1
            OR sc.user_id = $1;
            """,
            user.id,
        )
        story_ids = [story_record["id"] for story_record in story_records]

        # not using asyncio.gather because i don't know if that would work
        # with one db connection.
        stories = [await self.fetch_story(story_id) for story_id in story_ids]

        return Me(user=user, stories=stories)

    async def fetch_story(self, story_id: int, *, for_update: bool = False) -> Story:
        story_record = await self.connection.fetchrow(
            f"""
            SELECT * FROM stories WHERE id = $1 {"FOR UPDATE" if for_update else ""};
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

    async def add_turn_pass(self, user: User, story: Story) -> Story:
        turn_record = await self.connection.fetchrow(
            self.QUERY_INSERT_TURN, story.id, user.id, "pass", ""
        )
        turn = _pluck_turn(turn_record, {user.id: user})
        return Story(
            id=story.id,
            title=story.title,
            state=story.state,
            created_by=story.created_by,
            cowriters=story.cowriters,
            turns=(list(story.turns) + [turn]),
        )

    async def add_turn_write(
        self, user: User, story: Story, text_written: str
    ) -> Story:
        turn_record = await self.connection.fetchrow(
            self.QUERY_INSERT_TURN, story.id, user.id, "write", text_written,
        )
        turn = _pluck_turn(turn_record, {user.id: user})
        return Story(
            id=story.id,
            title=story.title,
            state=story.state,
            created_by=story.created_by,
            cowriters=story.cowriters,
            turns=(list(story.turns) + [turn]),
        )

    async def add_turn_finish(self, user: User, story: Story) -> Story:
        turn_record = await self.connection.fetchrow(
            self.QUERY_INSERT_TURN, story.id, user.id, "finish", "",
        )
        await self.connection.execute(
            """
            UPDATE stories SET state = 'done', updated_at = NOW()
            WHERE id = $1
            """,
            story.id,
        )
        turn = _pluck_turn(turn_record, {user.id: user})
        return Story(
            id=story.id,
            title=story.title,
            state="done",
            created_by=story.created_by,
            cowriters=story.cowriters,
            turns=(list(story.turns) + [turn]),
        )

    async def add_turn_write_and_finish(
        self, user: User, story: Story, text_written: str
    ) -> Story:
        turn_record = await self.connection.fetchrow(
            self.QUERY_INSERT_TURN, story.id, user.id, "write_and_finish", text_written,
        )
        await self.connection.execute(
            """
            UPDATE stories SET state = 'done', updated_at = NOW()
            WHERE id = $1
            """,
            story.id,
        )
        turn = _pluck_turn(turn_record, {user.id: user})
        return Story(
            id=story.id,
            title=story.title,
            state="done",
            created_by=story.created_by,
            cowriters=story.cowriters,
            turns=(list(story.turns) + [turn]),
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
    return User(
        id=user_record["id"],
        username=user_record["username"],
        email=user_record["email"],
        email_verification_status=user_record["email_verification_status"],
    )


def _pluck_turn(turn_record: Record, user_by_id: Dict[int, User]) -> Turn:
    return Turn(
        taken_by=user_by_id[turn_record["taken_by"]],
        action=turn_record["action"],
        text_written=turn_record["text_written"],
    )
