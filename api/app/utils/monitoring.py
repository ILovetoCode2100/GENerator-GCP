"""
Monitoring utilities for Prometheus metrics and system health tracking.

This module provides:
- Prometheus metrics collection
- Custom metrics for CLI commands
- Background task monitoring
- System resource tracking
"""

import asyncio
import os
import psutil
import time
from datetime import datetime, timedelta
from typing import Dict, Any, Optional, List
from functools import wraps

from prometheus_client import (
    Counter,
    Histogram,
    Gauge,
    Info,
    generate_latest,
    CollectorRegistry,
)

from ..config import settings
from ..utils.logger import get_logger

logger = get_logger(__name__)

# Create a custom registry to avoid conflicts
registry = CollectorRegistry()

# ========================================
# Prometheus Metrics
# ========================================

# Request metrics
request_count = Counter(
    "virtuoso_api_requests_total",
    "Total number of API requests",
    ["method", "endpoint", "status"],
    registry=registry,
)

request_duration = Histogram(
    "virtuoso_api_request_duration_seconds",
    "Request duration in seconds",
    ["method", "endpoint"],
    registry=registry,
)

# Command execution metrics
command_execution_count = Counter(
    "virtuoso_cli_commands_total",
    "Total number of CLI commands executed",
    ["command", "subcommand", "status"],
    registry=registry,
)

command_execution_duration = Histogram(
    "virtuoso_cli_command_duration_seconds",
    "CLI command execution duration in seconds",
    ["command", "subcommand"],
    registry=registry,
)

# System metrics
system_cpu_usage = Gauge(
    "virtuoso_api_cpu_usage_percent", "CPU usage percentage", registry=registry
)

system_memory_usage = Gauge(
    "virtuoso_api_memory_usage_bytes",
    "Memory usage in bytes",
    ["type"],  # rss, vms, available, percent
    registry=registry,
)

system_disk_usage = Gauge(
    "virtuoso_api_disk_usage_bytes",
    "Disk usage in bytes",
    ["mount_point", "type"],  # total, used, free, percent
    registry=registry,
)

# Service health metrics
service_health_status = Gauge(
    "virtuoso_api_service_health",
    "Service health status (1=healthy, 0=unhealthy)",
    ["service"],
    registry=registry,
)

service_response_time = Gauge(
    "virtuoso_api_service_response_time_ms",
    "Service response time in milliseconds",
    ["service"],
    registry=registry,
)

# Active connections and tasks
active_connections = Gauge(
    "virtuoso_api_active_connections", "Number of active connections", registry=registry
)

active_background_tasks = Gauge(
    "virtuoso_api_active_background_tasks",
    "Number of active background tasks",
    registry=registry,
)

# Error metrics
error_count = Counter(
    "virtuoso_api_errors_total",
    "Total number of errors",
    ["error_type", "endpoint"],
    registry=registry,
)

# Cache metrics
cache_hits = Counter(
    "virtuoso_api_cache_hits_total",
    "Total number of cache hits",
    ["cache_type"],
    registry=registry,
)

cache_misses = Counter(
    "virtuoso_api_cache_misses_total",
    "Total number of cache misses",
    ["cache_type"],
    registry=registry,
)

# API info
api_info = Info(
    "virtuoso_api_info", "API version and environment information", registry=registry
)

# Initialize API info
api_info.info(
    {
        "version": settings.VERSION,
        "environment": settings.ENVIRONMENT,
        "python_version": os.sys.version.split()[0],
    }
)

# ========================================
# Monitoring Decorators
# ========================================


def track_request_metrics(method: str, endpoint: str):
    """Decorator to track request metrics."""

    def decorator(func):
        @wraps(func)
        async def wrapper(*args, **kwargs):
            start_time = time.time()
            status = "success"

            try:
                result = await func(*args, **kwargs)
                return result
            except Exception:
                status = "error"
                raise
            finally:
                duration = time.time() - start_time
                request_count.labels(
                    method=method, endpoint=endpoint, status=status
                ).inc()
                request_duration.labels(method=method, endpoint=endpoint).observe(
                    duration
                )

        return wrapper

    return decorator


