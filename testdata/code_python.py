#!/usr/bin/env python3
"""
Python Code Test File - Tests syntax highlighting for Python language features
This file demonstrates various Python language constructs for testing code rendering
"""

from __future__ import annotations

import asyncio
import json
import logging
from abc import ABC, abstractmethod
from collections.abc import Callable, Iterator, Sequence
from contextlib import asynccontextmanager, contextmanager
from dataclasses import dataclass, field
from datetime import datetime, timedelta
from enum import Enum, auto
from functools import lru_cache, wraps
from pathlib import Path
from typing import (
    Any,
    ClassVar,
    Final,
    Generic,
    Literal,
    NamedTuple,
    NewType,
    Optional,
    Protocol,
    TypeAlias,
    TypedDict,
    TypeVar,
    Union,
    cast,
    overload,
    runtime_checkable,
)

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Type aliases
UserId = NewType("UserId", str)
JsonDict: TypeAlias = dict[str, Any]
T = TypeVar("T")
K = TypeVar("K")
V = TypeVar("V")


# Enums
class Status(Enum):
    PENDING = auto()
    RUNNING = auto()
    COMPLETE = auto()
    FAILED = auto()

    def __str__(self) -> str:
        return self.name.lower()


class Priority(Enum):
    LOW = 1
    MEDIUM = 2
    HIGH = 3
    CRITICAL = 4


# TypedDict
class UserDict(TypedDict):
    id: str
    name: str
    email: str
    created_at: str


class PartialUserDict(TypedDict, total=False):
    id: str
    name: str
    email: str
    metadata: dict[str, Any]


# NamedTuple
class Point(NamedTuple):
    x: float
    y: float
    label: str = ""

    def distance_from_origin(self) -> float:
        return (self.x**2 + self.y**2) ** 0.5


# Dataclasses
@dataclass
class Config:
    host: str = "localhost"
    port: int = 8080
    debug: bool = False
    tags: list[str] = field(default_factory=list)
    settings: dict[str, str] = field(default_factory=dict)


@dataclass(frozen=True, slots=True)
class ImmutableRecord:
    id: str
    value: int
    timestamp: datetime = field(default_factory=datetime.now)


# Protocol (structural subtyping)
@runtime_checkable
class Comparable(Protocol):
    def __lt__(self, other: Any) -> bool: ...
    def __eq__(self, other: object) -> bool: ...


class Serializable(Protocol):
    def to_json(self) -> str: ...

    @classmethod
    def from_json(cls, data: str) -> Serializable: ...


# Abstract base class
class Entity(ABC):
    """Abstract base class for all entities."""

    created_at: datetime
    updated_at: datetime | None

    @abstractmethod
    def validate(self) -> bool:
        """Validate the entity."""
        ...

    @abstractmethod
    def to_dict(self) -> JsonDict:
        """Convert to dictionary."""
        ...


# Generic class
class Result(Generic[T]):
    """Result type for operations that may fail."""

    __slots__ = ("_value", "_error")

    def __init__(self, value: T | None = None, error: Exception | None = None):
        self._value = value
        self._error = error

    @property
    def is_success(self) -> bool:
        return self._error is None

    @property
    def value(self) -> T:
        if self._error is not None:
            raise self._error
        return cast(T, self._value)

    @classmethod
    def success(cls, value: T) -> Result[T]:
        return cls(value=value)

    @classmethod
    def failure(cls, error: Exception) -> Result[T]:
        return cls(error=error)

    def map(self, fn: Callable[[T], K]) -> Result[K]:
        if self._error is not None:
            return Result.failure(self._error)
        return Result.success(fn(self.value))


# Cache implementation
class LRUCache(Generic[K, V]):
    """Simple LRU cache implementation."""

    MAX_SIZE: ClassVar[int] = 1000
    DEFAULT_TTL: Final[int] = 300

    def __init__(self, max_size: int = 100):
        self._max_size = min(max_size, self.MAX_SIZE)
        self._cache: dict[K, tuple[V, datetime]] = {}

    def get(self, key: K) -> V | None:
        if key not in self._cache:
            return None
        value, timestamp = self._cache[key]
        # Check TTL
        if (datetime.now() - timestamp).seconds > self.DEFAULT_TTL:
            del self._cache[key]
            return None
        return value

    def set(self, key: K, value: V) -> None:
        if len(self._cache) >= self._max_size:
            # Remove oldest entry
            oldest_key = min(self._cache, key=lambda k: self._cache[k][1])
            del self._cache[oldest_key]
        self._cache[key] = (value, datetime.now())

    def __contains__(self, key: K) -> bool:
        return key in self._cache


# Decorators
def retry(
    max_attempts: int = 3,
    delay: float = 1.0,
    exceptions: tuple[type[Exception], ...] = (Exception,),
) -> Callable[[Callable[..., T]], Callable[..., T]]:
    """Decorator to retry a function on failure."""

    def decorator(func: Callable[..., T]) -> Callable[..., T]:
        @wraps(func)
        def wrapper(*args: Any, **kwargs: Any) -> T:
            last_error: Exception | None = None
            for attempt in range(max_attempts):
                try:
                    return func(*args, **kwargs)
                except exceptions as e:
                    last_error = e
                    logger.warning(
                        f"Attempt {attempt + 1}/{max_attempts} failed: {e}"
                    )
                    if attempt < max_attempts - 1:
                        import time
                        time.sleep(delay * (2**attempt))
            raise last_error or RuntimeError("Retry failed")

        return wrapper

    return decorator


