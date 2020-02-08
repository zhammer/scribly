import os
import sys
from pathlib import Path
import sqlite3


def main():
    with open("./migrations/createdb.sql") as f:
        sql = f.read()

    db_path = Path(os.environ["DATABASE_URL"])
    os.makedirs(db_path.parent, exist_ok=True)

    if "--reset" in sys.argv[1:]:
        os.remove(db_path)

    connection = sqlite3.connect(db_path)

    try:
        connection = connection.executescript(sql)
    except Exception as e:
        print(e)


if __name__ == "__main__":
    main()
