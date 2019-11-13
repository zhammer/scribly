class ScriblyException(Exception):
    """Base exception for Scribly."""


class AuthError(ScriblyException):
    """Raised when encountering an auth error."""


class StoryNotFound(ScriblyException):
    """Raised when attempting to find a story that doesn't exist."""


class InputError(ScriblyException):
    """Raised upon encountering some bad input from the user.."""
