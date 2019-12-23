import json

import aio_pika

from scribly.consumers.constants import ANNOUNCE_USER_CREATED_EXCHANGE
from scribly.definitions import User, MessageGateway as MessageGatewayABC


class MessageGateway(MessageGatewayABC):
    def __init__(self, channel: aio_pika.RobustChannel):
        self.channel = channel

    async def announce_user_created(self, user: User) -> None:
        exchange = await self.channel.declare_exchange(
            ANNOUNCE_USER_CREATED_EXCHANGE, aio_pika.ExchangeType.FANOUT
        )
        await exchange.publish(aio_pika.Message(json.dumps(user.__dict__).encode()), "")
