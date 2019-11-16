import argon2


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
