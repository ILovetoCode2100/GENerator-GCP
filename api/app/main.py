"""
Virtuoso API CLI - FastAPI Application
Main entry point for the API that provides HTTP endpoints for Virtuoso CLI commands.
"""

import uuid
from contextlib import asynccontextmanager
from typing import Any, Dict

from fastapi import FastAPI, Request, status
from fastapi.middleware.cors import CORSMiddleware
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse
from starlette.middleware.base import BaseHTTPMiddleware
from starlette.exceptions import HTTPException as StarletteHTTPException

# Import OpenTelemetry for tracing support
try:
    import opentelemetry.trace
except ImportError:
    # OpenTelemetry is optional
    opentelemetry = None

# Import routers directly to avoid circular import
from app.routes import commands
from app.routes import tests  
from app.routes import sessions
from app.routes import health
from app.config import settings
from app.utils.logger import setup_logger

# Package metadata
__version__ = "1.0.0"
__title__ = "Virtuoso API CLI"
__description__ = """
## Virtuoso API CLI

A RESTful API wrapper for the Virtuoso CLI, providing programmatic access to test automation commands.

### Features

- **Command Execution**: Execute any Virtuoso CLI command via HTTP
- **Test Management**: Create and run test suites from YAML/JSON definitions
- **Session Management**: Manage test sessions and checkpoints
- **Health Monitoring**: Check API and CLI health status

### Authentication

Include your API key in the `X-API-Key` header for all requests.

### Rate Limiting

API requests are rate-limited. Check response headers for limit information.
"""

