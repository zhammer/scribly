import re
from itertools import cycle
from typing import Iterator, List, Sequence

from faker import Faker

from scribly.definitions import EmailVerificationState, Story, Turn, TurnAction, User
from scribly.emails import (
    build_added_to_story_emails,
    build_email_verification_email,
    build_turn_email_notifications,
)

fake = Faker()


def test_verification_email() -> None:
    # given a token we have generated for a user
    user = User(
        id=1,
        username="zach",
        email="zach@email.com",
        email_verification_status="pending",
    )
    token = "verification_token"

    # when we build an email for the user
    email = build_email_verification_email(user, token)

    # then
    assert email.subject == "Verify your email"
    assert email.to == user.email
    expected_body_regex = re.compile(
        (
            r"<!DOCTYPE html>\s*<html lang=\"en\">\s*"
            r"<head><style.*>.*</style></head>\s*"
            r"<body.*>\s*"
            r"<h1>verify your email</h1>\s*"
            r"<p>zach, click the following link to verify your email</p>\s*"
            r"<a class=\"button\".*"
            r"href=\"http://127\.0\.0\.1:8000/email-verification\?token=verification_token\".*>"
            r"verify your email</a>\s*"
            r"</body>\s*"
            r"</html>$"
        ),
        re.DOTALL,
    )
    assert expected_body_regex.match(email.body)


