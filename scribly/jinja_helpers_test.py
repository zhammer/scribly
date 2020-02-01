from scribly.jinja_helpers import RemoveNewlines


def test_remove_new_lines() -> None:
    source = """
<div>
    <p>
        this is a paragraph with long text. i'm adding a newline,
        but i don't want the newline to appear in the rendered output.
    </p>
</div>
"""

    extension = RemoveNewlines(None)
    output = extension.preprocess(source, name="")

    expected = """
<div>
    <p>this is a paragraph with long text. i'm adding a newline, but i don't want the newline to appear in the rendered output.</p>
</div>
"""
    assert output == expected

