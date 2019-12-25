import json
import logging

import aio_pika

from scribly.consumers.constants import (
    ANNOUNCE_COWRITERS_ADDED_EXCHANGE,
    ANNOUNCE_TURN_TAKEN_EXCHANGE,
    ANNOUNCE_USER_CREATED_EXCHANGE,
)
from scribly.definitions import MessageGateway as MessageGatewayABC
from scribly.definitions import Story, User

logger = logging.getLogger(__name__)


class MessageGateway(MessageGatewayABC):
    def __init__(self, channel: aio_pika.RobustChannel):
        self.channel = channel

    async def announce_user_created(self, user: User) -> None:
        exchange = await self.channel.declare_exchange(
            ANNOUNCE_USER_CREATED_EXCHANGE, aio_pika.ExchangeType.FANOUT
        )
        body = user.__dict__
        logger.info("Sending message %s to exchange %s", body, exchange.name)
        await exchange.publish(aio_pika.Message(json.dumps(body).encode()), "")

    async def announce_turn_taken(self, story: Story) -> None:
        exchange = await self.channel.declare_exchange(
            ANNOUNCE_TURN_TAKEN_EXCHANGE, aio_pika.ExchangeType.FANOUT
        )

        body = {"story_id": story.id, "turn_number": len(story.turns)}
        logger.info("Sending message %s to exchange %s", body, exchange.name)
        await exchange.publish(aio_pika.Message(json.dumps(body).encode()), "")

    async def announce_cowriters_added(self, story: Story) -> None:
        exchange = await self.channel.declare_exchange(
            ANNOUNCE_COWRITERS_ADDED_EXCHANGE, aio_pika.ExchangeType.FANOUT
        )

        body = {"story_id": story.id}
        logger.info("Sending message %s to exchange %s", body, exchange.name)
        await exchange.publish(aio_pika.Message(json.dumps(body).encode()), "")
