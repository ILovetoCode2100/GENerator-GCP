"""
Rate limiting middleware for FastAPI.

This middleware implements Redis-based rate limiting with multiple strategies,
configurable limits per endpoint, and sliding window algorithm.
"""

import time
import hashlib
from typing import Optional, Dict, Tuple, Any
from datetime import datetime, timezone
from enum import Enum
from fastapi import Request, HTTPException, status
from fastapi.responses import JSONResponse
from redis import asyncio as aioredis
from pydantic import BaseModel, Field

from ..config import settings
from ..models.responses import ErrorResponse, ErrorType
from ..utils.logger import get_logger

# GCP imports
if settings.is_gcp_enabled:
    from ..gcp.firestore_client import FirestoreClient


logger = get_logger(__name__)

# Initialize GCP clients if enabled
memorystore_client = None
firestore_client = None

if settings.is_gcp_enabled:
    # Memorystore client would be initialized here if we had one
    # For now, we'll use Firestore as fallback
    if settings.USE_FIRESTORE:
        firestore_client = FirestoreClient()


class RateLimitStrategy(str, Enum):
    """Rate limiting strategies"""

    PER_USER = "per_user"
    PER_IP = "per_ip"
    PER_API_KEY = "per_api_key"
    PER_TENANT = "per_tenant"
    GLOBAL = "global"


class RateLimitConfig(BaseModel):
    """Configuration for rate limiting"""

    requests: int = Field(..., description="Number of allowed requests")
    window_seconds: int = Field(..., description="Time window in seconds")
    strategy: RateLimitStrategy = Field(
        RateLimitStrategy.PER_USER, description="Rate limit strategy"
    )
    burst_multiplier: float = Field(1.0, description="Burst allowance multiplier")
    include_headers: bool = Field(
        True, description="Include rate limit headers in response"
    )


# Default rate limit configurations per endpoint pattern
DEFAULT_RATE_LIMITS: Dict[str, RateLimitConfig] = {
    # Health endpoints - very high limit
    "/health": RateLimitConfig(
        requests=1000, window_seconds=60, strategy=RateLimitStrategy.PER_IP
    ),
    # List operations - moderate limits
    "/api/v1/projects": RateLimitConfig(
        requests=100, window_seconds=60, strategy=RateLimitStrategy.PER_USER
    ),
    "/api/v1/goals": RateLimitConfig(
        requests=100, window_seconds=60, strategy=RateLimitStrategy.PER_USER
    ),
    "/api/v1/journeys": RateLimitConfig(
        requests=100, window_seconds=60, strategy=RateLimitStrategy.PER_USER
    ),
    "/api/v1/checkpoints": RateLimitConfig(
        requests=100, window_seconds=60, strategy=RateLimitStrategy.PER_USER
    ),
    # Create operations - lower limits
    "/api/v1/projects/create": RateLimitConfig(
        requests=20, window_seconds=60, strategy=RateLimitStrategy.PER_USER
    ),
    "/api/v1/goals/create": RateLimitConfig(
        requests=30, window_seconds=60, strategy=RateLimitStrategy.PER_USER
    ),
    "/api/v1/journeys/create": RateLimitConfig(
        requests=30, window_seconds=60, strategy=RateLimitStrategy.PER_USER
    ),
    "/api/v1/checkpoints/create": RateLimitConfig(
        requests=30, window_seconds=60, strategy=RateLimitStrategy.PER_USER
    ),
    # Step operations - moderate limits with burst
    "/api/v1/commands/step": RateLimitConfig(
        requests=200,
        window_seconds=60,
        strategy=RateLimitStrategy.PER_USER,
        burst_multiplier=1.5,
    ),
    # Batch operations - lower limits
    "/api/v1/commands/batch": RateLimitConfig(
        requests=10, window_seconds=60, strategy=RateLimitStrategy.PER_USER
    ),
    "/api/v1/tests/run": RateLimitConfig(
        requests=5, window_seconds=60, strategy=RateLimitStrategy.PER_USER
    ),
    # Execution operations - very low limits
    "/api/v1/executions/start": RateLimitConfig(
        requests=5, window_seconds=300, strategy=RateLimitStrategy.PER_TENANT
    ),
    # Default for unmatched endpoints
    "default": RateLimitConfig(
        requests=100, window_seconds=60, strategy=RateLimitStrategy.PER_USER
    ),
}