def track_command_metrics(command: str, subcommand: str = ""):
    """Decorator to track CLI command execution metrics."""

    def decorator(func):
        @wraps(func)
        async def wrapper(*args, **kwargs):
            start_time = time.time()
            status = "success"

            try:
                result = await func(*args, **kwargs)
                return result
            except Exception as e:
                status = "error"
                error_count.labels(
                    error_type=type(e).__name__, endpoint=f"{command}/{subcommand}"
                ).inc()
                raise
            finally:
                duration = time.time() - start_time
                command_execution_count.labels(
                    command=command, subcommand=subcommand, status=status
                ).inc()
                command_execution_duration.labels(
                    command=command, subcommand=subcommand
                ).observe(duration)

        return wrapper

    return decorator


# ========================================
# System Monitoring
# ========================================


class SystemMonitor:
    """Monitor system resources and health."""

    def __init__(self):
        self.process = psutil.Process()
        self._start_time = datetime.utcnow()

    def get_uptime(self) -> timedelta:
        """Get application uptime."""
        return datetime.utcnow() - self._start_time

    def get_cpu_usage(self) -> float:
        """Get current CPU usage percentage."""
        return self.process.cpu_percent(interval=0.1)

    def get_memory_info(self) -> Dict[str, Any]:
        """Get memory usage information."""
        memory = self.process.memory_info()
        virtual_memory = psutil.virtual_memory()

        return {
            "rss": memory.rss,  # Resident Set Size
            "vms": memory.vms,  # Virtual Memory Size
            "available": virtual_memory.available,
            "percent": virtual_memory.percent,
            "total": virtual_memory.total,
        }

    def get_disk_usage(self, path: str = "/") -> Dict[str, Any]:
        """Get disk usage information."""
        usage = psutil.disk_usage(path)

        return {
            "total": usage.total,
            "used": usage.used,
            "free": usage.free,
            "percent": usage.percent,
        }

    def get_network_connections(self) -> int:
        """Get number of network connections."""
        return len(self.process.connections())

    def get_open_files(self) -> int:
        """Get number of open files."""
        try:
            return len(self.process.open_files())
        except (psutil.AccessDenied, psutil.NoSuchProcess):
            return 0

    def get_thread_count(self) -> int:
        """Get number of threads."""
        return self.process.num_threads()

    async def update_metrics(self):
        """Update Prometheus metrics with current system stats."""
        try:
            # CPU usage
            cpu_usage = self.get_cpu_usage()
            system_cpu_usage.set(cpu_usage)

            # Memory usage
            memory_info = self.get_memory_info()
            system_memory_usage.labels(type="rss").set(memory_info["rss"])
            system_memory_usage.labels(type="vms").set(memory_info["vms"])
            system_memory_usage.labels(type="available").set(memory_info["available"])
            system_memory_usage.labels(type="percent").set(memory_info["percent"])

            # Disk usage
            disk_info = self.get_disk_usage()
            system_disk_usage.labels(mount_point="/", type="total").set(
                disk_info["total"]
            )
            system_disk_usage.labels(mount_point="/", type="used").set(
                disk_info["used"]
            )
            system_disk_usage.labels(mount_point="/", type="free").set(
                disk_info["free"]
            )
            system_disk_usage.labels(mount_point="/", type="percent").set(
                disk_info["percent"]
            )

            # Connection count
            active_connections.set(self.get_network_connections())

        except Exception as e:
            logger.error(f"Error updating system metrics: {e}")


# ========================================
# Health Check Monitoring
# ========================================


class HealthCheckMonitor:
    """Monitor health check results and response times."""

    def __init__(self):
        self.health_history: Dict[str, List[Dict[str, Any]]] = {}
        self.max_history = 100  # Keep last 100 checks per service

    def record_health_check(
        self,
        service: str,
        healthy: bool,
        response_time_ms: float,
        details: Optional[Dict] = None,
    ):
        """Record a health check result."""
        # Update Prometheus metrics
        service_health_status.labels(service=service).set(1 if healthy else 0)
        service_response_time.labels(service=service).set(response_time_ms)

        # Store in history
        if service not in self.health_history:
            self.health_history[service] = []

        check_result = {
            "timestamp": datetime.utcnow().isoformat(),
            "healthy": healthy,
            "response_time_ms": response_time_ms,
            "details": details or {},
        }

        self.health_history[service].append(check_result)

        # Trim history
        if len(self.health_history[service]) > self.max_history:
            self.health_history[service] = self.health_history[service][
                -self.max_history :
            ]

    def get_service_stats(self, service: str) -> Dict[str, Any]:
        """Get statistics for a service."""
        if service not in self.health_history or not self.health_history[service]:
            return {
                "checks_total": 0,
                "healthy_checks": 0,
                "health_percentage": 0.0,
                "avg_response_time_ms": 0.0,
                "last_check": None,
            }

        history = self.health_history[service]
        healthy_count = sum(1 for check in history if check["healthy"])
        response_times = [check["response_time_ms"] for check in history]

        return {
            "checks_total": len(history),
            "healthy_checks": healthy_count,
            "health_percentage": (healthy_count / len(history)) * 100,
            "avg_response_time_ms": sum(response_times) / len(response_times),
            "min_response_time_ms": min(response_times),
            "max_response_time_ms": max(response_times),
            "last_check": history[-1] if history else None,
        }


