from typing import Sequence

from scribly.definitions import Story, User
from scribly.exceptions import AuthError, InputError


def require_user_can_access_story(user: User, story: Story) -> None:
    if story.state == "draft":
        if not user.id == story.created_by.id:
            raise AuthError(
                f"User {user.id} cannot view story {story.id} in state {story.state}. Only creator has access."
            )
        return

    if story.cowriters is None:
        raise RuntimeError(f"Story {story} should have cowriters.")

    cowriter_ids = {cowriter.id for cowriter in story.cowriters}
    if not user.id in cowriter_ids:
        raise AuthError(
            f"User {user.id} cannot view story {story.id} in state {story.state}. Only cowriters have access."
        )


def require_user_can_add_cowriters(
    user: User, story: Story, cowriter_usernames: Sequence[str]
) -> None:
    if not story.created_by.id == user.id:
        raise AuthError(
            f"User {user.id} cannot add cowriters to story {story.id} created by {story.created_by.id}"
        )

    if not story.state == "draft":
        raise RuntimeError(
            f"Story must be in state 'draft' to add cowriters. Story {story.id} is in state {story.state}."
        )

    if user.username in cowriter_usernames:
        raise InputError(
            f"You cannot list yourself as a cowriter. {user.username} is your username."
        )


def require_valid_cowriters(
    cowriters: Sequence[User], cowriter_usernames: Sequence[str]
) -> None:
    requested_cowriter_set = set(cowriter_usernames)
    fetched_cowriter_set = set([cowriter.username for cowriter in cowriters])

    not_found_cowriters = requested_cowriter_set - fetched_cowriter_set
    if not_found_cowriters:
        raise InputError(f"Could not find users {', '.join(not_found_cowriters)}")


def require_user_can_take_turn_pass(user: User, story: Story) -> None:
    _require_user_can_take_turn(user, story)


def require_user_can_take_turn_write(
    user: User, story: Story, text_written: str
) -> None:
    _require_user_can_take_turn(user, story)

    if text_written == "":
        raise InputError(f"Text for a `write` turn cannot be empty.")


def _require_user_can_take_turn(user: User, story: Story) -> None:
    if not user.id in {cowriter.id for cowriter in story.cowriters or []}:
        raise AuthError(
            f"User {user.id} cannot take turn on story {story.id} as they are not a cowriter."
        )

    if not story.state == "in_progress":
        raise RuntimeError(
            f"Turn cannot be taken for story {story.id} in state {story.state}."
        )

    if not user.id == story.current_writers_turn:
        raise RuntimeError(
            f"User {user.id} cannot take turn as it is user {story.current_writers_turn.id}'s turn."
        )
