import aiohttp
from requests import sessions

from scribly.definitions import MessageGateway, Story, User


class GlitchMessageGateway(MessageGateway):
    """
    A lil' hacky message gateway that, instead of writing messages to a message
    broker, posts requests back to the glitch server. This means we only need
    one app running (no webserver + worker), don't need to have a message broker
    running, and don't have to really change the use case code since this just gets
    plugged into the current setup.
    """

    def __init__(self, port: int, session: aiohttp.ClientSession) -> None:
        self.session = session
        self.port = port

    @property
    def base(self) -> str:
        return f"http://127.0.0.1:{self.port}/consumers"

    async def announce_user_created(self, user: User) -> None:
        async with self.session.post(
            f"{self.base}/announce-user-created", json=user.__dict__
        ) as response:
            pass

    async def announce_turn_taken(self, story: Story) -> None:
        async with self.session.post(
            f"{self.base}/announce-turn-taken",
            json={"story_id": story.id, "turn_number": len(story.turns)},
        ) as response:
            pass

    async def announce_cowriters_added(self, story: Story) -> None:
        async with self.session.post(
            f"{self.base}/announce-cowriters-added", json={"story_id": story.id}
        ) as response:
            pass
