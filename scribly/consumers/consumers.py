import asyncio
import json
import logging

import aio_pika
from aio_pika.pool import Pool

from scribly.consumers.constants import ANNOUNCE_USER_CREATED_EXCHANGE
from scribly.definitions import User
from scribly import env

logger = logging.getLogger(__name__)


async def process_send_verification_email(message: aio_pika.IncomingMessage):
    async with message.process():
        user = User(**json.loads(message.body.decode()))
        print(user)


async def main():
    connection = await aio_pika.connect_robust(env.CLOUDAMQP_URL)
    async with connection:
        channel = await connection.channel()

        # announce user created exchange
        exchange = await channel.declare_exchange(
            ANNOUNCE_USER_CREATED_EXCHANGE, aio_pika.ExchangeType.FANOUT
        )

        email_verification_queue = await channel.declare_queue(
            "email-verification-queue"
        )

        await email_verification_queue.bind(exchange)
        while True:
            await email_verification_queue.consume(process_send_verification_email)


if __name__ == "__main__":
    asyncio.run(main())
