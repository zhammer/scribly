import base64
import binascii
import logging
from dataclasses import dataclass
from typing import Callable, Optional

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
from scribly.definitions import Context
from scribly.exceptions import AuthError
from scribly.use_scribly import Scribly

logger = logging.getLogger(__name__)


class ScriblyMiddleware:
    def __init__(self, app: ASGIApp):
        self.app = app

    async def __call__(self, scope: Scope, receive: Receive, send: Send) -> None:
        if not scope["type"] == "http":
            return await self.app(scope, receive, send)

        connection_pool = scope["app"].state.connection_pool
        if not connection_pool:
            raise RuntimeError("Requires an app with a connection pool")

        async with connection_pool.acquire() as connection:
            database = Database(connection)
            context = Context(database)
            scope["scribly"] = Scribly(context)

            return await self.app(scope, receive, send)


class BasicAuthBackend(AuthenticationBackend):
    # from https://www.starlette.io/authentication/
    async def authenticate(self, request):
        if "Authorization" not in request.headers:
            return

        auth = request.headers["Authorization"]
        try:
            scheme, credentials = auth.split()
            if scheme.lower() != "basic":
                return
            decoded = base64.b64decode(credentials).decode("ascii")
        except (ValueError, UnicodeDecodeError, binascii.Error) as exc:
            logger.error("Invalid basic auth credentials: %s", exc)
            raise AuthenticationError("Invalid basic auth credentials")

        username, _, password = decoded.partition(":")
        scribly = request.scope["scribly"]
        try:
            user = await scribly.log_in(username, password)
        except AuthError as e:
            logger.info("bad login attempt from %s, err: %s", username, e)
            return AuthCredentials, UnauthenticatedUser()
        except Exception as e:
            logger.error("unknown error when logging in user %s, err: %s", username, e)

        logger.info("request from user %d, %s", user.id, user.username)
        return AuthCredentials(["authenticated"]), SimpleUser(user.username)
