import random
from typing import Sequence, TypeVar

T = TypeVar("T")


def shuffle(original: Sequence[T]) -> Sequence[T]:
    """
    Shuffle a sequence, not in place.
    """
    return random.sample(original, len(original))


def read_once(filename: str) -> str:
    """
    Read a file once, synchronously, and return its contents.
    """
    with open(filename) as f:
        return f.read()
