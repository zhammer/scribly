"""One place for all of our env vars."""
from functools import lru_cache
import logging
import os

logger = logging.getLogger(__name__)

_DEFAULTS = {
    "SENDGRID_API_KEY": "test_sendgrid_api_key",
    "SENDGRID_BASE_URL": "https://api.sendgrid.com",
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
            "environment variable %s requested but not available, falling back to default.",
            name,
        )
        return _DEFAULTS[name]