# Setup logger
logger = setup_logger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    """
    Manage application lifecycle events.
    """
    # Startup
    logger.info(f"Starting {__title__} v{__version__}")
    logger.info(f"Environment: {settings.ENVIRONMENT}")
    logger.info(f"Debug mode: {settings.DEBUG}")

    # Import monitoring utilities
    from app.utils.monitoring import start_metrics_background_task

    # Validate GCP settings if enabled
    if settings.is_gcp_enabled:
        try:
            settings.validate_gcp_settings()
            logger.info(f"GCP services enabled for project: {settings.GCP_PROJECT_ID}")
        except ValueError as e:
            logger.error(f"GCP configuration error: {str(e)}")
            if settings.is_production:
                raise

    # Initialize GCP clients if enabled
    gcp_clients = {}
    if settings.is_gcp_enabled:
        from app import gcp

        # Initialize Firestore if enabled
        if settings.USE_FIRESTORE:
            try:
                firestore_client = gcp.get_firestore_client()
                gcp_clients["firestore"] = firestore_client
                logger.info("Initialized Firestore client")

                # Clean up expired data on startup
                asyncio.create_task(firestore_client.cleanup_expired_data())
            except Exception as e:
                logger.error(f"Failed to initialize Firestore: {str(e)}")
                if settings.is_production:
                    raise

        # Initialize Cloud Tasks if enabled
        if settings.USE_CLOUD_TASKS:
            try:
                tasks_client = gcp.get_cloud_tasks_client()
                gcp_clients["cloud_tasks"] = tasks_client
                logger.info("Initialized Cloud Tasks client")

                # Create default queues
                for queue_name in tasks_client.QUEUES.values():
                    asyncio.create_task(tasks_client.create_queue(queue_name))
            except Exception as e:
                logger.error(f"Failed to initialize Cloud Tasks: {str(e)}")
                if settings.is_production:
                    raise

        # Initialize Pub/Sub if enabled
        if settings.USE_PUBSUB:
            try:
                pubsub_client = gcp.get_pubsub_client()
                gcp_clients["pubsub"] = pubsub_client
                logger.info("Initialized Pub/Sub client")

                # Create default topics
                for topic_name in set(pubsub_client.TOPICS.values()):
                    asyncio.create_task(pubsub_client.create_topic(topic_name))
                asyncio.create_task(
                    pubsub_client.create_topic(pubsub_client.DEAD_LETTER_TOPIC)
                )
            except Exception as e:
                logger.error(f"Failed to initialize Pub/Sub: {str(e)}")
                if settings.is_production:
                    raise

        # Initialize Secret Manager if enabled
        if settings.USE_SECRET_MANAGER:
            try:
                secret_client = gcp.get_secret_manager_client()
                gcp_clients["secret_manager"] = secret_client
                logger.info("Initialized Secret Manager client")

                # Load secrets on startup
                asyncio.create_task(secret_client.load_virtuoso_credentials())
            except Exception as e:
                logger.error(f"Failed to initialize Secret Manager: {str(e)}")
                if settings.is_production:
                    raise

        # Initialize Cloud Storage if enabled
        if settings.USE_CLOUD_STORAGE:
            try:
                storage_client = gcp.get_cloud_storage_client()
                gcp_clients["cloud_storage"] = storage_client
                logger.info("Initialized Cloud Storage client")

                # Create default buckets
                asyncio.create_task(storage_client.create_default_buckets())
            except Exception as e:
                logger.error(f"Failed to initialize Cloud Storage: {str(e)}")
                if settings.is_production:
                    raise

        # Initialize Cloud Monitoring if enabled
        if settings.USE_CLOUD_MONITORING:
            try:
                monitoring_client = gcp.get_monitoring_client()
                gcp_clients["monitoring"] = monitoring_client
                logger.info("Initialized Cloud Monitoring client")

                # Create custom metrics
                asyncio.create_task(
                    monitoring_client.create_metric_descriptor(
                        "command_executions",
                        "Command Executions",
                        "Number of CLI command executions",
                        labels=[
                            {"key": "command", "description": "Command name"},
                            {"key": "status", "description": "Execution status"},
                            {"key": "user_id", "description": "User ID"},
                        ],
                    )
                )
            except Exception as e:
                logger.error(f"Failed to initialize Cloud Monitoring: {str(e)}")
                if settings.is_production:
                    raise

    # Store GCP clients in app state
    app.state.gcp_clients = gcp_clients

    # Perform startup checks
    logger.info("Performing startup checks...")

    # Check CLI binary
    from app.routes.health import check_cli_availability
    cli_check = await check_cli_availability()
    if cli_check.get("healthy", False):
        logger.info(
            f"CLI binary found and working: {cli_check.get('version', 'unknown')}"
        )
    else:
        logger.error(
            f"CLI binary check failed: {cli_check.get('error', 'Unknown error')}"
        )
        if settings.is_production:
            raise RuntimeError("CLI binary is required in production")

    # Check Redis if rate limiting is enabled and not using Firestore
    if settings.RATE_LIMIT_ENABLED and not settings.USE_FIRESTORE:
        from app.routes.health import check_redis_connection
        redis_check = await check_redis_connection()
        if redis_check.get("healthy", False):
            logger.info(f"Redis connected: {redis_check.get('version', 'unknown')}")
        else:
            logger.warning(
                f"Redis connection failed: {redis_check.get('error', 'Unknown error')}"
            )
            if settings.is_production:
                logger.warning(
                    "Rate limiting will be disabled due to Redis unavailability"
                )

    # Warm up caches if enabled
    if settings.CACHE_ENABLED:
        logger.info("Cache warming not implemented yet")

    # Start monitoring background task
    import asyncio

    monitoring_task = asyncio.create_task(start_metrics_background_task())
    logger.info("Started metrics background task")

    yield

    # Shutdown
    logger.info("Shutting down application")

    # Cancel monitoring task
    monitoring_task.cancel()
    try:
        await monitoring_task
    except asyncio.CancelledError:
        pass

    # Close GCP clients
    if gcp_clients:
        logger.info("Closing GCP clients...")

        for client_name, client in gcp_clients.items():
            try:
                if hasattr(client, "close"):
                    await client.close()
                logger.info(f"Closed {client_name} client")
            except Exception as e:
                logger.error(f"Error closing {client_name} client: {str(e)}")

    logger.info("Cleanup completed")


# Create FastAPI application
app = FastAPI(
    title=__title__,
    description=__description__,
    version=__version__,
    docs_url="/docs",
    redoc_url="/redoc",
    openapi_url="/openapi.json",
    lifespan=lifespan,
    # Custom API documentation
    openapi_tags=[
        {
            "name": "commands",
            "description": "Execute Virtuoso CLI commands",
        },
        {
            "name": "tests",
            "description": "Manage and run test suites",
        },
        {
            "name": "sessions",
            "description": "Manage test sessions and checkpoints",
        },
        {
            "name": "health",
            "description": "Health check endpoints",
        },
    ],
)