class RateLimitExceeded(HTTPException):
    """Rate limit exceeded exception"""

    def __init__(
        self,
        detail: str,
        retry_after: int,
        headers: Optional[dict] = None,
        request_id: Optional[str] = None,
    ):
        headers = headers or {}
        headers["Retry-After"] = str(retry_after)

        super().__init__(
            status_code=status.HTTP_429_TOO_MANY_REQUESTS,
            detail=detail,
            headers=headers,
        )
        self.request_id = request_id
        self.retry_after = retry_after


class RateLimiter:
    """
    Redis-based rate limiter with sliding window algorithm.
    """

    def __init__(self, redis_url: Optional[str] = None):
        """
        Initialize rate limiter.

        Args:
            redis_url: Redis connection URL (defaults to localhost)
        """
        self.redis_url = redis_url or settings.REDIS_URL
        self._redis: Optional[aioredis.Redis] = None
        self._connected = False
        self._local_cache: Dict[str, Tuple[int, float]] = {}  # Fallback cache
        self._cache_ttl = 1.0  # Local cache TTL in seconds
        self._use_firestore_fallback = (
            settings.USE_FIRESTORE and firestore_client is not None
        )

    async def connect(self):
        """Connect to Redis."""
        if self._connected:
            return

        try:
            self._redis = await aioredis.from_url(
                self.redis_url, encoding="utf-8", decode_responses=True
            )
            await self._redis.ping()
            self._connected = True
            logger.info("Connected to Redis for rate limiting")
        except Exception as e:
            logger.warning(
                f"Failed to connect to Redis: {e}. Using local cache fallback."
            )
            self._connected = False

    async def disconnect(self):
        """Disconnect from Redis."""
        if self._redis and self._connected:
            await self._redis.close()
            self._connected = False

    def _get_key(self, identifier: str, endpoint: str, window_start: int) -> str:
        """
        Generate Redis key for rate limit tracking.

        Args:
            identifier: User/IP/tenant identifier
            endpoint: API endpoint
            window_start: Window start timestamp

        Returns:
            Redis key
        """
        # Hash the endpoint to handle long URLs
        endpoint_hash = hashlib.md5(endpoint.encode()).hexdigest()[:8]
        return f"rate_limit:{identifier}:{endpoint_hash}:{window_start}"

    async def _check_redis_limit(
        self,
        identifier: str,
        endpoint: str,
        config: RateLimitConfig,
        current_time: float,
    ) -> Tuple[bool, int, Dict[str, Any]]:
        """
        Check rate limit using Redis.

        Returns:
            Tuple of (is_allowed, remaining_requests, metadata)
        """
        if not self._connected:
            return await self._check_local_limit(
                identifier, endpoint, config, current_time
            )

        try:
            # Calculate window boundaries
            window_start = int(current_time - config.window_seconds)

            # Create pipeline for atomic operations
            pipe = self._redis.pipeline()

            # Key pattern for sliding window
            key_pattern = f"rate_limit:{identifier}:{hashlib.md5(endpoint.encode()).hexdigest()[:8]}:*"

            # Get all keys in the current window
            keys = await self._redis.keys(key_pattern)

            # Filter keys within the sliding window
            valid_keys = []
            for key in keys:
                key_time = int(key.split(":")[-1])
                if key_time >= window_start:
                    valid_keys.append(key)

            # Count requests in the window
            request_count = 0
            if valid_keys:
                request_count = sum(
                    int(await self._redis.get(key) or 0) for key in valid_keys
                )

            # Calculate allowed requests (with burst)
            allowed_requests = int(config.requests * config.burst_multiplier)

            # Check if within limit
            if request_count >= allowed_requests:
                retry_after = config.window_seconds
                return (
                    False,
                    0,
                    {
                        "limit": allowed_requests,
                        "remaining": 0,
                        "reset": int(current_time + retry_after),
                        "retry_after": retry_after,
                    },
                )

            # Increment counter for current second
            current_key = self._get_key(identifier, endpoint, int(current_time))
            await pipe.incr(current_key)
            await pipe.expire(
                current_key, config.window_seconds + 60
            )  # Extra time for cleanup
            await pipe.execute()

            remaining = allowed_requests - request_count - 1

            return (
                True,
                remaining,
                {
                    "limit": allowed_requests,
                    "remaining": remaining,
                    "reset": int(current_time + config.window_seconds),
                    "retry_after": 0,
                },
            )

        except Exception as e:
            logger.error(f"Redis error in rate limiting: {e}")
            # Try Firestore fallback if available
            if self._use_firestore_fallback:
                return await self._check_firestore_limit(
                    identifier, endpoint, config, current_time
                )
            # Otherwise fallback to local cache
            return await self._check_local_limit(
                identifier, endpoint, config, current_time
            )

    async def _check_local_limit(
        self,
        identifier: str,
        endpoint: str,
        config: RateLimitConfig,
        current_time: float,
    ) -> Tuple[bool, int, Dict[str, Any]]:
        """
        Check rate limit using local cache (fallback).

        Returns:
            Tuple of (is_allowed, remaining_requests, metadata)
        """
        cache_key = f"{identifier}:{endpoint}"

        # Clean expired entries
        expired_keys = [
            k
            for k, (_, timestamp) in self._local_cache.items()
            if current_time - timestamp > config.window_seconds
        ]
        for key in expired_keys:
            del self._local_cache[key]

        # Get current count
        count, first_request_time = self._local_cache.get(cache_key, (0, current_time))

        # Reset if window expired
        if current_time - first_request_time > config.window_seconds:
            count = 0
            first_request_time = current_time

        # Calculate allowed requests
        allowed_requests = int(config.requests * config.burst_multiplier)

        # Check limit
        if count >= allowed_requests:
            retry_after = int(
                config.window_seconds - (current_time - first_request_time)
            )
            return (
                False,
                0,
                {
                    "limit": allowed_requests,
                    "remaining": 0,
                    "reset": int(first_request_time + config.window_seconds),
                    "retry_after": retry_after,
                },
            )

        # Increment counter
        self._local_cache[cache_key] = (count + 1, first_request_time)
        remaining = allowed_requests - count - 1

        return (
            True,
            remaining,
            {
                "limit": allowed_requests,
                "remaining": remaining,
                "reset": int(first_request_time + config.window_seconds),
                "retry_after": 0,
            },
        )

    async def _check_firestore_limit(
        self,
        identifier: str,
        endpoint: str,
        config: RateLimitConfig,
        current_time: float,
    ) -> Tuple[bool, int, Dict[str, Any]]:
        """
        Check rate limit using Firestore (GCP fallback).

        Returns:
            Tuple of (is_allowed, remaining_requests, metadata)
        """
        if not firestore_client:
            return await self._check_local_limit(
                identifier, endpoint, config, current_time
            )

        try:
            # Create rate limit key
            rate_limit_key = f"rate_limit:{identifier}:{hashlib.md5(endpoint.encode()).hexdigest()[:8]}"

            # Get current rate limit data from Firestore
            rate_data = await firestore_client.cache_get(rate_limit_key)

            # Initialize or update rate data
            window_start = int(current_time - config.window_seconds)

            if rate_data:
                # Filter requests within window
                requests = rate_data.get("requests", [])
                valid_requests = [
                    req for req in requests if req["timestamp"] >= window_start
                ]
                request_count = len(valid_requests)
            else:
                valid_requests = []
                request_count = 0

            # Calculate allowed requests
            allowed_requests = int(config.requests * config.burst_multiplier)

            # Check if within limit
            if request_count >= allowed_requests:
                retry_after = config.window_seconds
                return (
                    False,
                    0,
                    {
                        "limit": allowed_requests,
                        "remaining": 0,
                        "reset": int(current_time + retry_after),
                        "retry_after": retry_after,
                    },
                )

            # Add current request
            valid_requests.append(
                {"timestamp": int(current_time), "endpoint": endpoint}
            )

            # Update Firestore
            await firestore_client.cache_set(
                key=rate_limit_key,
                value={
                    "requests": valid_requests,
                    "identifier": identifier,
                    "updated_at": datetime.now(timezone.utc).isoformat(),
                },
                ttl_seconds=config.window_seconds + 60,  # Extra time for cleanup
            )

            remaining = allowed_requests - request_count - 1

            return (
                True,
                remaining,
                {
                    "limit": allowed_requests,
                    "remaining": remaining,
                    "reset": int(current_time + config.window_seconds),
                    "retry_after": 0,
                },
            )

        except Exception as e:
            logger.error(f"Firestore rate limiting error: {e}")
            # Final fallback to local cache
            return await self._check_local_limit(
                identifier, endpoint, config, current_time
            )

    async def check_rate_limit(
        self, request: Request, config: RateLimitConfig
    ) -> Tuple[bool, Dict[str, Any]]:
        """
        Check if request is within rate limit.

        Args:
            request: FastAPI request
            config: Rate limit configuration

        Returns:
            Tuple of (is_allowed, metadata)
        """
        current_time = time.time()

        # Determine identifier based on strategy
        identifier = self._get_identifier(request, config.strategy)
        if not identifier:
            # Cannot determine identifier, allow request
            return True, {}

        # Get endpoint
        endpoint = request.url.path

        # Check limit
        is_allowed, remaining, metadata = await self._check_redis_limit(
            identifier, endpoint, config, current_time
        )

        return is_allowed, metadata

    def _get_identifier(
        self, request: Request, strategy: RateLimitStrategy
    ) -> Optional[str]:
        """Get identifier for rate limiting based on strategy."""
        if strategy == RateLimitStrategy.PER_IP:
            if request.client:
                return f"ip:{request.client.host}"
            return None

        elif strategy == RateLimitStrategy.PER_USER:
            if hasattr(request.state, "user") and request.state.user:
                return f"user:{request.state.user.user_id}"
            # Fallback to IP if no user
            if request.client:
                return f"ip:{request.client.host}"
            return None

        elif strategy == RateLimitStrategy.PER_API_KEY:
            if hasattr(request.state, "user") and request.state.user:
                if request.state.user.api_key_id:
                    return f"api_key:{request.state.user.api_key_id}"
            return None

        elif strategy == RateLimitStrategy.PER_TENANT:
            if hasattr(request.state, "user") and request.state.user:
                return f"tenant:{request.state.user.tenant_id}"
            return None

        elif strategy == RateLimitStrategy.GLOBAL:
            return "global"

        return None


