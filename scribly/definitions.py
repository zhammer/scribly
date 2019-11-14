import abc
from contextlib import AbstractAsyncContextManager
from dataclasses import dataclass
from typing import AsyncIterator, Optional, Sequence

from typing_extensions import Literal


@dataclass
class User:
    id: int
    username: str


TurnAction = Literal["pass", "write", "finish", "write_and_finish"]


@dataclass
class Turn:
    taken_by: User
    action: TurnAction
    """Text written by the user on this turn. Only exists on `write` and `write_and_finish` actions."""
    text_written: Optional[str]


StoryState = Literal["draft", "in_progress", "done"]


@dataclass
class Story:
    id: int
    title: str
    state: StoryState
    """User that created the story."""
    created_by: User
    """In order list of users cowriting the story. Includes creator. Empty in `draft` state."""
    cowriters: Optional[Sequence[User]]
    """All of the turns taken by writers to get to the story's current state."""
    turns: Sequence[Turn]

    @property
    def current_writers_turn(self) -> Optional[User]:
        """The person whose turn it is to write. Only exists in `in_progress` state."""
        if not self.state == "in_progress":
            return None

        if not self.cowriters:
            raise RuntimeError(
                f"Story {id} in state {self.state} should have cowriters but has none."
            )

        num_turns = len(self.turns)
        current_writer_index = num_turns % len(self.cowriters)

        return self.cowriters[current_writer_index]


@dataclass
class Me:
    user: User
    stories: Sequence[Story]


class DatabaseGateway(abc.ABC):
    @abc.abstractmethod
    async def fetch_user(self, username: str, password: str) -> User:
        ...

    @abc.abstractmethod
    async def fetch_users(self, *, usernames: Sequence[str]) -> Sequence[User]:
        ...

    @abc.abstractmethod
    async def start_story(self, user: User, title: str, body: str) -> Story:
        ...

    @abc.abstractmethod
    async def add_cowriters(self, story: Story, cowriters: Sequence[User]) -> Story:
        ...

    @abc.abstractmethod
    async def fetch_story(self, story_id: int, *, for_update: bool = False) -> Story:
        ...

    @abc.abstractmethod
    async def fetch_me(self, user: User) -> Me:
        ...

    @abc.abstractmethod
    async def add_turn_pass(self, user: User, story: Story) -> Story:
        ...

    @abc.abstractmethod
    async def add_turn_finish(self, user: User, story: Story) -> Story:
        ...

    @abc.abstractmethod
    async def add_turn_write(
        self, user: User, story: Story, text_written: str
    ) -> Story:
        ...

    @abc.abstractmethod
    async def add_turn_write_and_finish(
        self, user: User, story: Story, text_written: str
    ) -> Story:
        ...

    @abc.abstractmethod
    def transaction(self) -> "AbstractAsyncContextManager[None]":
        ...


@dataclass
class Context:
    database: DatabaseGateway