# Async decorator
def async_retry(
    max_attempts: int = 3,
    delay: float = 1.0,
) -> Callable[
    [Callable[..., Any]], Callable[..., Any]
]:
    """Async version of retry decorator."""

    def decorator(
        func: Callable[..., Any]
    ) -> Callable[..., Any]:
        @wraps(func)
        async def wrapper(*args: Any, **kwargs: Any) -> Any:
            for attempt in range(max_attempts):
                try:
                    return await func(*args, **kwargs)
                except Exception as e:
                    if attempt == max_attempts - 1:
                        raise
                    logger.warning(f"Attempt {attempt + 1} failed: {e}")
                    await asyncio.sleep(delay)
            return None

        return wrapper

    return decorator


# Context managers
@contextmanager
def timer(name: str) -> Iterator[None]:
    """Context manager to measure execution time."""
    start = datetime.now()
    try:
        yield
    finally:
        elapsed = datetime.now() - start
        logger.info(f"{name} took {elapsed.total_seconds():.3f}s")


@asynccontextmanager
async def async_timer(name: str):
    """Async context manager to measure execution time."""
    start = datetime.now()
    try:
        yield
    finally:
        elapsed = datetime.now() - start
        logger.info(f"{name} took {elapsed.total_seconds():.3f}s")


# Overloaded functions
@overload
def process(value: str) -> str: ...


@overload
def process(value: int) -> int: ...


@overload
def process(value: list[T]) -> list[T]: ...


def process(value: str | int | list[T]) -> str | int | list[T]:
    """Process different types of values."""
    if isinstance(value, str):
        return value.upper()
    elif isinstance(value, int):
        return value * 2
    elif isinstance(value, list):
        return value[::-1]
    raise TypeError(f"Unsupported type: {type(value)}")


# Generator function
def fibonacci(n: int) -> Iterator[int]:
    """Generate Fibonacci numbers."""
    a, b = 0, 1
    for _ in range(n):
        yield a
        a, b = b, a + b


# Async generator
async def async_range(start: int, stop: int, delay: float = 0.1):
    """Async generator with delay."""
    for i in range(start, stop):
        await asyncio.sleep(delay)
        yield i


# Cached function
@lru_cache(maxsize=128)
def expensive_computation(n: int) -> int:
    """Cached expensive computation."""
    return sum(i**2 for i in range(n))


# Property decorator pattern
class Circle:
    """Circle with computed properties."""

    def __init__(self, radius: float):
        self._radius = radius

    @property
    def radius(self) -> float:
        return self._radius

    @radius.setter
    def radius(self, value: float) -> None:
        if value < 0:
            raise ValueError("Radius cannot be negative")
        self._radius = value

    @property
    def area(self) -> float:
        import math
        return math.pi * self._radius**2

    @property
    def circumference(self) -> float:
        import math
        return 2 * math.pi * self._radius


# Match statement (Python 3.10+)
def handle_command(command: dict[str, Any]) -> str:
    """Handle command using match statement."""
    match command:
        case {"action": "create", "name": name}:
            return f"Creating {name}"
        case {"action": "delete", "id": id_}:
            return f"Deleting {id_}"
        case {"action": "update", "id": id_, "data": data}:
            return f"Updating {id_} with {data}"
        case {"action": action}:
            return f"Unknown action: {action}"
        case _:
            return "Invalid command"


# Comprehensions
def comprehension_examples():
    """Examples of list, dict, set, and generator comprehensions."""
    # List comprehension
    squares = [x**2 for x in range(10)]

    # List comprehension with condition
    evens = [x for x in range(20) if x % 2 == 0]

    # Nested list comprehension
    matrix = [[i * j for j in range(5)] for i in range(5)]

    # Dict comprehension
    word_lengths = {word: len(word) for word in ["hello", "world", "python"]}

    # Set comprehension
    unique_lengths = {len(word) for word in ["a", "bb", "ccc", "dd"]}

    # Generator expression
    sum_of_squares = sum(x**2 for x in range(100))

    # Walrus operator (Python 3.8+)
    filtered = [y for x in range(10) if (y := x**2) > 10]

    return squares, evens, matrix, word_lengths, unique_lengths, sum_of_squares, filtered


# Async main function
async def main() -> None:
    """Main async entry point."""
    # String formatting
    name = "World"
    f_string = f"Hello, {name}!"
    format_string = "Value: {:.2f}".format(3.14159)

    # Multiline strings
    multiline = """
    This is a multiline string
    with multiple lines
    and preserved formatting
    """

    raw_string = r"This is a raw string with \n no escape"

    # Unpacking
    first, *middle, last = [1, 2, 3, 4, 5]

    # Dictionary unpacking
    defaults = {"timeout": 30, "retries": 3}
    config = {**defaults, "host": "localhost"}

    # Async operations
    async with async_timer("async operation"):
        results = []
        async for i in async_range(0, 5):
            results.append(i)

    # Using asyncio.gather
    tasks = [asyncio.create_task(asyncio.sleep(0.1)) for _ in range(3)]
    await asyncio.gather(*tasks)

    # Try/except/else/finally
    try:
        result = Result.success(42)
        value = result.value
    except Exception as e:
        logger.error(f"Error: {e}")
        value = 0
    else:
        logger.info(f"Success: {value}")
    finally:
        logger.info("Cleanup complete")

    print(f_string, format_string, multiline, raw_string, first, middle, last, config, results)


if __name__ == "__main__":
    asyncio.run(main())
