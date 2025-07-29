"""
Logging configuration and utilities.
"""

import logging
import sys
from typing import Optional
from pythonjsonlogger import jsonlogger


def setup_logger(
    name: Optional[str] = None,
    level: str = "INFO",
    format_str: Optional[str] = None,
    json_format: bool = True,
) -> logging.Logger:
    """
    Setup a logger with consistent configuration.

    Args:
        name: Logger name (defaults to root logger)
        level: Log level (DEBUG, INFO, WARNING, ERROR, CRITICAL)
        format_str: Custom format string
        json_format: Use JSON formatting for production

    Returns:
        Configured logger instance
    """
    logger = logging.getLogger(name)

    # Don't add handlers if they already exist
    if logger.handlers:
        return logger

    # Set log level
    logger.setLevel(getattr(logging, level.upper()))

    # Create console handler
    handler = logging.StreamHandler(sys.stdout)
    handler.setLevel(getattr(logging, level.upper()))

    # Configure formatter
    if json_format and level != "DEBUG":
        # JSON format for production
        json_formatter = jsonlogger.JsonFormatter(
            "%(asctime)s %(name)s %(levelname)s %(message)s",
            rename_fields={
                "asctime": "timestamp",
                "name": "logger",
                "levelname": "level",
                "message": "msg",
            },
        )
        handler.setFormatter(json_formatter)
    else:
        # Human-readable format for development
        if not format_str:
            format_str = "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
        formatter = logging.Formatter(format_str)
        handler.setFormatter(formatter)

    # Add handler to logger
    logger.addHandler(handler)

    # Prevent propagation to avoid duplicate logs
    logger.propagate = False

    return logger


def get_logger(name: str) -> logging.Logger:
    """
    Get a logger instance with the given name.

    Args:
        name: Logger name (usually __name__)

    Returns:
        Logger instance
    """
    return logging.getLogger(name)


class RequestLogger:
    """
    Context-aware logger for request handling.
    """

    def __init__(self, logger: logging.Logger, request_id: str):
        self.logger = logger
        self.request_id = request_id

    def _add_context(self, kwargs: dict) -> dict:
        """Add request context to log kwargs."""
        if "extra" not in kwargs:
            kwargs["extra"] = {}
        kwargs["extra"]["request_id"] = self.request_id
        return kwargs

    def debug(self, msg: str, *args, **kwargs):
        """Log debug message with request context."""
        self.logger.debug(msg, *args, **self._add_context(kwargs))

    def info(self, msg: str, *args, **kwargs):
        """Log info message with request context."""
        self.logger.info(msg, *args, **self._add_context(kwargs))

    def warning(self, msg: str, *args, **kwargs):
        """Log warning message with request context."""
        self.logger.warning(msg, *args, **self._add_context(kwargs))

    def error(self, msg: str, *args, **kwargs):
        """Log error message with request context."""
        self.logger.error(msg, *args, **self._add_context(kwargs))

    def critical(self, msg: str, *args, **kwargs):
        """Log critical message with request context."""
        self.logger.critical(msg, *args, **self._add_context(kwargs))
