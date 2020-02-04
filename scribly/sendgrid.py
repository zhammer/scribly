import logging

import aiohttp

from scribly.definitions import Email, EmailGateway

logger = logging.getLogger(__name__)


class SendGrid(EmailGateway):
    sender = "emails@scribly.app"

    def __init__(
        self, api_key: str, base_url: str, session: aiohttp.ClientSession
    ) -> None:
        self.session = session
        self.api_key = api_key
        self.base_url = base_url

    async def send_email(self, email: Email) -> None:
        logger.info("Sending email %s", email)
        url = f"{self.base_url}/v3/mail/send"
        body = {
            "personalizations": [
                {
                    "to": [{"email": email.to.email, "name": email.to.username}],
                    "subject": email.subject,
                }
            ],
            "from": {"email": self.sender},
            "content": [{"type": "text/html", "value": email.body}],
        }
        headers = {"Authorization": f"Bearer {self.api_key}"}
        async with self.session.post(url, json=body, headers=headers) as response:
            if response.status < 200 or response.status > 299:
                text = await response.text()
                raise RuntimeError(
                    f"Received non-200 response from sendgrid {response.status} - '{text}'."
                )
