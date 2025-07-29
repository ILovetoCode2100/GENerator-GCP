#!/usr/bin/env python3
"""
Test script for health check and monitoring endpoints.

This script tests the enhanced health check endpoints to ensure
comprehensive monitoring is working correctly.
"""

import asyncio
import httpx
import sys
from typing import Dict, Any


BASE_URL = "http://localhost:8000"


async def test_endpoint(
    client: httpx.AsyncClient, method: str, path: str, **kwargs
) -> Dict[str, Any]:
    """Test an endpoint and return the result."""
    try:
        response = await client.request(method, f"{BASE_URL}{path}", **kwargs)
        return {
            "status_code": response.status_code,
            "success": response.is_success,
            "data": response.json()
            if response.headers.get("content-type", "").startswith("application/json")
            else response.text,
            "headers": dict(response.headers),
        }
    except Exception as e:
        return {"status_code": None, "success": False, "error": str(e)}


async def run_tests():
    """Run all health check tests."""
    async with httpx.AsyncClient(timeout=30.0) as client:
        print("Testing Health Check Endpoints\n" + "=" * 50)

        # Test 1: Basic health check
        print("\n1. Basic Health Check (/health)")
        result = await test_endpoint(client, "GET", "/health")
        print(f"   Status: {result['status_code']}")
        if result["success"]:
            data = result["data"]["data"]
            print(f"   Healthy: {data['healthy']}")
            print(f"   Version: {data['api_version']}")
            print(f"   Environment: {data['environment']}")
        else:
            print(f"   Error: {result.get('error', 'Unknown error')}")

        # Test 2: Detailed health check
        print("\n2. Detailed Health Check (/health?detailed=true)")
        result = await test_endpoint(
            client, "GET", "/health", params={"detailed": "true"}
        )
        print(f"   Status: {result['status_code']}")
        if result["success"]:
            data = result["data"]["data"]
            print(f"   Healthy: {data['healthy']}")

            # Service checks
            if "services" in data:
                print("\n   Service Status:")
                for service, info in data["services"].items():
                    if isinstance(info, dict):
                        healthy = info.get("healthy", info.get("has_api_keys", False))
                        print(f"     - {service}: {'✓' if healthy else '✗'}")
                        if "error" in info:
                            print(f"       Error: {info['error']}")
                        if "response_time_ms" in info:
                            print(
                                f"       Response time: {info['response_time_ms']:.2f}ms"
                            )

            # System metrics
            if "system" in data:
                print("\n   System Metrics:")
                system = data["system"]
                print(f"     - Uptime: {system.get('uptime_seconds', 0):.0f}s")
                print(f"     - CPU Usage: {system.get('cpu_usage_percent', 0):.1f}%")
                if "memory" in system:
                    mem = system["memory"]
                    print(f"     - Memory: {mem.get('percent', 0):.1f}% used")
                print(f"     - Connections: {system.get('connections', 0)}")
                print(f"     - Threads: {system.get('threads', 0)}")

        # Test 3: Readiness check
        print("\n3. Readiness Check (/health/ready)")
        result = await test_endpoint(client, "GET", "/health/ready")
        print(f"   Status: {result['status_code']}")
        if result["success"]:
            data = result["data"]["data"]
            print(f"   Ready: {data['ready']}")
            if "checks" in data:
                for check, status in data["checks"].items():
                    print(f"   - {check}: {'✓' if status else '✗'}")

        # Test 4: Liveness check
        print("\n4. Liveness Check (/health/live)")
        result = await test_endpoint(client, "GET", "/health/live")
        print(f"   Status: {result['status_code']}")
        if result["success"]:
            data = result["data"]["data"]
            print(f"   Alive: {data['alive']}")
            print(f"   Uptime: {data.get('uptime_human', 'N/A')}")
            print(f"   PID: {data.get('pid', 'N/A')}")

        # Test 5: Prometheus metrics
        print("\n5. Prometheus Metrics (/health/metrics)")
        result = await test_endpoint(client, "GET", "/health/metrics")
        print(f"   Status: {result['status_code']}")
        if result["success"]:
            metrics = result["data"]
            if isinstance(metrics, str):
                # Count metric lines
                metric_lines = [
                    line
                    for line in metrics.split("\n")
                    if line and not line.startswith("#")
                ]
                print(f"   Metrics exported: {len(metric_lines)}")

                # Show some example metrics
                print("\n   Sample metrics:")
                for line in metric_lines[:5]:
                    print(f"     {line}")
                if len(metric_lines) > 5:
                    print(f"     ... and {len(metric_lines) - 5} more")

        # Test 6: Monitoring stats
        print("\n6. Monitoring Statistics (/health/stats)")
        result = await test_endpoint(client, "GET", "/health/stats")
        print(f"   Status: {result['status_code']}")
        if result["success"]:
            data = result["data"]["data"]

            if "system" in data:
                print("\n   System Stats:")
                system = data["system"]
                print(f"     - CPU: {system.get('cpu_usage_percent', 0):.1f}%")
                print(
                    f"     - Memory: {system.get('memory', {}).get('percent', 0):.1f}%"
                )
                print(f"     - Open files: {system.get('open_files', 0)}")

            if "tasks" in data:
                print(f"\n   Active Tasks: {data['tasks'].get('active', 0)}")

            if "health_checks" in data:
                print("\n   Health Check History:")
                for service, stats in data["health_checks"].items():
                    if stats.get("checks_total", 0) > 0:
                        print(f"     - {service}:")
                        print(f"       Total checks: {stats['checks_total']}")
                        print(f"       Health rate: {stats['health_percentage']:.1f}%")
                        print(
                            f"       Avg response: {stats['avg_response_time_ms']:.2f}ms"
                        )

        print("\n" + "=" * 50)
        print("Health check testing completed!")


if __name__ == "__main__":
    print("Starting health check endpoint tests...")
    print(f"Target: {BASE_URL}")
    print("\nMake sure the API server is running!")
    print("You can start it with: cd api && uvicorn app.main:app --reload")
    print("\n" + "-" * 50)

    try:
        asyncio.run(run_tests())
    except KeyboardInterrupt:
        print("\nTest interrupted by user")
        sys.exit(1)
    except Exception as e:
        print(f"\nTest failed with error: {e}")
        sys.exit(1)
