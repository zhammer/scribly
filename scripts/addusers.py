import asyncio
import os
import sys
from typing import List, Tuple

import asyncpg

from scribly.definitions import Context
from scribly.use_scribly import Scribly
from scribly.database import Database


async def main():
    """
    Add users like: python addusers.py zach:zachspass:zach@mail.com gabe:gabespass:gabe@mail.com
    """
    users: List[Tuple[str, str, str]] = []
    for user_string in sys.argv[1:]:
        user, password, email = user_string.split(":")
        users.append((user, password, email))
    if not users:
        sys.exit("no users to add")

    connection = await asyncpg.connect(os.environ["DATABASE_URL"])
    db = Database(connection)
    scribly = Scribly(Context(db))

    for username, password, email in users:
        await scribly.sign_up(username, password, email)

    await connection.close()


if __name__ == "__main__":
    asyncio.run(main())
