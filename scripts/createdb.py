import asyncio
import os
import sys

import asyncpg


async def main():
    with open("./migrations/createdb.sql") as f:
        sql = f.read()

    connection = await asyncpg.connect(os.environ["DATABASE_URL"])

    if "--reset" in sys.argv[1:]:
        await connection.execute(
            """
            DROP SCHEMA IF EXISTS public CASCADE;
            CREATE SCHEMA public;
            """
        )

    await connection.execute(sql)
    await connection.close()


if __name__ == "__main__":
    asyncio.run(main())
