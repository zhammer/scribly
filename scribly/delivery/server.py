import inspect
import logging
import traceback
from contextlib import asynccontextmanager
from inspect import FrameInfo
from typing import AsyncGenerator, Dict, Optional
from urllib.parse import urlparse

import aio_pika
import aiohttp
import asyncpg
from starlette.applications import Starlette
from starlette.middleware.authentication import AuthenticationMiddleware
from starlette.middleware.sessions import SessionMiddleware
from starlette.responses import HTMLResponse, RedirectResponse, Response
from starlette.routing import Route
from starlette.staticfiles import StaticFiles
from starlette.templating import Jinja2Templates
from user_agents import parse

from scribly import env, exceptions
from scribly.database import Database
from scribly.definitions import User
from scribly.jinja_helpers import RemoveNewlines
from scribly.rabbit import Rabbit
from scribly.sendgrid import SendGrid
from scribly.use_scribly import Scribly

DATABASE_URL = env.DATABASE_URL
SESSION_SECRET_KEY = env.SESSION_SECRET_KEY

logger = logging.getLogger(__name__)


templates = Jinja2Templates(directory="templates")
templates.env.add_extension(RemoveNewlines)


async def startup():
    connection_kwargs = {}
    if "pass@db/scribly" in DATABASE_URL:
        # for cypress testing
        connection_kwargs["statement_cache_size"] = 0
    app.state.connection_pool = await asyncpg.create_pool(
        dsn=DATABASE_URL, min_size=2, max_size=2, **connection_kwargs
    )

    app.state.rabbit_connection = await aio_pika.connect_robust(env.CLOUDAMQP_URL)


async def shutdown():
    await app.state.connection_pool.close()
    await app.state.rabbit_connection.close()


@asynccontextmanager
async def get_scribly(app: Starlette) -> AsyncGenerator[Scribly, None]:
    rabbit_channel = await app.state.rabbit_connection.channel()
    async with app.state.connection_pool.acquire() as db_connection, aiohttp.ClientSession() as sendgrid_session:
        database = Database(db_connection)
        emailer = SendGrid(
            env.SENDGRID_API_KEY, env.SENDGRID_BASE_URL, sendgrid_session
        )
        message_gateway = Rabbit(rabbit_channel)
        yield Scribly(database, emailer, message_gateway)

    await rabbit_channel.close()


def get_session_user(request) -> Optional[User]:
    session_user = request.session.get("user", None)
    if not session_user:
        return None

    try:
        return User(**session_user)
    except Exception as e:
        print(f"Error plucking user from session: {e}")
        return None


def set_session_user(request, user: User) -> None:
    request.session["user"] = user.__dict__


def clear_session_user(request) -> None:
    del request.session["user"]


async def homepage(request):
    user = get_session_user(request)
    if user:
        return RedirectResponse("/me")

    return templates.TemplateResponse("index.html", {"request": request})


async def me(request):
    user = get_session_user(request)
    if not user:
        return RedirectResponse("/")

    async with get_scribly(request.app) as scribly:
        me = await scribly.get_me(user)

    set_session_user(request, me.user)

    user_agent = parse(request.headers["user-agent"])
    return templates.TemplateResponse(
        "me.html", {"request": request, "me": me, "mobile": user_agent.is_mobile}
    )


async def log_in_page(request):
    user = get_session_user(request)
    if user:
        return RedirectResponse("/me")

    return templates.TemplateResponse("login.html", {"request": request})


async def login(request):
    form = await request.form()

    async with get_scribly(request.app) as scribly:
        user = await scribly.log_in(form["username"], form["password"])

    set_session_user(request, user)
    return RedirectResponse("/me", status_code=303)


async def sign_up(request):
    form = await request.form()
    username = form["username"]
    password = form["password"]
    password_confirmation = form["password_confirmation"]
    email = form["email"]

    if not password == password_confirmation:
        raise exceptions.InputError("Passwords do not match!")

    async with get_scribly(request.app) as scribly:
        user = await scribly.sign_up(username, password, email)

    set_session_user(request, user)

    return RedirectResponse(f"/me", status_code=303)


