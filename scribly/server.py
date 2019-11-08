from starlette.applications import Starlette
from starlette.staticfiles import StaticFiles
from starlette.templating import Jinja2Templates

templates = Jinja2Templates(directory="templates")

app = Starlette(debug=True)
app.mount("/static", StaticFiles(directory="static"), name="static")


@app.route("/")
async def homepage(request):
    return templates.TemplateResponse("index.html", {"request": request})
