"""
Health check function for Virtuoso API CLI
Performs comprehensive health checks on all services
"""

import json
import logging
import os
from datetime import datetime

import functions_framework
from google.cloud import firestore
from google.cloud import storage
import redis
import requests

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


@functions_framework.http
def health_check(request):
    """
    Comprehensive health check for all Virtuoso services.

    Returns:
        JSON response with health status of all components
    """
    health_status = {
        "timestamp": datetime.utcnow().isoformat(),
        "overall_status": "healthy",
        "services": {},
        "checks_passed": 0,
        "checks_failed": 0,
    }

    # Check Cloud Run API
    try:
        api_url = os.environ.get("CLOUD_RUN_URL", "")
        if api_url:
            response = requests.get(f"{api_url}/health", timeout=5)
            health_status["services"]["cloud_run_api"] = {
                "status": "healthy" if response.status_code == 200 else "unhealthy",
                "response_time_ms": int(response.elapsed.total_seconds() * 1000),
                "status_code": response.status_code,
            }
            if response.status_code == 200:
                health_status["checks_passed"] += 1
            else:
                health_status["checks_failed"] += 1
                health_status["overall_status"] = "degraded"
    except Exception as e:
        logger.error(f"Cloud Run API check failed: {e}")
        health_status["services"]["cloud_run_api"] = {
            "status": "unhealthy",
            "error": str(e),
        }
        health_status["checks_failed"] += 1
        health_status["overall_status"] = "unhealthy"

    # Check Firestore
    try:
        db = firestore.Client()
        # Write and read a health check document
        health_ref = db.collection("health_checks").document("latest")
        health_ref.set(
            {"timestamp": firestore.SERVER_TIMESTAMP, "source": "health_check_function"}
        )
        doc = health_ref.get()

        health_status["services"]["firestore"] = {
            "status": "healthy" if doc.exists else "unhealthy",
            "can_write": True,
            "can_read": doc.exists,
        }
        health_status["checks_passed"] += 1
    except Exception as e:
        logger.error(f"Firestore check failed: {e}")
        health_status["services"]["firestore"] = {
            "status": "unhealthy",
            "error": str(e),
        }
        health_status["checks_failed"] += 1
        health_status["overall_status"] = "unhealthy"

    # Check Redis
    try:
        redis_host = os.environ.get("REDIS_HOST")
        redis_port = int(os.environ.get("REDIS_PORT", 6379))
        redis_auth = os.environ.get("REDIS_AUTH", "")

        if redis_host:
            r = redis.Redis(
                host=redis_host,
                port=redis_port,
                password=redis_auth,
                decode_responses=True,
                socket_connect_timeout=5,
            )
            # Ping Redis
            r.ping()
            # Set and get a test key
            r.setex("health_check", 60, "ok")
            value = r.get("health_check")

            health_status["services"]["redis"] = {
                "status": "healthy" if value == "ok" else "degraded",
                "can_connect": True,
                "can_write": True,
                "can_read": value == "ok",
            }
            health_status["checks_passed"] += 1
    except Exception as e:
        logger.error(f"Redis check failed: {e}")
        health_status["services"]["redis"] = {"status": "unhealthy", "error": str(e)}
        health_status["checks_failed"] += 1
        if health_status["overall_status"] == "healthy":
            health_status["overall_status"] = "degraded"

    # Check Cloud Storage
    try:
        storage_client = storage.Client()
        bucket_name = os.environ.get("STORAGE_BUCKET", "")

        if bucket_name:
            bucket = storage_client.bucket(bucket_name)
            # List a few blobs to verify access
            blobs = list(bucket.list_blobs(max_results=1))

            health_status["services"]["cloud_storage"] = {
                "status": "healthy",
                "bucket_accessible": True,
                "can_list": True,
            }
            health_status["checks_passed"] += 1
    except Exception as e:
        logger.error(f"Cloud Storage check failed: {e}")
        health_status["services"]["cloud_storage"] = {
            "status": "unhealthy",
            "error": str(e),
        }
        health_status["checks_failed"] += 1
        if health_status["overall_status"] == "healthy":
            health_status["overall_status"] = "degraded"

    # Determine HTTP status code
    if health_status["overall_status"] == "healthy":
        status_code = 200
    elif health_status["overall_status"] == "degraded":
        status_code = 200  # Still return 200 for partial degradation
    else:
        status_code = 503

    # Log the health check result
    logger.info(f"Health check completed: {health_status['overall_status']}")

    return (
        json.dumps(health_status, indent=2),
        status_code,
        {
            "Content-Type": "application/json",
            "Cache-Control": "no-cache, no-store, must-revalidate",
        },
    )
