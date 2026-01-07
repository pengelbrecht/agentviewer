#!/usr/bin/env python3
"""User management module."""


class User:
    """Represents a user."""

    def __init__(self, name, email):
        self.name = name
        self.email = email

    def greet(self):
        print(f"Hello, {self.name}!")


def get_users():
    """Return list of users."""
    return [
        User("Alice", "alice@example.com"),
        User("Bob", "bob@example.com"),
    ]


def main():
    users = get_users()
    for user in users:
        user.greet()


if __name__ == "__main__":
    main()