class TestTurnNotificationEmail:
    def test_sends_notifications_write(self) -> None:
        # given
        zach = make_user("zach", "verified")
        gabe = make_user("gabe", "verified")
        rakesh = make_user("rakesh", "verified")
        cowriters = [zach, gabe, rakesh]

        story = Story(
            id=1,
            title="A god walks into a bar",
            state="in_progress",
            created_by=zach,
            cowriters=cowriters,
            turns=list(make_turns(["write", "write"], cowriters)),
        )

        # when
        emails = build_turn_email_notifications(story, turn_number=2)

        # then
        assert len(emails) == 2

        zach_email = emails[0]
        rakesh_email = emails[1]

        # it's not zach's turn, he sees that gabe took a turn
        assert zach_email.to == zach.email
        assert (
            zach_email.subject == f"{gabe.username} took their turn on {story.title}!"
        )
        expected_zach_email_regex = re.compile(
            (
                r"<!DOCTYPE html>\s*<html lang=\"en\">\s*"
                r"<head><style.*>.*</style></head>\s*"
                r"<body.*>\s*"
                r"<p>\s*<a href=\"http://127.0.0.1:8000/stories/1\">go to your story</a>\s*</p>\s*"
                r"<p>\s*"
                r"<span>\s*gabe wrote a section!\s*</span>\s*"
                r"<span> it's rakesh's turn</span>\s*"
                r"</p>\s*"
                r"<hr>\s*<em>"
                + story.turns[-1].text_written.replace("\n", "<br>")
                + r"</em>\s*"
                r"</body>\s*"
                r"</html>$"
            ),
            re.DOTALL,
        )
        assert expected_zach_email_regex.match(zach_email.body)

        # it's rakesh's turn, he sees that it's his turn
        assert rakesh_email.to == rakesh.email
        assert rakesh_email.subject == f"It's your turn on {story.title}!"
        expected_rakesh_email_regex = re.compile(
            (
                r"<!DOCTYPE html>\s*<html lang=\"en\">\s*"
                r"<head><style.*>.*</style></head>\s*"
                r"<body.*>\s*"
                r"<p>\s*<a href=\"http://127.0.0.1:8000/stories/1\">go to your story</a>\s*</p>\s*"
                r"<p>\s*"
                r"<span>\s*gabe wrote a section!\s*</span>\s*"
                r"<span> it's your turn</span>\s*"
                r"</p>\s*"
                r"<hr>\s*<em>"
                + story.turns[-1].text_written.replace("\n", "<br>")
                + r"</em>\s*"
                r"</body>\s*"
                r"</html>$"
            ),
            re.DOTALL,
        )
        assert expected_rakesh_email_regex.match(rakesh_email.body)

    def test_doesnt_send_email_to_non_verified(self) -> None:
        # given
        zach = make_user("zach", "verified")
        gabe = make_user("gabe", "verified")
        rakesh = make_user("rakesh", "pending")
        cowriters = [zach, gabe, rakesh]

        story = Story(
            id=1,
            title="A god walks into a bar",
            state="in_progress",
            created_by=zach,
            cowriters=cowriters,
            turns=list(make_turns(["write", "write"], cowriters)),
        )

        # when
        emails = build_turn_email_notifications(story, turn_number=2)

        # then rakesh doesn't receive an email as his email is not verified
        assert len(emails) == 1
        assert emails[0].to == zach.email

    def test_sends_notifications_write_and_finish(self) -> None:
        zach = make_user("zach", "verified")
        gabe = make_user("gabe", "verified")
        rakesh = make_user("rakesh", "verified")
        cowriters = [zach, gabe, rakesh]

        story = Story(
            id=1,
            title="A god walks into a bar",
            state="done",
            created_by=zach,
            cowriters=cowriters,
            turns=list(make_turns(["write", "write_and_finish"], cowriters)),
        )

        # when
        emails = build_turn_email_notifications(story, turn_number=2)

        # then
        assert len(emails) == 2
        zach_email = emails[0]
        rakesh_email = emails[1]

        assert zach_email.to == zach.email
        assert rakesh_email.to == rakesh.email

        expected_subject = f"{story.title} is done!"
        assert zach_email.subject == expected_subject
        assert rakesh_email.subject == expected_subject

        expected_email_regex = re.compile(
            (
                r"<!DOCTYPE html>\s*<html lang=\"en\">\s*"
                r"<head><style.*>.*</style></head>\s*"
                r"<body.*>\s*"
                r"<p>\s*<a href=\"http://127.0.0.1:8000/stories/1\">go to your story</a>\s*</p>\s*"
                r"<p>\s*"
                r"<span>\s*gabe wrote a section and finished the story!\s*</span>\s*"
                r"</p>\s*"
                r"<hr>\s*<em>"
                + story.turns[-1].text_written.replace("\n", "<br>")
                + r"</em>\s*"
                r"</body>\s*"
                r"</html>$"
            ),
            re.DOTALL,
        )
        assert expected_email_regex.match(zach_email.body)
        assert expected_email_regex.match(rakesh_email.body)

    def test_sends_notifications_finish(self) -> None:
        zach = make_user("zach", "verified")
        gabe = make_user("gabe", "verified")
        rakesh = make_user("rakesh", "verified")
        cowriters = [zach, gabe, rakesh]

        story = Story(
            id=1,
            title="A god walks into a bar",
            state="done",
            created_by=zach,
            cowriters=cowriters,
            turns=list(make_turns(["write", "finish"], cowriters)),
        )

        # when
        emails = build_turn_email_notifications(story, turn_number=2)

        # then
        assert len(emails) == 2
        zach_email = emails[0]
        rakesh_email = emails[1]

        assert zach_email.to == zach.email
        assert rakesh_email.to == rakesh.email

        expected_subject = f"{story.title} is done!"
        assert zach_email.subject == expected_subject
        assert rakesh_email.subject == expected_subject

        expected_email_regex = re.compile(
            (
                r"<!DOCTYPE html>\s*<html lang=\"en\">\s*"
                r"<head><style.*>.*</style></head>\s*"
                r"<body.*>\s*"
                r"<p>\s*<a href=\"http://127.0.0.1:8000/stories/1\">go to your story</a>\s*</p>\s*"
                r"<p>\s*"
                r"<span>\s*gabe finished the story!\s*</span>\s*"
                r"</p>\s*"
                r"</body>\s*"
                r"</html>$"
            ),
            re.DOTALL,
        )
        assert expected_email_regex.match(zach_email.body)
        assert expected_email_regex.match(rakesh_email.body)

    def test_sends_notifications_pass(self) -> None:
        zach = make_user("zach", "verified")
        gabe = make_user("gabe", "verified")
        rakesh = make_user("rakesh", "verified")
        cowriters = [zach, gabe, rakesh]

        story = Story(
            id=1,
            title="A god walks into a bar",
            state="in_progress",
            created_by=zach,
            cowriters=cowriters,
            turns=list(make_turns(["write", "pass"], cowriters)),
        )

        # when
        emails = build_turn_email_notifications(story, turn_number=2)

        # then
        assert len(emails) == 2
        zach_email = emails[0]
        rakesh_email = emails[1]

        assert zach_email.to == zach.email
        assert rakesh_email.to == rakesh.email

        assert rakesh_email.subject == f"It's your turn on {story.title}!"
        assert zach_email.subject == f"gabe took their turn on {story.title}!"

        expected_rakesh_email_regex = re.compile(
            (
                r"<!DOCTYPE html>\s*<html lang=\"en\">\s*"
                r"<head><style.*>.*</style></head>\s*"
                r"<body.*>\s*"
                r"<p>\s*<a href=\"http://127.0.0.1:8000/stories/1\">go to your story</a>\s*</p>\s*"
                r"<p>\s*"
                r"<span>\s*gabe passed!\s*</span>\s*"
                r"<span> it's your turn</span>\s*"
                r"</p>\s*"
                r"</body>\s*"
                r"</html>$"
            ),
            re.DOTALL,
        )
        assert expected_rakesh_email_regex.match(rakesh_email.body)

        expected_zach_email_regex = re.compile(
            (
                r"<!DOCTYPE html>\s*<html lang=\"en\">\s*"
                r"<head><style.*>.*</style></head>\s*"
                r"<body.*>\s*"
                r"<p>\s*<a href=\"http://127.0.0.1:8000/stories/1\">go to your story</a>\s*</p>\s*"
                r"<p>\s*"
                r"<span>\s*gabe passed!\s*</span>\s*"
                r"<span> it's rakesh's turn</span>\s*"
                r"</p>\s*"
                r"</body>\s*"
            ),
            re.DOTALL,
        )
        assert expected_zach_email_regex.match(zach_email.body)


