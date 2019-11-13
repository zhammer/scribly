import random
from typing import Sequence, TypeVar

T = TypeVar("T")


def shuffle(original: Sequence[T]) -> Sequence[T]:
    """
    Shuffle a sequence, not in place.
    """
    return random.sample(original, len(original))