# Request ID Middleware
class RequestIDMiddleware(BaseHTTPMiddleware):
    """
    Add request ID to all requests for tracing.
    """

    async def dispatch(self, request: Request, call_next):
        # Generate or extract request ID
        request_id = request.headers.get("X-Request-ID", str(uuid.uuid4()))

        # Add to request state
        request.state.request_id = request_id

        # Process request
        response = await call_next(request)

        # Add request ID to response headers
        response.headers["X-Request-ID"] = request_id

        return response


# Add middleware in order (innermost to outermost)
app.add_middleware(RequestIDMiddleware)

# Add rate limiting middleware
# TODO: Fix RateLimitMiddleware to work with ASGI
# app.add_middleware(
#     RateLimitMiddleware,
#     redis_url=settings.REDIS_URL,
#     exclude_paths=["/health", "/docs", "/redoc", "/openapi.json", "/"],
#     enabled=settings.RATE_LIMIT_ENABLED
# )

# Add authentication middleware
# TODO: Fix AuthMiddleware to work with ASGI
# app.add_middleware(
#     AuthMiddleware,
#     exclude_paths=["/health", "/docs", "/redoc", "/openapi.json", "/"],
#     require_auth_paths=["/api/v1"]  # All API v1 endpoints require auth
# )

# CORS configuration (outermost middleware)
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.CORS_ORIGINS,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
    expose_headers=[
        "X-Request-ID",
        "X-RateLimit-Limit",
        "X-RateLimit-Remaining",
        "X-RateLimit-Reset",
        "X-Authenticated-User",
        "X-Tenant-ID",
    ],
)


# Exception handlers
@app.exception_handler(StarletteHTTPException)
async def http_exception_handler(
    request: Request, exc: StarletteHTTPException
) -> JSONResponse:
    """
    Handle HTTP exceptions with consistent format.
    """
    request_id = getattr(request.state, "request_id", "unknown")

    return JSONResponse(
        status_code=exc.status_code,
        content={
            "error": {
                "message": exc.detail,
                "type": "http_error",
                "status_code": exc.status_code,
            },
            "request_id": request_id,
        },
    )


@app.exception_handler(RequestValidationError)
async def validation_exception_handler(
    request: Request, exc: RequestValidationError
) -> JSONResponse:
    """
    Handle request validation errors with detailed information.
    """
    request_id = getattr(request.state, "request_id", "unknown")

    # Format validation errors
    errors = []
    for error in exc.errors():
        errors.append(
            {
                "field": ".".join(str(loc) for loc in error["loc"]),
                "message": error["msg"],
                "type": error["type"],
            }
        )

    return JSONResponse(
        status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
        content={
            "error": {
                "message": "Validation error",
                "type": "validation_error",
                "details": errors,
            },
            "request_id": request_id,
        },
    )


@app.exception_handler(Exception)
async def general_exception_handler(request: Request, exc: Exception) -> JSONResponse:
    """
    Handle unexpected exceptions.
    """
    request_id = getattr(request.state, "request_id", "unknown")

    # Log the full exception
    logger.error(
        f"Unhandled exception for request {request_id}",
        exc_info=exc,
        extra={
            "request_id": request_id,
            "path": request.url.path,
            "method": request.method,
        },
    )

    # Report to GCP Error Reporting if enabled
    if settings.USE_CLOUD_MONITORING and hasattr(app.state, "gcp_clients"):
        monitoring_client = app.state.gcp_clients.get("monitoring")
        if monitoring_client:
            user_id = getattr(request.state, "user_id", None)
            monitoring_client.report_error(
                exc,
                user=user_id,
                http_context={
                    "method": request.method,
                    "url": str(request.url),
                    "user_agent": request.headers.get("user-agent", ""),
                    "remote_ip": request.client.host if request.client else "",
                    "referrer": request.headers.get("referer", ""),
                },
            )

    # Return generic error (don't expose internal details in production)
    error_detail = str(exc) if settings.DEBUG else "Internal server error"

    return JSONResponse(
        status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
        content={
            "error": {
                "message": error_detail,
                "type": "internal_error",
            },
            "request_id": request_id,
        },
    )


