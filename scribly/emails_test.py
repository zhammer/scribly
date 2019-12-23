import re

from scribly.definitions import User
from scribly.emails import build_email_verification_email


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
