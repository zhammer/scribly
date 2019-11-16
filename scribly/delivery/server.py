import logging
import os
from typing import Tuple
from typing import List, Tuple

import asyncpg
from starlette.applications import Starlette
from starlette.middleware.authentication import AuthenticationMiddleware
from starlette.responses import HTMLResponse, RedirectResponse, Response
from starlette.staticfiles import StaticFiles
from starlette.templating import Jinja2Templates

from scribly import exceptions
from scribly.definitions import Context, User
from scribly.delivery.middleware import BasicAuthBackend, ScriblyMiddleware
from scribly.use_scribly import Scribly

DATABASE_URL = os.environ["DATABASE_URL"]

logger = logging.getLogger(__name__)


templates = Jinja2Templates(directory="templates")

app = Starlette(debug=True)
app.add_middleware(AuthenticationMiddleware, backend=BasicAuthBackend())
app.add_middleware(ScriblyMiddleware)
app.mount("/static", StaticFiles(directory="static"), name="static")


@app.on_event("startup")
async def startup():
    connection_kwargs = {}
    if "localhost" in DATABASE_URL:
        # for cypress testing
        connection_kwargs["statement_cache_size"] = 0

    app.state.connection_pool = await asyncpg.create_pool(
        dsn=DATABASE_URL, min_size=2, max_size=2, **connection_kwargs
    )


@app.on_event("shutdown")
async def shutdown():
    await app.state.connection_pool.close()


@app.route("/")
async def homepage(request):
    return templates.TemplateResponse("index.html", {"request": request})


@app.route("/me")
async def me(request):
    if not isinstance(request.user, User):
        return RedirectResponse("/")

    scribly: Scribly = request.scope["scribly"]
    me = await scribly.get_me(request.user)

    return templates.TemplateResponse("me.html", {"request": request, "me": me})


@app.route("/login", methods=["POST", "GET"])
async def login(request):
    if isinstance(request.user, User):
        return RedirectResponse("/me", status_code=303)

    return Response(status_code=401, headers={"WWW-Authenticate": 'Basic realm="Site"'})


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

    return RedirectResponse(f"/me", status_code=303)


@app.route("/signup", methods=["GET"])
async def sign_up_page(request):
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

    scribly = request.scope["scribly"]

    story = await scribly.start_story(request.user, form["title"], form["body"])

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
    if action == "write":
        await scribly.take_turn_write(request.user, story_id, form["text"])
    if action == "pass":
        await scribly.take_turn_pass(request.user, story_id)
    if action == "finish":
        await scribly.take_turn_finish(request.user, story_id)
    if action == "write_and_finish":
        await scribly.take_turn_write_and_finish(request.user, story_id, form["text"])

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
