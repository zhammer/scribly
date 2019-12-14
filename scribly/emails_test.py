from scribly.definitions import User
from scribly.emails import build_email_verification_email


def test_verification_email() -> None:
    # given a token we have generated for a user
    user = User(id=1, username="zach", email="zach@email.com")
    token = "verification_token"

    # when we build an email for the user
    email = build_email_verification_email(user, token)

    # then
    assert email.subject == "Verify your email"
    assert email.to == user.email
    expected_body = """<!DOCTYPE html>
<html lang="en">

<head><style type="text/css">button:hover {background:lightgray}
button:focus {background:lightgray}
a.button:hover {background:lightgray}
a.button:focus {background:lightgray}
a.button:hover {color:#191919;text-decoration:none}
a.button:focus {color:#191919;text-decoration:none}
a.button:visited {color:#191919;text-decoration:none}</style></head>
<body style="color:#191919; font-family:monospace; font-size:16px">
  <h1>verify your email</h1>
  <p>zach, click the following link to verify your email</p>
  <a class="button" href="http://127.0.0.1:8000/email-verification?token=verification_token" style="background:transparent; border:solid 1px; cursor:pointer; font-size:11px; padding:1em; color:#191919; text-decoration:none">verify your email</a>
</body>

</html>
"""
    assert email.body == expected_body