async def logout(request):
    clear_session_user(request)
    return RedirectResponse("/", status_code=303)


async def sign_up_page(request):
    user = get_session_user(request)
    if user:
        return RedirectResponse("/me")

    return templates.TemplateResponse("signup.html", {"request": request})


async def new_story(request):
    user = get_session_user(request)
    if not user:
        return RedirectResponse("/")

    return templates.TemplateResponse("newstory.html", {"request": request})


async def new_story_submit(request):
    user = get_session_user(request)
    if not user:
        return RedirectResponse("/", status_code=303)

    form = await request.form()
    logger.info("received new story submission: %s", form)

    async with get_scribly(request.app) as scribly:
        story = await scribly.start_story(user, form["title"], form["body"])

    return RedirectResponse(f"/stories/{story.id}", status_code=303)


async def add_cowriters(request):
    user = get_session_user(request)
    if not user:
        return RedirectResponse("/", status_code=303)

    story_id = int(request.path_params["story_id"])
    logger.info("request to add cowriters to story %s", story_id)

    form = await request.form()
    cowriter_usernames = [form["person-1"]]
    if form["person-2"]:
        cowriter_usernames.append(form["person-2"])
    if form["person-3"]:
        cowriter_usernames.append(form["person-3"])

    async with get_scribly(request.app) as scribly:
        await scribly.add_cowriters(user, story_id, cowriter_usernames)

    return RedirectResponse(f"/stories/{story_id}", status_code=303)


async def request_email_verification_email(request):
    user = get_session_user(request)
    if not user:
        return RedirectResponse("/", status_code=303)

    async with get_scribly(request.app) as scribly:
        await scribly.send_verification_email(user)

    return templates.TemplateResponse(
        "emailverificationrequested.html", {"request": request, "user": user}
    )


async def verify_email_link(request):
    token = request.query_params["token"]

    async with get_scribly(request.app) as scribly:
        email = await scribly.verify_email(token)
    return templates.TemplateResponse(
        "emailverificationsuccess.html", {"request": request, "email": email}
    )


async def submit_turn(request):
    user = get_session_user(request)
    if not user:
        return RedirectResponse("/", status_code=303)

    story_id = int(request.path_params["story_id"])

    form = await request.form()
    action = form["action"]
    if not action in {"write", "pass", "finish", "write_and_finish"}:
        raise RuntimeError(
            f"Unknown turn action {action} from user {user.id} for story {story_id}."
        )

    async with get_scribly(request.app) as scribly:
        if action == "write":
            await scribly.take_turn_write(user, story_id, form["text"])
        if action == "pass":
            await scribly.take_turn_pass(user, story_id)
        if action == "finish":
            await scribly.take_turn_finish(user, story_id)
        if action == "write_and_finish":
            await scribly.take_turn_write_and_finish(user, story_id, form["text"])

    return RedirectResponse(f"/stories/{story_id}", status_code=303)


async def story_page(request):
    user = get_session_user(request)
    if not user:
        return RedirectResponse("/", status_code=303)

    story_id = int(request.path_params["story_id"])

    async with get_scribly(request.app) as scribly:
        story = await scribly.get_story(user, story_id)

    if story.state == "draft":
        return templates.TemplateResponse(
            "addpeopletostory.html", {"request": request, "story": story},
        )

    if story.state in {"in_progress", "done"}:
        return templates.TemplateResponse(
            "story.html", {"request": request, "user": user, "story": story,},
        )


async def nudge(request):
    user = get_session_user(request)
    if not user:
        return RedirectResponse("/", status_code=303)

    story_id = int(request.path_params["story_id"])
    nudgee_id = int(request.path_params["nudgee_id"])

    async with get_scribly(request.app) as scribly:
        await scribly.nudge(user, nudgee_id, story_id)

    return templates.TemplateResponse(
        "nudged.html", {"request": request, "story_id": story_id, "user": user}
    )


