class ScriblyException(Exception):
    """Base exception for Scribly."""


class AuthError(ScriblyException):
    """Raised when encountering an auth error."""
