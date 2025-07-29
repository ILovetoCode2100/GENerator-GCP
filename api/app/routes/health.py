"""
Health check endpoints.

These endpoints are public and do not require authentication.
"""

import os
import subprocess
import time
from datetime import datetime
from typing import Dict, Any

from fastapi import APIRouter, Query, Response
import redis.asyncio as aioredis
from prometheus_client import CONTENT_TYPE_LATEST

from ..config import settings
from ..models.responses import BaseResponse, ResponseStatus
from ..utils.logger import get_logger
from ..utils.monitoring import (
    health_monitor,
    system_monitor,
    collect_metrics,
    get_monitoring_stats,
)

router = APIRouter()
logger = get_logger(__name__)


async def check_cli_availability() -> Dict[str, Any]:
    """Check if CLI binary is available and executable."""
    start_time = time.time()
    details = {
        "path": settings.CLI_PATH,
        "exists": False,
        "executable": False,
        "version": None,
        "response_time_ms": 0,
    }

    try:
        # Check if file exists
        if not os.path.exists(settings.CLI_PATH):
            details["error"] = "CLI binary not found"
            return {"healthy": False, "details": details}

        details["exists"] = True

        # Check if executable
        if not os.access(settings.CLI_PATH, os.X_OK):
            details["error"] = "CLI binary not executable"
            return {"healthy": False, "details": details}

        details["executable"] = True

        # Try to run version command
        result = subprocess.run(
            [settings.CLI_PATH, "--version"], capture_output=True, text=True, timeout=5
        )

        if result.returncode == 0:
            details["version"] = result.stdout.strip()
            details["healthy"] = True
        else:
            details["error"] = f"Version check failed: {result.stderr}"
            details["healthy"] = False

    except subprocess.TimeoutExpired:
        details["error"] = "CLI version check timed out"
        details["healthy"] = False
    except Exception as e:
        logger.error(f"Error checking CLI availability: {e}")
        details["error"] = str(e)
        details["healthy"] = False

    finally:
        details["response_time_ms"] = (time.time() - start_time) * 1000
        health_monitor.record_health_check(
            "cli", details.get("healthy", False), details["response_time_ms"], details
        )

    return details


async def check_redis_connection() -> Dict[str, Any]:
    """Check if Redis is available (for rate limiting)."""
    start_time = time.time()
    details = {
        "url": settings.REDIS_URL.replace(settings.REDIS_PASSWORD or "", "***")
        if settings.REDIS_PASSWORD
        else settings.REDIS_URL,
        "required": settings.RATE_LIMIT_ENABLED,
        "response_time_ms": 0,
    }

    if not settings.RATE_LIMIT_ENABLED:
        details["healthy"] = True
        details["info"] = "Redis not required (rate limiting disabled)"
        return details

    try:
        redis = await aioredis.from_url(
            settings.REDIS_URL,
            socket_connect_timeout=settings.REDIS_TIMEOUT,
            password=settings.REDIS_PASSWORD,
        )

        # Ping Redis
        await redis.ping()

        # Get Redis info
        info = await redis.info()
        details["version"] = info.get("redis_version", "unknown")
        details["connected_clients"] = info.get("connected_clients", 0)
        details["used_memory_human"] = info.get("used_memory_human", "unknown")
        details["healthy"] = True

        await redis.close()

    except Exception as e:
        logger.warning(f"Redis connection check failed: {e}")
        details["error"] = str(e)
        details["healthy"] = False

    finally:
        details["response_time_ms"] = (time.time() - start_time) * 1000
        health_monitor.record_health_check(
            "redis", details.get("healthy", False), details["response_time_ms"], details
        )

    return details


async def check_database_connection() -> Dict[str, Any]:
    """Check database connection if configured."""
    # Placeholder for future database integration
    details = {"healthy": True, "info": "No database configured", "response_time_ms": 0}
    return details


async def check_virtuoso_api() -> Dict[str, Any]:
    """Check Virtuoso API connectivity."""
    start_time = time.time()
    details = {
        "base_url": settings.VIRTUOSO_BASE_URL,
        "configured": bool(settings.VIRTUOSO_API_KEY),
        "response_time_ms": 0,
    }

    if not settings.VIRTUOSO_API_KEY:
        details["healthy"] = True
        details["info"] = "Virtuoso API key not configured"
        return details

    try:
        # Try a simple API call (e.g., list projects with limit 1)
        import httpx

        async with httpx.AsyncClient(timeout=10.0) as client:
            response = await client.get(
                f"{settings.VIRTUOSO_BASE_URL}/projects",
                headers={"Authorization": f"Bearer {settings.VIRTUOSO_API_KEY}"},
                params={"limit": 1},
            )

            if response.status_code == 200:
                details["healthy"] = True
                details["status_code"] = response.status_code
            else:
                details["healthy"] = False
                details["status_code"] = response.status_code
                details["error"] = f"API returned status {response.status_code}"

    except Exception as e:
        logger.warning(f"Virtuoso API check failed: {e}")
        details["error"] = str(e)
        details["healthy"] = False

    finally:
        details["response_time_ms"] = (time.time() - start_time) * 1000
        health_monitor.record_health_check(
            "virtuoso_api",
            details.get("healthy", False),
            details["response_time_ms"],
            details,
        )

    return details


