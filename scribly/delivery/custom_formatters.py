from typing import Tuple

from scribly.definitions import Turn, TurnText, TurnTextComponent


def format_turn_text(turn: Turn) -> str:
    if not turn.text_written:
        return ""

    if not len(turn.text_written) == 1 and turn.text_written[0].component == "text":
        raise NotImplementedError(
            "formatting a turn that is not just one text block is not yet supported!"
        )

    return turn.text_written[0].text


def pluck_story_text(text: str) -> TurnText:
    return tuple([TurnTextComponent(text, "text")])
