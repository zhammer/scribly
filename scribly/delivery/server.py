import asyncio
import logging
import os
from typing import List, Tuple

import aio_pika
import asyncpg
from starlette.applications import Starlette
from starlette.middleware.authentication import AuthenticationMiddleware
from starlette.middleware.sessions import SessionMiddleware
from starlette.responses import HTMLResponse, RedirectResponse, Response
from starlette.staticfiles import StaticFiles
from starlette.templating import Jinja2Templates

from scribly import exceptions
from scribly.definitions import User
from scribly.delivery import custom_formatters
from scribly.delivery.middleware import (
    SessionAuthBackend,
    ScriblyMiddleware,
    WaitForStartupCompleteMiddleware,
)
from scribly import env
from scribly.use_scribly import Scribly

DATABASE_URL = env.DATABASE_URL
SESSION_SECRET_KEY = env.SESSION_SECRET_KEY

logger = logging.getLogger(__name__)


templates = Jinja2Templates(directory="templates")

app = Starlette(debug=True)
startup_complete_event = asyncio.Event()
app.add_middleware(AuthenticationMiddleware, backend=SessionAuthBackend())
app.add_middleware(ScriblyMiddleware)
app.add_middleware(SessionMiddleware, secret_key=SESSION_SECRET_KEY)
app.add_middleware(
    WaitForStartupCompleteMiddleware, startup_complete_event=startup_complete_event
)
app.mount("/static", StaticFiles(directory="static"), name="static")
templates.env.filters["format_turn_text"] = custom_formatters.format_turn_text


@app.on_event("startup")
async def startup():
    connection_kwargs = {}
    if "pass@db/scribly" in DATABASE_URL:
        # for cypress testing
        connection_kwargs["statement_cache_size"] = 0
    app.state.connection_pool = await asyncpg.create_pool(
        dsn=DATABASE_URL, min_size=2, max_size=2, **connection_kwargs
    )

    app.state.rabbit_connection = await aio_pika.connect_robust(env.CLOUDAMQP_URL)

    startup_complete_event.set()


@app.on_event("shutdown")
async def shutdown():
    await app.state.connection_pool.close()
    await app.state.rabbit_connection.close()


@app.route("/")
async def homepage(request):
    if isinstance(request.user, User):
        return RedirectResponse("/me")

    return templates.TemplateResponse("index.html", {"request": request})


@app.route("/me")
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


@app.route("/login")
async def log_in_page(request):
    if isinstance(request.user, User):
        return RedirectResponse("/me")

    return templates.TemplateResponse("login.html", {"request": request})


@app.route("/login", methods=["POST"])
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


@app.route("/signup", methods=["POST"])
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


@app.route("/logout", methods=["GET", "POST"])
async def logout(request):
    request.session.clear()
    return RedirectResponse("/", status_code=303)


@app.route("/signup", methods=["GET"])
async def sign_up_page(request):
    if isinstance(request.user, User):
        return RedirectResponse("/me")

    return templates.TemplateResponse("signup.html", {"request": request})


@app.route("/new")
async def new_story(request):
    if not isinstance(request.user, User):
        return RedirectResponse("/")

    return templates.TemplateResponse("newstory.html", {"request": request})


@app.route("/new", methods=["POST"])
async def new_story_submit(request):
    if not isinstance(request.user, User):
        return RedirectResponse("/", status_code=303)

    form = await request.form()
    logger.info("received new story submission: %s", form)

    scribly: Scribly = request.scope["scribly"]

    story = await scribly.start_story(
        request.user, form["title"], custom_formatters.pluck_story_text(form["body"])
    )

    return RedirectResponse(f"/stories/{story.id}", status_code=303)


@app.route("/stories/{story_id}/addcowriters", methods=["POST"])
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


@app.route("/email-verification", methods=["POST"])
async def request_email_verification_email(request):
    if not isinstance(request.user, User):
        return RedirectResponse("/", status_code=303)

    scribly: Scribly = request.scope["scribly"]
    await scribly.send_verification_email(request.user)
    return templates.TemplateResponse(
        "emailverificationrequested.html", {"request": request, "user": request.user}
    )


@app.route("/email-verification", methods=["GET"])
async def verify_email_link(request):
    token = request.query_params["token"]
    scribly: Scribly = request.scope["scribly"]
    email = await scribly.verify_email(token)
    return templates.TemplateResponse(
        "emailverificationsuccess.html", {"request": request, "email": email}
    )


@app.route("/stories/{story_id}/turn", methods=["POST"])
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

    turn_text = custom_formatters.pluck_story_text(form["text"])
    if action == "write":
        await scribly.take_turn_write(request.user, story_id, turn_text)
    if action == "pass":
        await scribly.take_turn_pass(request.user, story_id)
    if action == "finish":
        await scribly.take_turn_finish(request.user, story_id)
    if action == "write_and_finish":
        await scribly.take_turn_write_and_finish(request.user, story_id, turn_text)

    return RedirectResponse(f"/stories/{story_id}", status_code=303)


@app.route("/stories/{story_id}")
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