# ========================================
# Background Task Monitoring
# ========================================


class BackgroundTaskMonitor:
    """Monitor background tasks."""

    def __init__(self):
        self.tasks: Dict[str, Dict[str, Any]] = {}

    def start_task(self, task_id: str, task_name: str, details: Optional[Dict] = None):
        """Record task start."""
        self.tasks[task_id] = {
            "name": task_name,
            "started_at": datetime.utcnow(),
            "status": "running",
            "details": details or {},
        }
        active_background_tasks.inc()

    def complete_task(
        self, task_id: str, success: bool = True, result: Optional[Any] = None
    ):
        """Record task completion."""
        if task_id in self.tasks:
            task = self.tasks[task_id]
            task["completed_at"] = datetime.utcnow()
            task["duration_ms"] = (
                task["completed_at"] - task["started_at"]
            ).total_seconds() * 1000
            task["status"] = "completed" if success else "failed"
            task["result"] = result
            active_background_tasks.dec()

    def get_active_tasks(self) -> List[Dict[str, Any]]:
        """Get list of active tasks."""
        return [
            {
                "id": task_id,
                "name": task["name"],
                "running_for_ms": (
                    datetime.utcnow() - task["started_at"]
                ).total_seconds()
                * 1000,
                "details": task["details"],
            }
            for task_id, task in self.tasks.items()
            if task["status"] == "running"
        ]

    def cleanup_old_tasks(self, max_age_hours: int = 24):
        """Remove completed tasks older than max_age_hours."""
        cutoff_time = datetime.utcnow() - timedelta(hours=max_age_hours)

        to_remove = []
        for task_id, task in self.tasks.items():
            if (
                task["status"] != "running"
                and task.get("completed_at", datetime.utcnow()) < cutoff_time
            ):
                to_remove.append(task_id)

        for task_id in to_remove:
            del self.tasks[task_id]


# ========================================
# Global Instances
# ========================================

system_monitor = SystemMonitor()
health_monitor = HealthCheckMonitor()
task_monitor = BackgroundTaskMonitor()


# ========================================
# Metrics Collection
# ========================================


async def collect_metrics() -> bytes:
    """Collect all metrics and return in Prometheus format."""
    # Update system metrics
    await system_monitor.update_metrics()

    # Generate metrics
    return generate_latest(registry)


async def start_metrics_background_task():
    """Start background task to periodically update metrics."""
    while True:
        try:
            await system_monitor.update_metrics()
            task_monitor.cleanup_old_tasks()
        except Exception as e:
            logger.error(f"Error in metrics background task: {e}")

        await asyncio.sleep(30)  # Update every 30 seconds


# ========================================
# Utility Functions
# ========================================


def track_cache_access(cache_type: str, hit: bool):
    """Track cache hit/miss."""
    if hit:
        cache_hits.labels(cache_type=cache_type).inc()
    else:
        cache_misses.labels(cache_type=cache_type).inc()


def increment_error_count(error_type: str, endpoint: str):
    """Increment error counter."""
    error_count.labels(error_type=error_type, endpoint=endpoint).inc()


def get_monitoring_stats() -> Dict[str, Any]:
    """Get comprehensive monitoring statistics."""
    return {
        "system": {
            "uptime_seconds": system_monitor.get_uptime().total_seconds(),
            "cpu_usage_percent": system_monitor.get_cpu_usage(),
            "memory": system_monitor.get_memory_info(),
            "disk": system_monitor.get_disk_usage(),
            "connections": system_monitor.get_network_connections(),
            "threads": system_monitor.get_thread_count(),
            "open_files": system_monitor.get_open_files(),
        },
        "tasks": {
            "active": len(task_monitor.get_active_tasks()),
            "tasks": task_monitor.get_active_tasks(),
        },
        "health_checks": {
            service: health_monitor.get_service_stats(service)
            for service in health_monitor.health_history
        },
    }
