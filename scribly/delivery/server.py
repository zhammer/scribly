import asyncio
import inspect
import logging
import os
import traceback
from inspect import FrameInfo
from typing import Dict, List, Tuple

import aio_pika
import asyncpg
from starlette.applications import Starlette
from starlette.middleware.authentication import AuthenticationMiddleware
from starlette.middleware.sessions import SessionMiddleware
from starlette.responses import HTMLResponse, RedirectResponse, Response
from starlette.staticfiles import StaticFiles
from starlette.templating import Jinja2Templates
from starlette.routing import Route

from scribly import env, exceptions
from scribly.definitions import User
from scribly.delivery.middleware import (
    ScriblyMiddleware,
    SessionAuthBackend,
)
from scribly.use_scribly import Scribly

DATABASE_URL = env.DATABASE_URL
SESSION_SECRET_KEY = env.SESSION_SECRET_KEY

logger = logging.getLogger(__name__)


templates = Jinja2Templates(directory="templates")


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


async def homepage(request):
    if isinstance(request.user, User):
        return RedirectResponse("/me")

    return templates.TemplateResponse("index.html", {"request": request})


async def me(request):
    if not isinstance(request.user, User):
        return RedirectResponse("/")

    scribly: Scribly = request.scope["scribly"]
    me = await scribly.get_me(request.user)
    user = me.user

    request.session["user"] = {
        "id": user.id,
        "username": user.username,
        "email": user.email,
        "email_verification_status": user.email_verification_status,
    }

    return templates.TemplateResponse("me.html", {"request": request, "me": me})


async def log_in_page(request):
    if isinstance(request.user, User):
        return RedirectResponse("/me")

    return templates.TemplateResponse("login.html", {"request": request})


async def login(request):
    form = await request.form()
    scribly: Scribly = request.scope["scribly"]

    user = await scribly.log_in(form["username"], form["password"])
    request.session["user"] = {
        "id": user.id,
        "username": user.username,
        "email": user.email,
        "email_verification_status": user.email_verification_status,
    }

    return RedirectResponse("/me", status_code=303)


async def sign_up(request):
    form = await request.form()
    username = form["username"]
    password = form["password"]
    password_confirmation = form["password_confirmation"]
    email = form["email"]

    if not password == password_confirmation:
        raise exceptions.InputError("Passwords do not match!")

    scribly: Scribly = request.scope["scribly"]

    user = await scribly.sign_up(username, password, email)
    request.session["user"] = {
        "id": user.id,
        "username": user.username,
        "email": user.email,
        "email_verification_status": user.email_verification_status,
    }

    return RedirectResponse(f"/me", status_code=303)


async def logout(request):
    request.session.clear()
    return RedirectResponse("/", status_code=303)


async def sign_up_page(request):
    if isinstance(request.user, User):
        return RedirectResponse("/me")

    return templates.TemplateResponse("signup.html", {"request": request})


async def new_story(request):
    if not isinstance(request.user, User):
        return RedirectResponse("/")

    return templates.TemplateResponse("newstory.html", {"request": request})


async def new_story_submit(request):
    if not isinstance(request.user, User):
        return RedirectResponse("/", status_code=303)

    form = await request.form()
    logger.info("received new story submission: %s", form)

    scribly = request.scope["scribly"]

    story = await scribly.start_story(request.user, form["title"], form["body"])

    return RedirectResponse(f"/stories/{story.id}", status_code=303)


async def add_cowriters(request):
    if not isinstance(request.user, User):
        return RedirectResponse("/", status_code=303)

    story_id = int(request.path_params["story_id"])
    logger.info("request to add cowriters to story %s", story_id)

    form = await request.form()
    cowriter_usernames = [form["person-1"]]
    if form["person-2"]:
        cowriter_usernames.append(form["person-2"])
    if form["person-3"]:
        cowriter_usernames.append(form["person-3"])

    scribly: Scribly = request.scope["scribly"]
    await scribly.add_cowriters(request.user, story_id, cowriter_usernames)

    return RedirectResponse(f"/stories/{story_id}", status_code=303)


async def request_email_verification_email(request):
    if not isinstance(request.user, User):
        return RedirectResponse("/", status_code=303)

    scribly: Scribly = request.scope["scribly"]
    await scribly.send_verification_email(request.user)
    return templates.TemplateResponse(
        "emailverificationrequested.html", {"request": request, "user": request.user}
    )


async def verify_email_link(request):
    token = request.query_params["token"]
    scribly: Scribly = request.scope["scribly"]
    email = await scribly.verify_email(token)
    return templates.TemplateResponse(
        "emailverificationsuccess.html", {"request": request, "email": email}
    )


async def submit_turn(request):
    if not isinstance(request.user, User):
        return RedirectResponse("/", status_code=303)

    story_id = int(request.path_params["story_id"])

    form = await request.form()
    action = form["action"]
    if not action in {"write", "pass", "finish", "write_and_finish"}:
        raise RuntimeError(
            f"Unknown turn action {action} from user {request.user.id} for story {story_id}."
        )

    scribly: Scribly = request.scope["scribly"]
    if action == "write":
        await scribly.take_turn_write(request.user, story_id, form["text"])
    if action == "pass":
        await scribly.take_turn_pass(request.user, story_id)
    if action == "finish":
        await scribly.take_turn_finish(request.user, story_id)
    if action == "write_and_finish":
        await scribly.take_turn_write_and_finish(request.user, story_id, form["text"])

    return RedirectResponse(f"/stories/{story_id}", status_code=303)


async def story_page(request):
    if not isinstance(request.user, User):
        return RedirectResponse("/", status_code=303)

    story_id = int(request.path_params["story_id"])

    scribly = request.scope["scribly"]
    story = await scribly.get_story(request.user, story_id)

    if story.state == "draft":
        return templates.TemplateResponse(
            "addpeopletostory.html", {"request": request, "story": story},
        )

    if story.state in {"in_progress", "done"}:
        return templates.TemplateResponse(
            "story.html", {"request": request, "user": request.user, "story": story,},
        )


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
        Route("/exception", exception),
    ],
    exception_handlers={500: server_error},
)
app.add_middleware(AuthenticationMiddleware, backend=SessionAuthBackend())
app.add_middleware(ScriblyMiddleware)
app.add_middleware(SessionMiddleware, secret_key=SESSION_SECRET_KEY)
app.mount("/static", StaticFiles(directory="static"), name="static")
