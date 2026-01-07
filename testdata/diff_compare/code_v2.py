#!/usr/bin/env python3
"""User management module with database support."""

from dataclasses import dataclass
from datetime import datetime
from typing import Optional
import sqlite3
import logging

logger = logging.getLogger(__name__)


@dataclass
class User:
    """Represents a user in the system."""

    id: int
    name: str
    email: str
    created_at: datetime
    is_active: bool = True

    def greet(self) -> str:
        """Return a greeting message."""
        return f"Hello, {self.name}!"

    def to_dict(self) -> dict:
        """Convert user to dictionary."""
        return {
            "id": self.id,
            "name": self.name,
            "email": self.email,
            "created_at": self.created_at.isoformat(),
            "is_active": self.is_active,
        }


class UserRepository:
    """Repository for user data operations."""

    def __init__(self, db_path: str):
        self.db_path = db_path
        self._conn: Optional[sqlite3.Connection] = None

    def connect(self) -> None:
        """Establish database connection."""
        self._conn = sqlite3.connect(self.db_path)
        self._conn.row_factory = sqlite3.Row
        logger.info("Connected to database: %s", self.db_path)

    def close(self) -> None:
        """Close database connection."""
        if self._conn:
            self._conn.close()
            self._conn = None

    def get_user(self, user_id: int) -> Optional[User]:
        """Get user by ID."""
        if not self._conn:
            raise RuntimeError("Database not connected")

        cursor = self._conn.execute(
            "SELECT id, name, email, created_at, is_active FROM users WHERE id = ?",
            (user_id,),
        )
        row = cursor.fetchone()

        if row is None:
            return None

        return User(
            id=row["id"],
            name=row["name"],
            email=row["email"],
            created_at=datetime.fromisoformat(row["created_at"]),
            is_active=bool(row["is_active"]),
        )

    def get_all_users(self, active_only: bool = False) -> list[User]:
        """Return list of all users."""
        if not self._conn:
            raise RuntimeError("Database not connected")

        query = "SELECT id, name, email, created_at, is_active FROM users"
        if active_only:
            query += " WHERE is_active = 1"

        cursor = self._conn.execute(query)
        return [
            User(
                id=row["id"],
                name=row["name"],
                email=row["email"],
                created_at=datetime.fromisoformat(row["created_at"]),
                is_active=bool(row["is_active"]),
            )
            for row in cursor.fetchall()
        ]


def main() -> None:
    """Main entry point."""
    logging.basicConfig(level=logging.INFO)

    repo = UserRepository("users.db")
    repo.connect()

    try:
        users = repo.get_all_users(active_only=True)
        for user in users:
            print(user.greet())
            logger.debug("Processed user: %s", user.to_dict())
    finally:
        repo.close()


if __name__ == "__main__":
    main()
