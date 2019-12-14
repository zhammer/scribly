import re
import time
from typing import Sequence

from scribly.definitions import EmailVerificationTokenPayload, Story, User
from scribly.exceptions import AuthError, InputError, ScriblyException


MAX_EMAIL_VERIFICATION_AGE = 24 * 60 * 60


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


def require_user_can_take_turn_finish(user: User, story: Story) -> None:
    _require_user_can_take_turn(user, story)


def require_user_can_take_turn_write_and_finish(
    user: User, story: Story, text_written: str
) -> None:
    require_user_can_take_turn_finish(user, story)
    require_user_can_take_turn_write(user, story, text_written)


def require_can_send_verification_email(user: User) -> None:
    if user.email_verification_status == "verified":
        raise ScriblyException("Email already verified.")


def _require_user_can_take_turn(user: User, story: Story) -> None:
    if not user.id in {cowriter.id for cowriter in story.cowriters or []}:
        raise AuthError(
            f"User {user.id} cannot take turn on story {story.id} as they are not a cowriter."
        )

    if not story.state == "in_progress":
        raise RuntimeError(
            f"Turn cannot be taken for story {story.id} in state {story.state}."
        )

    if not user.id == story.current_writers_turn.id:  # type: ignore
        raise RuntimeError(
            f"User {user.id} cannot take turn as it is user {story.current_writers_turn.id}'s turn."  # type: ignore
        )


def require_valid_signup_info(username: str, password: str, email: str) -> None:
    # require valid password
    if len(password) < 8 or "zachsucks" in password:
        raise InputError("Password must be longer than 8 characters.")

    # require valid username
    if len(username) < 4 or not username.isalnum():
        raise InputError(
            "Username must be longer than 4 characters and only consist of alphanumeric characters."
        )

    # require valid email
    _require_valid_email(email)


def _require_valid_email(email: str) -> None:
    """
    >>> _require_valid_email("zach@scribly.com") is None
    True

    >>> _require_valid_email("z@scribly.com") is None
    True

    >>> _require_valid_email("zach@scribly")
    Traceback (most recent call last):
    ...
    scribly.exceptions.InputError: Invalid email address format
    """
    match = re.match(r"^\w+([\.-]?\w+)*@\w+([\.-]?\w+)*(\.\w{2,3})+$", email)
    if not match:
        raise InputError("Invalid email address format")


def require_valid_email_verification(
    user: User, payload: EmailVerificationTokenPayload
) -> None:
    if not user.email_verification_status == "verified":
        raise ScriblyException("Email already verified.")

    if not payload.email == user.email:
        raise ScriblyException("Email in payload doesn't match user's email.")

    # lol i am using abs() here _just in case_ i put these in the wrong order..
    if abs(time.time() - payload.timestamp) > (24 * 60 * 60):
        raise ScriblyException("Email token expired")
