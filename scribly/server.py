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

logger = logging.getLogger(__name__)


USERS = ("zach.the.hammer@gmail.com:password", "gsnussbaum@gmail.com:password")

STORY_STARTERS = [
    (
        "iris in the evening",
        "Years of gardening had taken their toll. Without legs to walk on, with his family gone, all he had was his iris, and how it glowed as the evening sky came out, below the moonlight and the stars.",
    ),
    (
        "alice strongbow",
        "Alice strongbow knew what was to come as she laid her fathers to rest, hand in hand, as they drifted out to sea. She moved the tip of her arrow, smelling of gasoline, closer to the flame.",
    ),
    (
        "the man ghost and a very living dog",
        "'Here boy, here!' It seemed that the stories were true; dogs really *could* see ghosts. His pup turned around and ran right at him, and through him, and he was happy to know that at least for now, he had a friend.",
    ),
    (
        "will you walk into my wavetrap?",
        "Will you walk into my wavetrap? said the spiter to the shy.\n\nIf we each could always do all we ever did.",
    ),
    (
        "alas, the island shrinks!",
        "A manuscript found in the wreckage of an ancient ship tells the strange, and seemingly unbelievable, story of a small brigade of pirates cursed to only know the sea. It begins as such:",
    ),
    (
        "chemicals are making wood invisible",
        "What a strange sight to see the redwood forest, all clear, with light from the sun bending and warping as it passed through the leaves, spraying prismatic colors onto the brush and the water. Things on earth had certainly changed.",
    ),
    (
        "obama's baby bug",
        "As nights grew colder, and days grew shorter, at 1600 Penn -- as his daughters were off to school and his wife touring the country in promotion of better food -- Obama only took solace in visiting the small blue bug he had found in the back corner of an old, unused closet.",
    ),
    (
        "all the windows are fogging in new york",
        "It's december\nIt's getting colder\nAll the windows are fogging in new york.\n\nThe neighbors windows\nLook like icy boulders\nAnd yet I know inside it's awfully warm.",
    ),
]


def random_story_starter() -> Tuple[str, str]:
    return random.choice(STORY_STARTERS)


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

    random_title_suggestion, random_intro_suggestion = random_story_starter()
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
