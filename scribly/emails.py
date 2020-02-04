from typing import List

from jinja2 import Environment, FileSystemLoader
from premailer import Premailer

from scribly import env
from scribly.definitions import Email, Story, Turn, User
from scribly.jinja_helpers import RemoveNewlines
from scribly.util import read_once

website_url = env.WEBSITE_URL
premailer = Premailer(
    css_text=read_once("static/style.css"), cssutils_logging_level="CRITICAL"
)
jinja_env = Environment(
    loader=FileSystemLoader("email_templates"), extensions=[RemoveNewlines]
)
jinja_env.globals["website_url"] = website_url


def build_added_to_story_emails(story: Story) -> List[Email]:
    assert story.cowriters

    recipients = [
        user
        for user in story.cowriters
        if user.email_verification_status == "verified"
        and not (user.id == story.created_by.id)
    ]

    return [_build_added_to_story_email(story, recipient) for recipient in recipients]


def _build_added_to_story_email(story: Story, recipient: User) -> Email:
    body = _render_template_with_css(
        "addedtostory.html", story=story, recipient=recipient
    )
    subject = f"{story.created_by.username} started the story {story.title}"
    if story.current_writers_turn.id == recipient.id:
        subject += " - it's your turn!"
    return Email(subject=subject, body=body, to=recipient)


def build_turn_email_notifications(story: Story, turn_number: int) -> List[Email]:
    assert len(story.turns) >= turn_number
    assert story.cowriters

    turn = story.turns[turn_number - 1]
    recipients = [
        user
        for user in story.cowriters
        if (user.email_verification_status == "verified")
        and (not user.id == turn.taken_by.id)
    ]

    return [
        _build_turn_email_notification(story, turn, recipient)
        for recipient in recipients
    ]


def _build_turn_email_notification(story: Story, turn: Turn, recipient: User) -> Email:
    body = _render_template_with_css(
        "storyturnnotification.html", story=story, turn=turn, recipient=recipient
    )
    subject: str
    if turn.action in ("pass", "write"):
        assert story.current_writers_turn
        if story.current_writers_turn.id == recipient.id:
            subject = f"It's your turn on {story.title}!"
        else:
            subject = f"{turn.taken_by.username} took their turn on {story.title}!"
    if turn.action in ("write_and_finish", "finish"):
        subject = f"{story.title} is done!"

    return Email(subject=subject, body=body, to=recipient)


def build_email_verification_email(user: User, token: str) -> Email:
    verification_link = f"{website_url}/email-verification?token={token}"
    body = _render_template_with_css(
        "verification.html", verification_link=verification_link, user=user
    )
    return Email(subject="Verify your email", body=body, to=user)


def build_nudge_email(nudger: User, nudgee: User, story: Story) -> Email:
    subject = f"{nudger.username} nudged you to take your turn on {story.title}"
    body = _render_template_with_css(
        "nudge.html", nudger=nudger, nudgee=nudgee, story=story
    )
    return Email(subject, body, to=nudgee)


def _render_template_with_css(template_name: str, **template_kwargs):
    """
    Render an email template and inline styles from stylesheet.
    """
    template = jinja_env.get_template(template_name)
    rendered_html = template.render(**template_kwargs)
    return premailer.transform(rendered_html)
