import asyncio
import json
import logging

import aio_pika
import aiohttp
import asyncpg

from scribly.consumers.constants import ANNOUNCE_USER_CREATED_EXCHANGE
from scribly.definitions import Context, User
from scribly.database import Database
from scribly.message_gateway import MessageGateway
from scribly.sendgrid import SendGrid
from scribly import env
from scribly.use_scribly import Scribly

logger = logging.getLogger(__name__)


class SendVerificationEmailConsumer:
    def __init__(self, scribly: Scribly) -> None:
        self.scribly = scribly

    async def consume(self, message: aio_pika.IncomingMessage) -> None:
        async with message.process():
            user = User(**json.loads(message.body.decode()))
            await self.scribly.send_verification_email(user)


async def main():
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

    # announce user created exchange
    exchange = await channel.declare_exchange(
        ANNOUNCE_USER_CREATED_EXCHANGE, aio_pika.ExchangeType.FANOUT
    )

    email_verification_queue = await channel.declare_queue("email-verification-queue")

    await email_verification_queue.bind(exchange)
    await email_verification_queue.consume(
        SendVerificationEmailConsumer(scribly).consume
    )


if __name__ == "__main__":
    loop = asyncio.get_event_loop()
    loop.create_task(main())
    loop.run_forever()