@router.get("/test")
async def simple_test():
    """Simple test endpoint."""
    return {"status": "ok", "message": "API is running"}


@router.get("/", response_model=BaseResponse[Dict[str, Any]])
async def health_check(
    detailed: bool = Query(False, description="Include detailed health information"),
) -> BaseResponse[Dict[str, Any]]:
    """
    Comprehensive health check endpoint.

    Args:
        detailed: If True, includes detailed service checks and system metrics

    Returns:
        Health status with service checks
    """
    start_time = time.time()

    # For now, return a simple health check
    response_data = {
        "healthy": True,
        "api_version": settings.VERSION,
        "environment": settings.ENVIRONMENT,
        "timestamp": datetime.utcnow().isoformat(),
        "response_time_ms": (time.time() - start_time) * 1000,
    }

    return BaseResponse(
        status=ResponseStatus.SUCCESS, data=response_data, message="Service is healthy"
    )


@router.get("/ready", response_model=BaseResponse[Dict[str, Any]])
async def readiness_check() -> BaseResponse[Dict[str, Any]]:
    """
    Readiness check endpoint.

    Verifies that all required dependencies are available.

    Returns:
        Readiness status
    """
    # Perform checks
    cli_check = await check_cli_availability()
    redis_check = await check_redis_connection()
    config_ready = bool(settings.VIRTUOSO_API_KEY or settings.API_KEYS)

    # All required services must be ready
    is_ready = (
        cli_check.get("healthy", False)
        and config_ready
        and (not settings.RATE_LIMIT_ENABLED or redis_check.get("healthy", False))
    )

    checks = {
        "cli": cli_check.get("healthy", False),
        "redis": redis_check.get("healthy", False),
        "config": config_ready,
        "virtuoso_api": bool(settings.VIRTUOSO_API_KEY),
    }

    return BaseResponse(
        status=ResponseStatus.SUCCESS if is_ready else ResponseStatus.ERROR,
        data={
            "ready": is_ready,
            "checks": checks,
            "version": settings.VERSION,
            "environment": settings.ENVIRONMENT,
            "timestamp": datetime.utcnow().isoformat(),
        },
        message="Service ready" if is_ready else "Service not ready",
    )


@router.get("/live", response_model=BaseResponse[Dict[str, Any]])
async def liveness_check() -> BaseResponse[Dict[str, Any]]:
    """
    Liveness check endpoint.

    Simple check to verify the service is running.

    Returns:
        Liveness status
    """
    uptime = system_monitor.get_uptime()

    return BaseResponse(
        status=ResponseStatus.SUCCESS,
        data={
            "alive": True,
            "uptime_seconds": uptime.total_seconds(),
            "uptime_human": f"{uptime.days}d {uptime.seconds // 3600}h {(uptime.seconds % 3600) // 60}m",
            "pid": os.getpid(),
            "timestamp": datetime.utcnow().isoformat(),
        },
        message="Service is alive",
    )


@router.get("/metrics")
async def metrics_endpoint(response: Response):
    """
    Prometheus metrics endpoint.

    Returns metrics in Prometheus text format.

    Returns:
        Prometheus formatted metrics
    """
    try:
        # Collect and return metrics
        metrics_data = await collect_metrics()
        response.headers["Content-Type"] = CONTENT_TYPE_LATEST
        return Response(content=metrics_data, media_type=CONTENT_TYPE_LATEST)
    except Exception as e:
        logger.error(f"Error generating metrics: {e}")
        return Response(
            content=f"# Error generating metrics: {str(e)}",
            media_type="text/plain",
            status_code=500,
        )


@router.get("/stats", response_model=BaseResponse[Dict[str, Any]])
async def monitoring_stats() -> BaseResponse[Dict[str, Any]]:
    """
    Get comprehensive monitoring statistics.

    Returns detailed statistics about system performance, health checks, and active tasks.

    Returns:
        Monitoring statistics
    """
    try:
        stats = get_monitoring_stats()

        return BaseResponse(
            status=ResponseStatus.SUCCESS,
            data=stats,
            message="Monitoring statistics retrieved successfully",
        )
    except Exception as e:
        logger.error(f"Error retrieving monitoring stats: {e}")
        return BaseResponse(
            status=ResponseStatus.ERROR,
            data={},
            message=f"Error retrieving statistics: {str(e)}",
        )
