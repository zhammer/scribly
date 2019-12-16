import asyncio
import base64
import binascii
import logging
from dataclasses import dataclass
from typing import Callable, Optional

import aiohttp
import asyncpg
from starlette.authentication import (
    AuthCredentials,
    AuthenticationBackend,
    AuthenticationError,
    SimpleUser,
    UnauthenticatedUser,
)
from starlette.types import ASGIApp, Message, Receive, Scope, Send

from scribly.database import Database
from scribly.definitions import Context, User
from scribly import env
from scribly.exceptions import AuthError
from scribly.sendgrid import SendGrid
from scribly.use_scribly import Scribly

logger = logging.getLogger(__name__)


@dataclass
class WaitForStartupCompleteMiddleware:
    """
    Middleware that waits for a startup_complete_event asyncio.Event to
    start handling http requests. (Temporary workaround until
    https://github.com/encode/starlette/issues/733 is resolved.)
    """

    def __init__(self, app: ASGIApp, startup_complete_event: asyncio.Event):
        self.app = app
        self.startup_complete_event = startup_complete_event

    async def __call__(self, scope: Scope, receive: Receive, send: Send) -> None:
        if not scope["type"] == "http":
            return await self.app(scope, receive, send)

        await self.startup_complete_event.wait()
        return await self.app(scope, receive, send)


class ScriblyMiddleware:
    def __init__(self, app: ASGIApp):
        self.app = app

    async def __call__(self, scope: Scope, receive: Receive, send: Send) -> None:
        if not scope["type"] == "http":
            return await self.app(scope, receive, send)

        connection_pool = scope["app"].state.connection_pool
        if not connection_pool:
            raise RuntimeError("Requires an app with a connection pool")

        async with connection_pool.acquire() as connection, aiohttp.ClientSession() as sendgrid_session:
            database = Database(connection)
            emailer = SendGrid(
                env.SENDGRID_API_KEY, env.SENDGRID_BASE_URL, sendgrid_session
            )
            context = Context(database, emailer)
            scope["scribly"] = Scribly(context)

            return await self.app(scope, receive, send)


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

