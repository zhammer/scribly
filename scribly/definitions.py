import abc
from contextlib import AbstractAsyncContextManager
from dataclasses import dataclass
from typing import FrozenSet, List, Optional, Sequence, Tuple

from typing_extensions import Literal

EmailVerificationState = Literal["pending", "verified"]


@dataclass
class User:
    id: int
    username: str
    email: str
    email_verification_status: EmailVerificationState


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
    turns: List[Turn]

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
class StoryWithUserMeta:
    """Story with some metadata that applies to a user's preferences on a story."""

    story: Story
    hidden: bool


@dataclass
class Me:
    user: User
    stories: Sequence[Story]
    hidden_story_ids: FrozenSet[int]

    @property
    def your_turn(self) -> Sequence[StoryWithUserMeta]:
        return [
            story_with_meta
            for story_with_meta in self.in_progress
            if story_with_meta.story.current_writers_turn.id == self.user.id  # type: ignore
        ]

    @property
    def waiting_for_others(self) -> Sequence[StoryWithUserMeta]:
        return [
            story_with_meta
            for story_with_meta in self.in_progress
            if not story_with_meta.story.current_writers_turn.id == self.user.id  # type: ignore
        ]

    @property
    def in_progress(self) -> Sequence[StoryWithUserMeta]:
        return [
            self._story_with_user_meta(story)
            for story in self.stories
            if story.state == "in_progress"
        ]

    @property
    def drafts(self) -> Sequence[StoryWithUserMeta]:
        return [
            self._story_with_user_meta(story)
            for story in self.stories
            if story.state == "draft"
        ]

    @property
    def done(self) -> Sequence[StoryWithUserMeta]:
        return [
            self._story_with_user_meta(story)
            for story in self.stories
            if story.state == "done"
        ]

    def _story_with_user_meta(self, story: Story) -> StoryWithUserMeta:
        hidden = story.id in self.hidden_story_ids
        return StoryWithUserMeta(story, hidden=hidden)


@dataclass
class Email:
    subject: str
    body: str
    to: User


@dataclass
class EmailVerificationTokenPayload:
    user_id: int
    email: str
    timestamp: float


class MessageGateway(abc.ABC):
    @abc.abstractmethod
    async def announce_user_created(self, user: User) -> None:
        ...

    @abc.abstractmethod
    async def announce_turn_taken(self, story: Story) -> None:
        ...

    @abc.abstractmethod
    async def announce_cowriters_added(self, story: Story) -> None:
        ...


class DatabaseGateway(abc.ABC):
    @abc.abstractmethod
    async def fetch_user_with_password_hash(self, username: str) -> Tuple[User, str]:
        ...

    @abc.abstractmethod
    async def add_user(self, username: str, password_hash: str, email: str) -> User:
        ...

    @abc.abstractmethod
    async def update_password(self, user: User, password_hash: str) -> None:
        ...

    @abc.abstractmethod
    async def fetch_user(self, user_id: int, for_update: bool) -> User:
        ...

    @abc.abstractmethod
    async def fetch_users(self, *, usernames: Sequence[str]) -> Sequence[User]:
        ...

    @abc.abstractmethod
    async def fetch_all_users(self) -> Sequence[User]:
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
    async def hide_story(self, user: User, story: Story) -> None:
        ...

    @abc.abstractmethod
    async def unhide_story(self, user: User, story: Story) -> None:
        ...

    @abc.abstractmethod
    def transaction(self) -> "AbstractAsyncContextManager[None]":
        ...

    @abc.abstractmethod
    async def update_email_verification_status(
        self, user: User, status: EmailVerificationState
    ) -> User:
        ...


class EmailGateway(abc.ABC):
    @abc.abstractmethod
    async def send_email(self, email: Email) -> None:
        ...
