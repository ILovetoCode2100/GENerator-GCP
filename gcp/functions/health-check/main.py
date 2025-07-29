"""
Cloud Function for comprehensive health monitoring of Virtuoso services.
"""

import json
import asyncio
import os
from typing import Dict, Any, Tuple
from datetime import datetime
import aiohttp
from google.cloud import firestore, secretmanager
import redis.asyncio as redis
from flask import Request, Response
import logging

# Configure structured logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Environment variables
PROJECT_ID = os.environ.get("GCP_PROJECT", "virtuoso-api")
CLOUD_RUN_URL = os.environ.get(
    "CLOUD_RUN_URL", "https://virtuoso-api-service-abcdef.run.app"
)
REDIS_HOST = os.environ.get("REDIS_HOST", "10.0.0.3")
REDIS_PORT = int(os.environ.get("REDIS_PORT", "6379"))


class HealthChecker:
    """Performs health checks on various services."""

    def __init__(self):
        self.results: Dict[str, Dict[str, Any]] = {}
        self.start_time = datetime.utcnow()

    async def check_cloud_run(self) -> Tuple[str, Dict[str, Any]]:
        """Check Cloud Run service health."""
        service_name = "cloud_run"
        try:
            async with aiohttp.ClientSession() as session:
                async with session.get(
                    f"{CLOUD_RUN_URL}/health", timeout=aiohttp.ClientTimeout(total=5)
                ) as response:
                    if response.status == 200:
                        data = await response.json()
                        return service_name, {
                            "status": "healthy",
                            "response_time_ms": int(
                                (datetime.utcnow() - self.start_time).total_seconds()
                                * 1000
                            ),
                            "details": data,
                        }
                    else:
                        return service_name, {
                            "status": "unhealthy",
                            "error": f"HTTP {response.status}",
                            "response_time_ms": int(
                                (datetime.utcnow() - self.start_time).total_seconds()
                                * 1000
                            ),
                        }
        except Exception as e:
            logger.error(f"Cloud Run health check failed: {str(e)}")
            return service_name, {
                "status": "error",
                "error": str(e),
                "type": type(e).__name__,
            }

    async def check_firestore(self) -> Tuple[str, Dict[str, Any]]:
        """Check Firestore connectivity."""
        service_name = "firestore"
        try:
            start = datetime.utcnow()

            # Test write and read
            db = firestore.AsyncClient()
            test_doc_ref = db.collection("health_checks").document("test")

            # Write test document
            await test_doc_ref.set(
                {"timestamp": firestore.SERVER_TIMESTAMP, "test": True}
            )

            # Read test document
            doc = await test_doc_ref.get()
            if doc.exists:
                # Clean up
                await test_doc_ref.delete()

                response_time = int((datetime.utcnow() - start).total_seconds() * 1000)
                return service_name, {
                    "status": "healthy",
                    "response_time_ms": response_time,
                    "operations": ["write", "read", "delete"],
                }
            else:
                return service_name, {
                    "status": "unhealthy",
                    "error": "Test document not found",
                }

        except Exception as e:
            logger.error(f"Firestore health check failed: {str(e)}")
            return service_name, {
                "status": "error",
                "error": str(e),
                "type": type(e).__name__,
            }

    async def check_memorystore(self) -> Tuple[str, Dict[str, Any]]:
        """Check Memorystore (Redis) connectivity."""
        service_name = "memorystore"
        try:
            start = datetime.utcnow()

            # Connect to Redis
            r = await redis.from_url(
                f"redis://{REDIS_HOST}:{REDIS_PORT}", decode_responses=True
            )

            # Test operations
            test_key = "health_check:test"
            await r.set(test_key, "test_value", ex=10)
            value = await r.get(test_key)
            await r.delete(test_key)

            # Check memory usage
            info = await r.info("memory")

            await r.close()

            response_time = int((datetime.utcnow() - start).total_seconds() * 1000)

            return service_name, {
                "status": "healthy" if value == "test_value" else "unhealthy",
                "response_time_ms": response_time,
                "memory_used_mb": round(info.get("used_memory", 0) / 1024 / 1024, 2),
                "memory_peak_mb": round(
                    info.get("used_memory_peak", 0) / 1024 / 1024, 2
                ),
            }

        except Exception as e:
            logger.error(f"Memorystore health check failed: {str(e)}")
            return service_name, {
                "status": "error",
                "error": str(e),
                "type": type(e).__name__,
            }

    async def check_secret_manager(self) -> Tuple[str, Dict[str, Any]]:
        """Check Secret Manager access."""
        service_name = "secret_manager"
        try:
            start = datetime.utcnow()

            client = secretmanager.SecretManagerServiceAsyncClient()

            # List secrets (limited to first 5)
            parent = f"projects/{PROJECT_ID}"
            secrets = []

            async for secret in client.list_secrets(
                request={"parent": parent, "page_size": 5}
            ):
                secrets.append(secret.name.split("/")[-1])

            response_time = int((datetime.utcnow() - start).total_seconds() * 1000)

            return service_name, {
                "status": "healthy",
                "response_time_ms": response_time,
                "accessible_secrets_count": len(secrets),
                "sample_secrets": secrets[:3],  # Only show first 3 for security
            }

        except Exception as e:
            logger.error(f"Secret Manager health check failed: {str(e)}")
            return service_name, {
                "status": "error",
                "error": str(e),
                "type": type(e).__name__,
            }

    async def run_all_checks(self) -> Dict[str, Any]:
        """Run all health checks concurrently."""
        tasks = [
            self.check_cloud_run(),
            self.check_firestore(),
            self.check_memorystore(),
            self.check_secret_manager(),
        ]

        results = await asyncio.gather(*tasks, return_exceptions=True)

        health_status = {}
        overall_status = "healthy"
        unhealthy_services = []

        for result in results:
            if isinstance(result, Exception):
                logger.error(f"Health check task failed: {str(result)}")
                continue

            service_name, status = result
            health_status[service_name] = status

            if status.get("status") != "healthy":
                overall_status = "unhealthy"
                unhealthy_services.append(service_name)

        return {
            "timestamp": self.start_time.isoformat(),
            "overall_status": overall_status,
            "unhealthy_services": unhealthy_services,
            "services": health_status,
            "total_check_time_ms": int(
                (datetime.utcnow() - self.start_time).total_seconds() * 1000
            ),
        }


