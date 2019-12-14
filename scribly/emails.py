import os

from jinja2 import Environment, FileSystemLoader
from premailer import Premailer

from scribly.definitions import Email, User
from scribly.util import read_once

website_url = os.environ.get("WEBSITE_URL", "http://127.0.0.1:8000")
premailer = Premailer(css_text=read_once("static/style.css"))
jinja_env = Environment(loader=FileSystemLoader("email_templates"))


def build_email_verification_email(user: User, token: str) -> Email:
    verification_link = f"{website_url}/email-verification?token={token}"
    body = _render_template_with_css(
        "verification.html", verification_link=verification_link, user=user
    )
    return Email(subject="Verify your email", body=body, to=user.email)


def _render_template_with_css(template_name: str, **template_kwargs):
    """
    Render an email template and inline styles from stylesheet.
    """
    template = jinja_env.get_template(template_name)
    rendered_html = template.render(**template_kwargs)
    return premailer.transform(rendered_html)
