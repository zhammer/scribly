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

    try:
        await connection.execute(sql)
    except asyncpg.exceptions.DuplicateObjectError:
        print("db already exists, skipping migration")

    await connection.close()


if __name__ == "__main__":
    asyncio.run(main())