def health_check(request: Request) -> Response:
    """
    Cloud Function entry point for health monitoring.

    Args:
        request: The Flask request object

    Returns:
        Response with health check results
    """
    # Handle CORS
    if request.method == "OPTIONS":
        headers = {
            "Access-Control-Allow-Origin": "*",
            "Access-Control-Allow-Methods": "GET, POST",
            "Access-Control-Allow-Headers": "Content-Type",
            "Access-Control-Max-Age": "3600",
        }
        return Response("", 204, headers)

    headers = {"Access-Control-Allow-Origin": "*", "Content-Type": "application/json"}

    try:
        # Check for simple ping
        if request.args.get("ping"):
            return Response(
                json.dumps(
                    {"status": "ok", "timestamp": datetime.utcnow().isoformat()}
                ),
                200,
                headers,
            )

        # Run comprehensive health checks
        checker = HealthChecker()
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)

        try:
            results = loop.run_until_complete(checker.run_all_checks())

            # Determine HTTP status code
            status_code = 200 if results["overall_status"] == "healthy" else 503

            return Response(json.dumps(results, indent=2), status_code, headers)
        finally:
            loop.close()

    except Exception as e:
        logger.error(f"Health check failed: {str(e)}")
        error_response = {
            "timestamp": datetime.utcnow().isoformat(),
            "overall_status": "error",
            "error": str(e),
            "type": type(e).__name__,
        }
        return Response(json.dumps(error_response), 500, headers)


# For local testing
if __name__ == "__main__":
    from flask import Flask, request as flask_request

    app = Flask(__name__)

    @app.route("/", methods=["GET", "POST", "OPTIONS"])
    def main():
        return health_check(flask_request)

    app.run(debug=True, port=8080)
