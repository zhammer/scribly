import asyncio
import os
import sys
from typing import List, Tuple

import asyncpg


async def main():
    """
    Add users like: python addusers.py zach:zachspass gabe:gabespass
    """
    users: List[Tuple[str, str]] = []
    for user_string in sys.argv[1:]:
        user, _, password = user_string.partition(":")
        users.append((user, password))
    if not users:
        sys.exit("no users to add")

    connection = await asyncpg.connect(os.environ["DATABASE_URL"])

    await connection.executemany(
        """
        INSERT INTO users (username, password)
        VALUES ($1, $2)
        """,
        users,
    )
    await connection.close()


if __name__ == "__main__":
    asyncio.run(main())
