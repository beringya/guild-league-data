from app.core.security import hash_password, verify_password


def test_password_hash_roundtrip():
    hashed = hash_password("a-strong-password")
    assert verify_password("a-strong-password", hashed)
    assert not verify_password("wrong-password", hashed)