async def hide_story(request):
    user = get_session_user(request)
    if not user:
        return RedirectResponse("/", status_code=303)

    story_id = int(request.path_params["story_id"])

    async with get_scribly(request.app) as scribly:
        await scribly.hide_story(user, story_id)

    return RedirectResponse(
        _path_from_referer(request.headers["referer"]), status_code=303
    )


async def unhide_story(request):
    user = get_session_user(request)
    if not user:
        return RedirectResponse("/", status_code=303)

    story_id = int(request.path_params["story_id"])

    async with get_scribly(request.app) as scribly:
        await scribly.unhide_story(user, story_id)

    return RedirectResponse(
        _path_from_referer(request.headers["referer"]), status_code=303
    )


def _path_from_referer(referer: str) -> str:
    """
    >>> _path_from_referer("http://127.0.0.1:8000/me?show_hidden=0")
    '/me?show_hidden=0'

    >>> _path_from_referer("scribly.app/me?show_hidden=0")
    '/me?show_hidden=0'

    >>> _path_from_referer("scribly.app/me")
    '/me'
    """
    parsed = urlparse(referer)
    if parsed.query:
        return f"{parsed.path}?{parsed.query}"
    return parsed.path


async def exception(request):
    raise Exception("Raising an exception, intentionally!")


async def server_error(request, exception: Exception):

    # most of this is copied from the debug starlette error page (https://github.com/encode/starlette/blob/c80558e04d06e6f55831fbe6c38dfcc5393fc56d/starlette/middleware/errors.py#L210-L227)
    # main purpose is to remake that page but with scribly styles
    limit = 7
    traceback_obj = traceback.TracebackException.from_exception(
        exception, capture_locals=True
    )
    frames = inspect.getinnerframes(
        traceback_obj.exc_traceback, limit  # type: ignore
    )
    center_line_number = int((limit - 1) / 2)
    return templates.TemplateResponse(
        "exception.html",
        {
            "request": request,
            "frame_infos": [
                _build_frame_info(frame, center_line_number) for frame in frames
            ],
            "traceback": traceback_obj,
        },
        status_code=500,
    )


def _build_frame_info(frame: FrameInfo, center_line_number: int) -> Dict:
    code_lines = [
        {
            "line": line,
            "line_number": frame.lineno + (position - center_line_number),
            "center": position == center_line_number,
        }
        for position, line in enumerate(frame.code_context or [])
    ]

    return {"frame": frame, "code_lines": code_lines}


app = Starlette(
    on_startup=[startup],
    on_shutdown=[shutdown],
    routes=[
        Route("/", homepage),
        Route("/me", me),
        Route("/login", log_in_page),
        Route("/login", login, methods=["POST"]),
        Route("/signup", sign_up, methods=["POST"]),
        Route("/logout", logout, methods=["GET", "POST"]),
        Route("/signup", sign_up_page),
        Route("/new", new_story),
        Route("/new", new_story_submit, methods=["POST"]),
        Route("/stories/{story_id}/addcowriters", add_cowriters, methods=["POST"]),
        Route(
            "/email-verification", request_email_verification_email, methods=["POST"]
        ),
        Route("/email-verification", verify_email_link),
        Route("/stories/{story_id}/turn", submit_turn, methods=["POST"]),
        Route("/stories/{story_id}", story_page),
        Route("/stories/{story_id}/nudge/{nudgee_id}", nudge, methods=["POST"]),
        Route("/stories/{story_id}/hide", hide_story, methods=["POST"]),
        Route("/stories/{story_id}/unhide", unhide_story, methods=["POST"]),
        Route("/exception", exception),
    ],
    exception_handlers={500: server_error},
)
app.add_middleware(SessionMiddleware, secret_key=SESSION_SECRET_KEY)
app.mount("/static", StaticFiles(directory="static"), name="static")
