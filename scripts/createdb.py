import asyncio
import os

import asyncpg


async def main():
    connection = await asyncpg.connect(os.environ["DATABASE_URL"])
    await connection.execute(
        """
        CREATE TABLE IF NOT EXISTS users (
            id SERIAL,
            username TEXT NOT NULL UNIQUE,
            password TEXT NOT NULL
        );

        INSERT INTO users (username, password)
        VALUES
            ('zach.the.hammer@gmail.com', 'password'),
            ('gsnussbaum@gmail.com', 'password')
        ;
        """
    )
    await connection.close()


if __name__ == "__main__":
    asyncio.run(main())
