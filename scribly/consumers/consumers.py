import asyncio
import json
import logging

import aio_pika
import aiohttp
import asyncpg

from scribly.consumers.constants import (
    ANNOUNCE_COWRITERS_ADDED_EXCHANGE,
    ANNOUNCE_USER_CREATED_EXCHANGE,
    ANNOUNCE_TURN_TAKEN_EXCHANGE,
)
from scribly.definitions import Context, User
from scribly.database import Database
from scribly.message_gateway import MessageGateway
from scribly.sendgrid import SendGrid
from scribly import env
from scribly.use_scribly import Scribly

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class SendVerificationEmailConsumer:
    def __init__(self, scribly: Scribly) -> None:
        self.scribly = scribly

    async def consume(self, message: aio_pika.IncomingMessage) -> None:
        async with message.process():
            logger.info("Consuming message %s", message.body)
            user = User(**json.loads(message.body.decode()))
            await self.scribly.send_verification_email(user)


class SendTurnNotificationEmailsConsumer:
    def __init__(self, scribly: Scribly) -> None:
        self.scribly = scribly

    async def consume(self, message: aio_pika.IncomingMessage) -> None:
        async with message.process():
            logger.info("Consuming message %s", message.body)
            body = json.loads(message.body.decode())
            await self.scribly.send_turn_email_notifications(
                body["story_id"], body["turn_number"]
            )


class SendAddedToStoryEmailsConsumer:
    def __init__(self, scribly: Scribly) -> None:
        self.scribly = scribly

    async def consume(self, message: aio_pika.IncomingMessage) -> None:
        async with message.process():
            logger.info("Consuming message %s", message.body)
            body = json.loads(message.body.decode())
            await self.scribly.send_added_to_story_emails(body["story_id"])


async def main():
    logger.info("Starting consumers")

    logger.info("Creating an instance of Scribly")
    # make scribly
    db_connection_kwargs = {}
    if "pass@db/scribly" in env.DATABASE_URL:
        # for cypress testing
        db_connection_kwargs["statement_cache_size"] = 0

    rabbit_connection = await aio_pika.connect_robust(env.CLOUDAMQP_URL)
    db_connection = await asyncpg.connect(dsn=env.DATABASE_URL, **db_connection_kwargs)
    sendgrid_session = aiohttp.ClientSession()
    channel = await rabbit_connection.channel()

    message_gateway = MessageGateway(channel)
    database = Database(db_connection)
    emailer = SendGrid(env.SENDGRID_API_KEY, env.SENDGRID_BASE_URL, sendgrid_session)
    scribly = Scribly(Context(database, emailer, message_gateway))

    logger.info("Setting up exchanges")
    announce_user_created_exchange = await channel.declare_exchange(
        ANNOUNCE_USER_CREATED_EXCHANGE, aio_pika.ExchangeType.FANOUT
    )
    announce_turn_taken_exchange = await channel.declare_exchange(
        ANNOUNCE_TURN_TAKEN_EXCHANGE, aio_pika.ExchangeType.FANOUT
    )
    announce_cowriters_added_exchange = await channel.declare_exchange(
        ANNOUNCE_COWRITERS_ADDED_EXCHANGE, aio_pika.ExchangeType.FANOUT
    )

    logger.info("Setting up queues")
    email_verification_queue = await channel.declare_queue("email-verification-queue")
    await email_verification_queue.bind(announce_user_created_exchange)
    await email_verification_queue.consume(
        SendVerificationEmailConsumer(scribly).consume
    )

    turn_notification_email_queue = await channel.declare_queue(
        "turn-notification-email-queue"
    )
    await turn_notification_email_queue.bind(announce_turn_taken_exchange)
    await turn_notification_email_queue.consume(
        SendTurnNotificationEmailsConsumer(scribly).consume
    )

    added_to_story_notification_email_queue = await channel.declare_queue(
        "added-to-story-notification-email-queue"
    )
    await added_to_story_notification_email_queue.bind(
        announce_cowriters_added_exchange
    )
    await added_to_story_notification_email_queue.consume(
        SendAddedToStoryEmailsConsumer(scribly).consume
    )


if __name__ == "__main__":
    loop = asyncio.get_event_loop()
    loop.create_task(main())
    loop.run_forever()