class RateLimitMiddleware:
    """
    Rate limiting middleware for FastAPI.
    """

    def __init__(
        self,
        app,
        redis_url: Optional[str] = None,
        custom_limits: Optional[Dict[str, RateLimitConfig]] = None,
        exclude_paths: Optional[list] = None,
        enabled: bool = True,
    ):
        """
        Initialize rate limit middleware.

        Args:
            app: FastAPI application
            redis_url: Redis connection URL
            custom_limits: Custom rate limit configurations
            exclude_paths: Paths to exclude from rate limiting
            enabled: Whether rate limiting is enabled
        """
        self.app = app
        self.rate_limiter = RateLimiter(redis_url)
        self.rate_limits = DEFAULT_RATE_LIMITS.copy()
        if custom_limits:
            self.rate_limits.update(custom_limits)
        self.exclude_paths = exclude_paths or ["/docs", "/openapi.json", "/"]
        self.enabled = enabled and settings.RATE_LIMIT_ENABLED

    async def __call__(self, request: Request, call_next):
        """Process request through rate limiter."""
        # Skip if disabled
        if not self.enabled:
            return await call_next(request)

        # Skip excluded paths
        path = request.url.path
        if any(path.startswith(exclude) for exclude in self.exclude_paths):
            return await call_next(request)

        # Ensure rate limiter is connected
        await self.rate_limiter.connect()

        # Find matching rate limit config
        config = self._get_rate_limit_config(path)

        try:
            # Check rate limit
            is_allowed, metadata = await self.rate_limiter.check_rate_limit(
                request, config
            )

            if not is_allowed:
                # Rate limit exceeded
                retry_after = metadata.get("retry_after", config.window_seconds)

                error_response = ErrorResponse(
                    error_type=ErrorType.RATE_LIMIT,
                    message=f"Rate limit exceeded. Please retry after {retry_after} seconds.",
                    details=[
                        {
                            "message": "Rate limit exceeded",
                            "context": {
                                "limit": metadata.get("limit"),
                                "window_seconds": config.window_seconds,
                                "retry_after": retry_after,
                            },
                        }
                    ],
                    request_id=getattr(request.state, "request_id", None),
                )

                return JSONResponse(
                    status_code=status.HTTP_429_TOO_MANY_REQUESTS,
                    content=error_response.model_dump(),
                    headers={
                        "Retry-After": str(retry_after),
                        "X-RateLimit-Limit": str(
                            metadata.get("limit", config.requests)
                        ),
                        "X-RateLimit-Remaining": "0",
                        "X-RateLimit-Reset": str(metadata.get("reset", 0)),
                    },
                )

            # Process request
            response = await call_next(request)

            # Add rate limit headers if enabled
            if config.include_headers and metadata:
                response.headers["X-RateLimit-Limit"] = str(
                    metadata.get("limit", config.requests)
                )
                response.headers["X-RateLimit-Remaining"] = str(
                    metadata.get("remaining", 0)
                )
                response.headers["X-RateLimit-Reset"] = str(metadata.get("reset", 0))
                response.headers["X-RateLimit-Strategy"] = config.strategy.value

            return response

        except Exception as e:
            logger.error(f"Error in rate limit middleware: {e}")
            # On error, allow request to proceed
            return await call_next(request)

    def _get_rate_limit_config(self, path: str) -> RateLimitConfig:
        """Get rate limit config for a path."""
        # Try exact match first
        if path in self.rate_limits:
            return self.rate_limits[path]

        # Try prefix match
        for pattern, config in self.rate_limits.items():
            if pattern != "default" and path.startswith(pattern):
                return config

        # Return default
        return self.rate_limits.get(
            "default",
            RateLimitConfig(
                requests=settings.RATE_LIMIT_REQUESTS,
                window_seconds=settings.RATE_LIMIT_PERIOD,
            ),
        )


