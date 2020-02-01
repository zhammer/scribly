import asyncio
import base64
import binascii
import logging
from dataclasses import dataclass
from typing import Callable, Optional

import aio_pika
import aiohttp
import asyncpg
from starlette.authentication import (
    AuthCredentials,
    AuthenticationBackend,
    AuthenticationError,
    SimpleUser,
    UnauthenticatedUser,
)

from scribly.database import Database
from scribly.definitions import User
from scribly import env
from scribly.rabbit import Rabbit
from scribly.exceptions import AuthError
from scribly.sendgrid import SendGrid
from scribly.use_scribly import Scribly

logger = logging.getLogger(__name__)


class SessionAuthBackend(AuthenticationBackend):
    # from https://www.starlette.io/authentication/
    async def authenticate(self, request):
        session_user = request.session.get("user", None)
        if not session_user:
            return AuthCredentials, None

        try:
            user = User(
                id=session_user["id"],
                username=session_user["username"],
                email=session_user["email"],
                email_verification_status=session_user["email_verification_status"],
            )
        except KeyError as e:
            # in the case that the user structure updates and a user tries to visit the site
            # using an outdated session, we should clear out the session token and have them
            # log in again.
            logger.error(
                "Error plucking user from session user (%s). Error: %s.",
                session_user,
                e,
            )
            return AuthCredentials, None

        return (AuthCredentials(["authenticated"]), user)
