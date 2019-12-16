import os
import time
from typing import Tuple

import argon2
import itsdangerous

from scribly.definitions import EmailVerificationTokenPayload, User
from scribly import env

email_verification_secret = env.EMAIL_VERIFICATION_SECRET
email_verification_serializer = itsdangerous.Serializer(email_verification_secret)
password_hasher = argon2.PasswordHasher()


def hash_password(password: str) -> str:
    """
    Hash a password using argon2.
    """
    return password_hasher.hash(password)


def verify_password_hash(password_hash: str, password: str) -> bool:
    """
    Return True if `password_hash` is valid for `password`, otherwise False.
    """
    try:
        password_hasher.verify(password_hash, password)
    except argon2.exceptions.VerificationError:
        return False
    return True


def check_needs_rehash(password_hash: str) -> bool:
    """
    Return True if `password_hash` is due for a rehash, as per argon2 spec.
    """
    return password_hasher.check_needs_rehash(password_hash)


def build_email_verification_token(user: User) -> str:
    serialized = email_verification_serializer.dumps(
        {"user_id": user.id, "email": user.email, "timestamp": time.time()}
    )
    return itsdangerous.base64_encode(serialized).decode()


def parse_email_verification_token(
    serialized_token: str,
) -> EmailVerificationTokenPayload:
    token = itsdangerous.base64_decode(serialized_token)
    token_content = email_verification_serializer.loads(token)
    return EmailVerificationTokenPayload(**token_content)
