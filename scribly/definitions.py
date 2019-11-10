import abc
from dataclasses import dataclass
from typing import Optional, Sequence

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
    """The person whose turn it is to write. Only exists in `in_progress` state."""
    current_writers_turn: Optional[User]
    """All of the turns taken by writers to get to the story's current state."""
    turns: Sequence[Turn]


class DatabaseGateway(abc.ABC):
    @abc.abstractmethod
    async def fetch_user(self, username: str, password: str) -> User:
        ...

    @abc.abstractmethod
    async def start_story(self, user: User, title: str, body: str) -> Story:
        ...


@dataclass
class Context:
    database: DatabaseGateway
