import json

import aio_pika

from scribly.consumers.constants import (
    ANNOUNCE_USER_CREATED_EXCHANGE,
    ANNOUNCE_TURN_TAKEN_EXCHANGE,
)
from scribly.definitions import MessageGateway as MessageGatewayABC
from scribly.definitions import Story, User


class MessageGateway(MessageGatewayABC):
    def __init__(self, channel: aio_pika.RobustChannel):
        self.channel = channel

    async def announce_user_created(self, user: User) -> None:
        exchange = await self.channel.declare_exchange(
            ANNOUNCE_USER_CREATED_EXCHANGE, aio_pika.ExchangeType.FANOUT
        )
        await exchange.publish(aio_pika.Message(json.dumps(user.__dict__).encode()), "")

    async def announce_turn_taken(self, story: Story) -> None:
        exchange = await self.channel.declare_exchange(
            ANNOUNCE_TURN_TAKEN_EXCHANGE, aio_pika.ExchangeType.FANOUT
        )

        body = {"story_id": story.id, "turn_number": len(story.turns)}
        await exchange.publish(aio_pika.Message(json.dumps(body).encode()), "")

