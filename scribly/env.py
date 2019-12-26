"""One place for all of our env vars."""
from functools import lru_cache
import logging
import os
import typing

logger = logging.getLogger(__name__)

if typing.TYPE_CHECKING:
    CLOUDAMQP_URL: str
    DATABASE_URL: str
    EMAIL_VERIFICATION_SECRET: str
    SENDGRID_API_KEY: str
    SENDGRID_BASE_URL: str
    SESSION_SECRET_KEY: str
    WEBSITE_URL: str

_DEFAULTS = {
    "CLOUDAMQP_URL": "amqp://guest:guest@localhost:5672/%2F",
    "DATABASE_URL": "postgres://scribly:pass@localhost/scribly",
    "EMAIL_VERIFICATION_SECRET": "myemailverificationsecret",
    "SENDGRID_API_KEY": "test_sendgrid_api_key",
    "SENDGRID_BASE_URL": "https://api.sendgrid.com",
    "SESSION_SECRET_KEY": "dev_session_secret",
    "WEBSITE_URL": "http://127.0.0.1:8000",
}


@lru_cache()
def __getattr__(name: str) -> str:
    """
    Get a value from the environment. If the value does not exist in the environment, use
    the default provided in _DEFAULTS. If there is no default, raise a KeyError.
    """
    if name not in _DEFAULTS:
        logger.warning(
            "environment variable %s requested, no default is provided.", name
        )
        return os.environ[name]

    try:
        return os.environ[name]
    except KeyError:
        logger.warning(
            "environment variable '%s' requested but not available, falling back to default: '%s'.",
            name,
            _DEFAULTS[name],
        )
        return _DEFAULTS[name]
