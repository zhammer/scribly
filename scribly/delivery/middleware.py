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
from scribly.definitions import Context, User
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


class SessionAuthBackend(AuthenticationBackend):
    # from https://www.starlette.io/authentication/
    async def authenticate(self, request):
        user = request.session.get("user", None)
        if not user:
            return AuthCredentials, None

        return (
            AuthCredentials(["authenticated"]),
            User(
                id=user["id"],
                username=user["username"],
                email=user["email"],
                email_verified=user["email_verified"],
            ),
        )