# Include routers
app.include_router(health.router, prefix="/health", tags=["health"])
app.include_router(commands.router, prefix="/api/v1/commands", tags=["commands"])
app.include_router(tests.router, prefix="/api/v1/tests", tags=["tests"])
app.include_router(sessions.router, prefix="/api/v1/sessions", tags=["sessions"])


# Root endpoint
@app.get("/", tags=["root"])
async def root() -> Dict[str, Any]:
    """
    API root endpoint with service information.
    """
    return {
        "service": __title__,
        "version": __version__,
        "description": "RESTful API for Virtuoso CLI",
        "documentation": {
            "swagger": "/docs",
            "redoc": "/redoc",
            "openapi": "/openapi.json",
        },
        "endpoints": {
            "health": "/health",
            "commands": "/api/v1/commands",
            "tests": "/api/v1/tests",
            "sessions": "/api/v1/sessions",
        },
    }


# Request logging middleware
@app.middleware("http")
async def log_requests(request: Request, call_next):
    """
    Log all incoming requests and responses.
    """
    # Import monitoring utilities
    from app.utils.monitoring import (
        request_count,
        request_duration,
        active_connections,
        error_count,
    )

    # Skip logging for health checks
    if request.url.path == "/health":
        return await call_next(request)

    # Get request ID
    request_id = getattr(request.state, "request_id", "unknown")

    # Track active connections
    active_connections.inc()

    # Log request
    logger.info(
        "Incoming request",
        extra={
            "request_id": request_id,
            "method": request.method,
            "path": request.url.path,
            "client": request.client.host if request.client else "unknown",
        },
    )

    # Track request start time
    import time

    start_time = time.time()

    # Create trace span if Cloud Monitoring is enabled
    monitoring_client = None
    span_context = None
    if settings.USE_CLOUD_MONITORING and hasattr(app.state, "gcp_clients"):
        monitoring_client = app.state.gcp_clients.get("monitoring")

    try:
        # Process request with tracing
        if monitoring_client:
            async with monitoring_client.trace_span(
                f"{request.method} {request.url.path}",
                attributes={
                    "http.method": request.method,
                    "http.url": str(request.url),
                    "http.client_ip": request.client.host
                    if request.client
                    else "unknown",
                    "request.id": request_id,
                },
                kind=opentelemetry.trace.SpanKind.SERVER if opentelemetry else None,
            ) as span:
                span_context = span
                response = await call_next(request)
                span.set_attribute("http.status_code", response.status_code)
        else:
            response = await call_next(request)

        # Track metrics
        duration = time.time() - start_time
        request_count.labels(
            method=request.method,
            endpoint=request.url.path,
            status=response.status_code,
        ).inc()
        request_duration.labels(
            method=request.method, endpoint=request.url.path
        ).observe(duration)

        # Track GCP metrics if enabled
        if monitoring_client:
            await monitoring_client.track_api_request(
                endpoint=request.url.path,
                method=request.method,
                status_code=response.status_code,
                duration_ms=duration * 1000,
            )

        # Log response
        logger.info(
            "Request completed",
            extra={
                "request_id": request_id,
                "status_code": response.status_code,
                "duration_ms": int(duration * 1000),
            },
        )

        return response

    except Exception as e:
        # Track error
        duration = time.time() - start_time
        error_count.labels(error_type=type(e).__name__, endpoint=request.url.path).inc()
        request_count.labels(
            method=request.method, endpoint=request.url.path, status=500
        ).inc()
        request_duration.labels(
            method=request.method, endpoint=request.url.path
        ).observe(duration)

        logger.error(
            "Request failed",
            extra={
                "request_id": request_id,
                "error": str(e),
                "duration_ms": int(duration * 1000),
            },
            exc_info=True,
        )

        raise

    finally:
        # Decrement active connections
        active_connections.dec()


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(
        "app.main:app",
        host=settings.HOST,
        port=settings.PORT,
        reload=settings.DEBUG,
        log_level=settings.LOG_LEVEL.lower(),
    )
