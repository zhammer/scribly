import base64
import binascii
import logging
import random
from typing import Tuple

from starlette.applications import Starlette
from starlette.authentication import (
    AuthCredentials,
    AuthenticationBackend,
    AuthenticationError,
    SimpleUser,
    UnauthenticatedUser,
)
from starlette.middleware.authentication import AuthenticationMiddleware
from starlette.responses import RedirectResponse, Response
from starlette.staticfiles import StaticFiles
from starlette.templating import Jinja2Templates

from scribly.delivery.constants import STORY_STARTERS

logger = logging.getLogger(__name__)


USERS = ("zach.the.hammer@gmail.com:password", "gsnussbaum@gmail.com:password")


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

        username, _, _ = decoded.partition(":")
        if decoded not in USERS:
            logger.info("bad login attempt from %s", username)
            return AuthCredentials, UnauthenticatedUser()

        logger.info("request from user %s", username)
        return AuthCredentials(["authenticated"]), SimpleUser(username)


templates = Jinja2Templates(directory="templates")

app = Starlette(debug=True)
app.add_middleware(AuthenticationMiddleware, backend=BasicAuthBackend())
app.mount("/static", StaticFiles(directory="static"), name="static")


@app.route("/")
async def homepage(request):
    return templates.TemplateResponse("index.html", {"request": request})


@app.route("/me")
async def me(request):
    if not request.user.is_authenticated:
        return RedirectResponse("/")

    return templates.TemplateResponse(
        "me.html", {"request": request, "username": request.user.username}
    )


@app.route("/login", methods=["POST", "GET"])
async def login(request):
    if request.user.is_authenticated:
        return RedirectResponse("/me", status_code=303)

    return Response(status_code=401, headers={"WWW-Authenticate": 'Basic realm="Site"'})


@app.route("/new")
async def new_story(request):
    if not request.user.is_authenticated:
        return RedirectResponse("/")

    random_title_suggestion, random_intro_suggestion = random.choice(STORY_STARTERS)
    return templates.TemplateResponse(
        "newstory.html",
        {
            "request": request,
            "random_title_suggestion": random_title_suggestion,
            "random_intro_suggestion": random_intro_suggestion,
        },
    )


@app.route("/new", methods=["POST"])
async def new_story_submit(request):
    if not request.user.is_authenticated:
        return RedirectResponse("/", status_code=303)

    form = await request.form()
    logger.info("received new story submission: %s", form)

    return Response(status_code=200)