class TestAddedToStoryEmail:
    def test_builds_added_to_story_emails(self) -> None:
        # given
        zach = make_user("zach", "verified")
        gabe = make_user("gabe", "verified")
        rakesh = make_user("rakesh", "verified")
        cowriters = [zach, gabe, rakesh]

        story = Story(
            id=1,
            title="Phaedrus",
            state="in_progress",
            created_by=zach,
            cowriters=cowriters,
            turns=[
                Turn(
                    taken_by=zach,
                    action="write",
                    text_written="We find Phaedrus.\nIn his caveus.",
                )
            ],
        )

        # when
        emails = build_added_to_story_emails(story)

        # then
        assert len(emails) == 2
        gabe_email = emails[0]
        rakesh_email = emails[1]

        assert gabe_email.to == gabe.email
        assert gabe_email.subject == "zach started the story Phaedrus - it's your turn!"

        assert rakesh_email.to == rakesh.email
        assert rakesh_email.subject == "zach started the story Phaedrus"

        expected_gabe_email_regex = re.compile(
            (
                r"<!DOCTYPE html>\s*<html lang=\"en\">\s*"
                r"<head><style.*>.*</style></head>\s*"
                r"<body.*>\s*"
                r"<p>\s*"
                r"<span>zach started the story <a\s*href=\"http://127.0.0.1:8000/stories/1\">Phaedrus</a>!\s*</span>\s*"
                r"<span> it's your turn</span>\s*"
                r"</p>\s*"
                r"<div>\s*"
                r"<h3.*>cowriters:</h3>\s*"
                r"<ul>\s*"
                r"<li>zach</li>\s*"
                r"<li>gabe</li>\s*"
                r"<li>rakesh</li>\s*"
                r"</ul>\s*"
                r"</div>\s*"
                r"<hr>\s*"
                r"<h1.*>Phaedrus</h1>\s*"
                r"<p>We find Phaedrus.<br>In his caveus.</p>\s*"
                r"</body>\s*"
            ),
            re.DOTALL,
        )
        assert expected_gabe_email_regex.match(gabe_email.body)

        expected_rakesh_email_regex = re.compile(
            (
                r"<!DOCTYPE html>\s*<html lang=\"en\">\s*"
                r"<head><style.*>.*</style></head>\s*"
                r"<body.*>\s*"
                r"<p>\s*"
                r"<span>zach started the story <a\s*href=\"http://127.0.0.1:8000/stories/1\">Phaedrus</a>!\s*</span>\s*"
                r"<span> it's gabe's turn</span>\s*"
                r"</p>\s*"
                r"<div>\s*"
                r"<h3.*>cowriters:</h3>\s*"
                r"<ul>\s*"
                r"<li>zach</li>\s*"
                r"<li>gabe</li>\s*"
                r"<li>rakesh</li>\s*"
                r"</ul>\s*"
                r"</div>\s*"
                r"<hr>\s*"
                r"<h1.*>Phaedrus</h1>\s*"
                r"<p>We find Phaedrus.<br>In his caveus.</p>\s*"
                r"</body>\s*"
            ),
            re.DOTALL,
        )
        assert expected_rakesh_email_regex.match(rakesh_email.body)


user_id = 1


def make_user(name: str, email_verification_status: EmailVerificationState) -> User:
    global user_id
    user = User(
        id=user_id,
        username=name,
        email=f"{name}@mail.com",
        email_verification_status=email_verification_status,
    )
    user_id += 1
    return user


def make_turns(actions: List[TurnAction], users: List[User]) -> Iterator[Turn]:
    for action, user in zip(actions, cycle(users)):
        text = fake.text() if action in ("write", "write_and_finish") else None
        yield Turn(taken_by=user, action=action, text_written=text)
