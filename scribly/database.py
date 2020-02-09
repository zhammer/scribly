from contextlib import asynccontextmanager
from dataclasses import replace
from typing import AsyncIterator, Dict, Optional, Sequence, Tuple

import aiosqlite

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
        VALUES (?, ?, ?, ?)
        """

    def __init__(self, connection: aiosqlite.Connection) -> None:
        self.connection = connection

    async def add_user(self, username: str, password: str, email: str) -> User:
        await self.connection.execute(
            """
            INSERT INTO users (username, password, email)
            VALUES (?, ?, ?);
            """,
            (username, password, email,),
        )
        cursor = await self.connection.execute(
            """
            SELECT * FROM users WHERE rowid = last_insert_rowid();
            """
        )
        row = await cursor.fetchone()
        user = _pluck_user(row)
        await self.connection.commit()
        return user

    async def update_password(self, user: User, password: str) -> None:
        await self.connection.execute(
            """
            UPDATE users SET password = ?, updated_at = CURRENT_TIMESTAMP
            WHERE id = ?;
            """,
            (password, user.id,),
        )

    async def fetch_user_with_password_hash(self, username: str) -> Tuple[User, str]:
        cursor = await self.connection.execute(
            "SELECT * FROM users WHERE username = ?", (username,)
        )
        row = await cursor.fetchone()

        if not row:
            raise AuthError()

        return _pluck_user(row), row["password"]

    async def fetch_users(self, *, usernames: Sequence[str]) -> Sequence[User]:
        # VVV hack app here with sql injection
        unsafe_names = ", ".join(f"'{name}'" for name in usernames)
        rows = await self.connection.execute_fetchall(
            f"SELECT * FROM users WHERE username IN ({unsafe_names})"
        )

        return [_pluck_user(row) for row in rows]

    async def add_cowriters(self, story: Story, cowriters: Sequence[User]) -> Story:
        await self.connection.executemany(
            """
            INSERT INTO story_cowriters (story_id, user_id, turn_index)
            VALUES (?, ?, ?);
            """,
            [
                (story.id, cowriter.id, index)
                for index, cowriter in enumerate(cowriters)
            ],
        )

        await self.connection.execute(
            """
            UPDATE stories SET state = 'in_progress', updated_at = CURRENT_TIMESTAMP
            WHERE id = ?;
            """,
            (story.id,),
        )

        return replace(story, state="in_progress", cowriters=cowriters)

    async def update_email_verification_status(
        self, user: User, status: EmailVerificationState
    ) -> User:
        await self.connection.execute(
            """
            UPDATE users SET email_verification_status = :status, updated_at = CURRENT_TIMESTAMP
            WHERE id = :id;
            """,
            {"status": status, "id": user.id},
        )
        return replace(user, email_verification_status=status)

    async def fetch_user(self, user_id: int, for_update: bool = False) -> User:
        cursor = await self.connection.execute(
            """
            SELECT * FROM users WHERE id = ?
            """,
            (user_id,),
        )
        user_record = await cursor.fetchone()
        if not user_record:
            raise ScriblyException(f"User {user_id} doesn't exist.")

        return _pluck_user(user_record)

    async def start_story(self, user: User, title: str, body: str) -> Story:
        await self.connection.execute(
            """
            INSERT INTO stories (title, state, created_by)
            VALUES (?, 'draft', ?);
            """,
            (title, user.id,),
        )
        cursor = await self.connection.execute(
            """
            SELECT * FROM stories WHERE id = last_insert_rowid();
            """
        )
        story_record = await cursor.fetchone()
        story_id = story_record["id"]
        await self.connection.execute(
            self.QUERY_INSERT_TURN, (story_id, user.id, "write", body,)
        )
        cursor = await self.connection.execute(
            """
            SELECT * FROM turns WHERE id = last_insert_rowid();
            """
        )
        turn_record = await cursor.fetchone()
        return _pluck_story(story_record, [turn_record], {user.id: user})

    async def hide_story(self, user: User, story: Story) -> None:
        await self.connection.execute(
            """
            INSERT INTO user_story_hides (user_id, story_id, hidden_status)
            VALUES (?, ?, 'hidden')
            ON CONFLICT (user_id, story_id) DO UPDATE
            SET hidden_status = 'hidden', updated_at = CURRENT_TIMESTAMP
            """,
            (user.id, story.id,),
        )

    async def unhide_story(self, user: User, story: Story) -> None:
        await self.connection.execute(
            """
            INSERT INTO user_story_hides (user_id, story_id, hidden_status)
            VALUES (?, ?, 'unhidden')
            ON CONFLICT (user_id, story_id) DO UPDATE
            SET hidden_status = 'unhidden', updated_at = CURRENT_TIMESTAMP
            """,
            (user.id, story.id,),
        )

    async def fetch_me(self, user: User) -> Me:
        """
        Slow quick implementation of this at the moment using self.fetch_story.
        """
        user = await self.fetch_user(user.id)
        story_records = await self.connection.execute_fetchall(
            """
            SELECT DISTINCT s.id FROM stories s
            LEFT JOIN story_cowriters sc on sc.story_id = s.id
            WHERE s.created_by = :user_id
            OR sc.user_id = :user_id;
            """,
            {"user_id": user.id},
        )
        story_ids = [story_record["id"] for story_record in story_records]

        # not using asyncio.gather because i don't know if that would work
        # with one db connection.
        stories = [await self.fetch_story(story_id) for story_id in story_ids]

        hidden_story_records = await self.connection.execute_fetchall(
            """
            SELECT story_id FROM user_story_hides
            WHERE user_id = ?
            AND hidden_status = 'hidden'
            """,
            (user.id,),
        )
        hidden_story_ids = frozenset(
            hidden_story_record["story_id"]
            for hidden_story_record in hidden_story_records
        )

        return Me(user=user, stories=stories, hidden_story_ids=hidden_story_ids)

    async def fetch_story(self, story_id: int, *, for_update: bool = False) -> Story:
        cursor = await self.connection.execute(
            """
            SELECT * FROM stories WHERE id = ?;
            """,
            (story_id,),
        )
        story_record = await cursor.fetchone()

        if not story_record:
            raise StoryNotFound(f"Could not find story {story_id}")

        turn_records = await self.connection.execute_fetchall(
            """
            SELECT * FROM turns WHERE story_id = ? ORDER BY id;
            """,
            (story_id,),
        )

        cowriter_records = await self.connection.execute_fetchall(
            """
            SELECT u.* FROM story_cowriters sc
            JOIN users u on sc.user_id = u.id
            WHERE sc.story_id = ?
            ORDER BY sc.turn_index;
            """,
            (story_id,),
        )

        cursor = await self.connection.execute(
            "SELECT * FROM users WHERE id = ?", (story_record["created_by"],)
        )
        created_by = await cursor.fetchone()

        return _pluck_story_existing(
            story_record, turn_records, created_by, cowriter_records
        )

    async def add_turn_pass(self, user: User, story: Story) -> Story:
        await self.connection.execute(
            self.QUERY_INSERT_TURN, (story.id, user.id, "pass", "")
        )
        cursor = await self.connection.execute(
            "SELECT * FROM turns WHERE rowid = last_insert_rowid()"
        )
        turn_record = await cursor.fetchone()
        turn = _pluck_turn(turn_record, {user.id: user})
        return replace(story, turns=story.turns + [turn])

    async def add_turn_write(
        self, user: User, story: Story, text_written: str
    ) -> Story:
        await self.connection.execute(
            self.QUERY_INSERT_TURN, (story.id, user.id, "write", text_written,)
        )
        cursor = await self.connection.execute(
            "SELECT * FROM turns WHERE rowid = last_insert_rowid()"
        )
        turn_record = await cursor.fetchone()
        turn = _pluck_turn(turn_record, {user.id: user})
        return replace(story, turns=story.turns + [turn])

    async def add_turn_finish(self, user: User, story: Story) -> Story:
        await self.connection.execute(
            self.QUERY_INSERT_TURN, (story.id, user.id, "finish", "",)
        )
        cursor = await self.connection.execute(
            "SELECT * FROM turns WHERE rowid = last_insert_rowid()"
        )
        turn_record = await cursor.fetchone()
        await self.connection.execute(
            """
            UPDATE stories SET state = 'done', updated_at = CURRENT_TIMESTAMP
            WHERE id = ?
            """,
            (story.id,),
        )
        turn = _pluck_turn(turn_record, {user.id: user})
        return replace(story, state="done", turns=story.turns + [turn])

    async def add_turn_write_and_finish(
        self, user: User, story: Story, text_written: str
    ) -> Story:
        await self.connection.execute(
            self.QUERY_INSERT_TURN,
            (story.id, user.id, "write_and_finish", text_written,),
        )
        cursor = await self.connection.execute(
            "SELECT * FROM turns WHERE rowid = last_insert_rowid()"
        )
        turn_record = await cursor.fetchone()
        await self.connection.execute(
            """
            UPDATE stories SET state = 'done', updated_at = CURRENT_TIMESTAMP
            WHERE id = ?
            """,
            (story.id,),
        )
        turn = _pluck_turn(turn_record, {user.id: user})
        return replace(story, state="done", turns=story.turns + [turn])

    @asynccontextmanager
    async def transaction(self) -> AsyncIterator[None]:
        await self.connection.execute("begin")
        try:
            yield
        except Exception as e:
            await self.connection.execute("rollback")
            raise e
        await self.connection.execute("commit")


def _pluck_story_existing(
    story_record: Dict,
    turn_records: Sequence[Dict],
    created_by_record: Dict,
    cowriter_records: Optional[Sequence[Dict]] = None,
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
    story_record: Dict, turn_records: Sequence[Dict], user_by_id: Dict[int, User]
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


def _pluck_user(user_record: aiosqlite.Row) -> User:
    return User(
        id=user_record["id"],
        username=user_record["username"],
        email=user_record["email"],
        email_verification_status=user_record["email_verification_status"],
    )


def _pluck_turn(turn_record: Dict, user_by_id: Dict[int, User]) -> Turn:
    return Turn(
        taken_by=user_by_id[turn_record["taken_by"]],
        action=turn_record["action"],
        text_written=turn_record["text_written"],
    )
