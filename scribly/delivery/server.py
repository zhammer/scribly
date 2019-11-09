import logging
import os
import random
from typing import Tuple

import asyncpg
from starlette.applications import Starlette
from starlette.middleware.authentication import AuthenticationMiddleware
from starlette.responses import RedirectResponse, Response
from starlette.staticfiles import StaticFiles
from starlette.templating import Jinja2Templates

from scribly.definitions import Context
from scribly.delivery.constants import STORY_STARTERS
from scribly.delivery.middleware import BasicAuthBackend, ScriblyMiddleware

DATABASE_URL = os.environ["DATABASE_URL"]

logger = logging.getLogger(__name__)


templates = Jinja2Templates(directory="templates")

app = Starlette(debug=True)
app.add_middleware(AuthenticationMiddleware, backend=BasicAuthBackend())
app.add_middleware(ScriblyMiddleware)
app.mount("/static", StaticFiles(directory="static"), name="static")


@app.on_event("startup")
async def startup():
    app.state.connection_pool = await asyncpg.create_pool(dsn=DATABASE_URL)


@app.on_event("shutdown")
async def shutdown():
    await app.state.connection_pool.close()


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