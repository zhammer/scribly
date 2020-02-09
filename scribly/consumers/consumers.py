import asyncio
import json
import logging
from typing import List

import aio_pika
import aiohttp
import aiosqlite
from typing_extensions import Protocol, Type

from scribly import env
from scribly.consumers.constants import (
    ANNOUNCE_COWRITERS_ADDED_EXCHANGE,
    ANNOUNCE_TURN_TAKEN_EXCHANGE,
    ANNOUNCE_USER_CREATED_EXCHANGE,
)
from scribly.database import Database
from scribly.definitions import User
from scribly.rabbit import Rabbit
from scribly.sendgrid import SendGrid
from scribly.use_scribly import Scribly

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


EXCHANGES = [
    ANNOUNCE_USER_CREATED_EXCHANGE,
    ANNOUNCE_TURN_TAKEN_EXCHANGE,
    ANNOUNCE_COWRITERS_ADDED_EXCHANGE,
]


class ScriblyConsumer(Protocol):
    QUEUE_NAME: str
    BOUND_TO: str

    def __init__(self, scribly: Scribly) -> None:
        ...

    async def consume(self, message: aio_pika.IncomingMessage) -> None:
        ...


class SendVerificationEmailConsumer(ScriblyConsumer):
    QUEUE_NAME = "email-verification-queue"
    BOUND_TO = ANNOUNCE_USER_CREATED_EXCHANGE

    def __init__(self, scribly: Scribly) -> None:
        self.scribly = scribly

    async def consume(self, message: aio_pika.IncomingMessage) -> None:
        async with message.process():
            logger.info("Consuming message %s", message.body)
            user = User(**json.loads(message.body.decode()))
            await self.scribly.send_verification_email(user)


class SendTurnNotificationEmailsConsumer(ScriblyConsumer):
    QUEUE_NAME = "turn-notification-email-queue"
    BOUND_TO = ANNOUNCE_TURN_TAKEN_EXCHANGE

    def __init__(self, scribly: Scribly) -> None:
        self.scribly = scribly

    async def consume(self, message: aio_pika.IncomingMessage) -> None:
        async with message.process():
            logger.info("Consuming message %s", message.body)
            body = json.loads(message.body.decode())
            await self.scribly.send_turn_email_notifications(
                body["story_id"], body["turn_number"]
            )


class SendAddedToStoryEmailsConsumer(ScriblyConsumer):
    QUEUE_NAME = "added-to-story-notification-email-queue"
    BOUND_TO = ANNOUNCE_COWRITERS_ADDED_EXCHANGE

    def __init__(self, scribly: Scribly) -> None:
        self.scribly = scribly

    async def consume(self, message: aio_pika.IncomingMessage) -> None:
        async with message.process():
            logger.info("Consuming message %s", message.body)
            body = json.loads(message.body.decode())
            await self.scribly.send_added_to_story_emails(body["story_id"])


CONSUMERS: List[Type[ScriblyConsumer]] = [
    SendVerificationEmailConsumer,
    SendTurnNotificationEmailsConsumer,
    SendAddedToStoryEmailsConsumer,
]


async def main():
    logger.info("Starting consumers")

    logger.info("Creating an instance of Scribly")
    # make scribly

    rabbit_connection = await aio_pika.connect_robust(env.CLOUDAMQP_URL)
    db_connection = await aiosqlite.connect(env.DATABASE_URL, isolation_level=None)
    db_connection.row_factory = aiosqlite.Row
    sendgrid_session = aiohttp.ClientSession()
    channel = await rabbit_connection.channel()

    message_gateway = Rabbit(channel)
    database = Database(db_connection)
    emailer = SendGrid(env.SENDGRID_API_KEY, env.SENDGRID_BASE_URL, sendgrid_session)
    scribly = Scribly(database, emailer, message_gateway)

    logger.info("Setting up exchanges")
    exchange_by_name = {
        exchange_name: await channel.declare_exchange(
            exchange_name, aio_pika.ExchangeType.FANOUT
        )
        for exchange_name in EXCHANGES
    }

    logger.info("Setting up consumers")
    for consumer in CONSUMERS:
        exchange = exchange_by_name[consumer.BOUND_TO]
        queue = await channel.declare_queue(consumer.QUEUE_NAME)
        await queue.bind(exchange)
        await queue.consume(consumer(scribly).consume)


if __name__ == "__main__":
    loop = asyncio.get_event_loop()
    loop.create_task(main())
    loop.run_forever()
