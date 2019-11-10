import asyncio
import os

import asyncpg


async def main():
    with open("./migrations/createdb.sql") as f:
        sql = f.read()

    connection = await asyncpg.connect(os.environ["DATABASE_URL"])
    await connection.execute(sql)
    await connection.close()


if __name__ == "__main__":
    asyncio.run(main())
