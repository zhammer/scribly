import abc
from dataclasses import dataclass


@dataclass
class User:
    id: int
    username: str


class DatabaseGateway(abc.ABC):
    @abc.abstractmethod
    async def fetch_user(self, username: str, password: str) -> User:
        ...


@dataclass
class Context:
    database: DatabaseGateway