# Dependency for custom rate limits on specific endpoints
def rate_limit(
    requests: int,
    window_seconds: int = 60,
    strategy: RateLimitStrategy = RateLimitStrategy.PER_USER,
):
    """
    Dependency factory for custom rate limits on endpoints.

    Args:
        requests: Number of allowed requests
        window_seconds: Time window in seconds
        strategy: Rate limiting strategy

    Returns:
        Dependency function
    """
    config = RateLimitConfig(
        requests=requests, window_seconds=window_seconds, strategy=strategy
    )

    async def rate_limit_check(request: Request):
        # Create rate limiter instance
        rate_limiter = RateLimiter()
        await rate_limiter.connect()

        # Check rate limit
        is_allowed, metadata = await rate_limiter.check_rate_limit(request, config)

        if not is_allowed:
            retry_after = metadata.get("retry_after", window_seconds)
            raise RateLimitExceeded(
                detail=f"Rate limit exceeded. Maximum {requests} requests per {window_seconds} seconds.",
                retry_after=retry_after,
                request_id=getattr(request.state, "request_id", None),
            )

        # Add headers to request state for later use
        if not hasattr(request.state, "rate_limit_headers"):
            request.state.rate_limit_headers = {}

        request.state.rate_limit_headers.update(
            {
                "X-RateLimit-Limit": str(metadata.get("limit", requests)),
                "X-RateLimit-Remaining": str(metadata.get("remaining", 0)),
                "X-RateLimit-Reset": str(metadata.get("reset", 0)),
            }
        )

    return rate_limit_check
