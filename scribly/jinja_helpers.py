import re
from typing import Optional

from jinja2.ext import Extension


class RemoveNewlines(Extension):
    def preprocess(self, source: str, name: str, filename: Optional[str] = None) -> str:
        return re.sub(r"(?<!>)\n\s+(?!<|\s)", " ", source)
