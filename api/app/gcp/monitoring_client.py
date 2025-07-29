"""
Cloud Operations (Monitoring, Tracing, Error Reporting) client.

This module provides integration with Google Cloud Operations suite for
metrics, distributed tracing, error reporting, and performance profiling.
"""

import asyncio
import time
from typing import Any, Dict, List, Optional, Union
from contextlib import asynccontextmanager
from functools import wraps

from google.api_core import retry, exceptions
from google.cloud import error_reporting
from google.cloud.trace_v2 import TraceServiceAsyncClient
from google.cloud.monitoring_v3 import (
    MetricServiceAsyncClient,
    AlertPolicyServiceAsyncClient,
)
from google.cloud.monitoring_v3.types import (
    TimeSeries,
    Point,
    TimeInterval,
    TypedValue,
    MetricDescriptor,
    LabelDescriptor,
    ValueType,
    MetricKind,
)
from google.protobuf import timestamp_pb2
from opentelemetry import trace, metrics
from opentelemetry.trace import Status, StatusCode
from opentelemetry.exporter.cloud_trace import CloudTraceSpanExporter
from opentelemetry.exporter.cloud_monitoring import CloudMonitoringMetricsExporter
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.sdk.metrics import MeterProvider
from opentelemetry.sdk.metrics.export import PeriodicExportingMetricReader

from app.config import settings
from app.utils.logger import setup_logger

logger = setup_logger(__name__)


class MonitoringClient:
    """
    Client for Google Cloud Operations (Monitoring, Tracing, Error Reporting).
    """

    def __init__(self):
        """Initialize monitoring clients."""
        self.project_id = settings.GCP_PROJECT_ID
        self.project_name = f"projects/{self.project_id}"

        # Service name for tracing
        self.service_name = "virtuoso-api"
        self.service_version = settings.VERSION

        # Initialize clients
        self._metric_client: Optional[MetricServiceAsyncClient] = None
        self._trace_client: Optional[TraceServiceAsyncClient] = None
        self._alert_client: Optional[AlertPolicyServiceAsyncClient] = None
        self._error_client: Optional[error_reporting.Client] = None

        # OpenTelemetry setup
        self._tracer_provider: Optional[TracerProvider] = None
        self._meter_provider: Optional[MeterProvider] = None
        self._tracer: Optional[trace.Tracer] = None
        self._meter: Optional[metrics.Meter] = None

        # Metric descriptors cache
        self._metric_descriptors: Dict[str, MetricDescriptor] = {}

        # Retry configuration
        self._retry = retry.AsyncRetry(
            initial=0.1,
            maximum=60.0,
            multiplier=2.0,
            timeout=300.0,
            predicate=retry.if_exception_type(
                exceptions.GoogleAPIError,
                asyncio.TimeoutError,
            ),
        )

        # Initialize OpenTelemetry
        self._setup_opentelemetry()

        logger.info(
            f"Initialized Cloud Operations client for project: {self.project_id}"
        )

    def _setup_opentelemetry(self):
        """Set up OpenTelemetry exporters and providers."""
        # Set up tracing
        self._tracer_provider = TracerProvider()
        trace.set_tracer_provider(self._tracer_provider)

        # Add Cloud Trace exporter
        cloud_trace_exporter = CloudTraceSpanExporter(project_id=self.project_id)
        span_processor = BatchSpanProcessor(cloud_trace_exporter)
        self._tracer_provider.add_span_processor(span_processor)

        # Get tracer
        self._tracer = trace.get_tracer(self.service_name, self.service_version)

        # Set up metrics
        cloud_monitoring_exporter = CloudMonitoringMetricsExporter(
            project_id=self.project_id
        )

        metric_reader = PeriodicExportingMetricReader(
            exporter=cloud_monitoring_exporter,
            export_interval_millis=60000,  # Export every minute
        )

        self._meter_provider = MeterProvider(metric_readers=[metric_reader])
        metrics.set_meter_provider(self._meter_provider)

        # Get meter
        self._meter = metrics.get_meter(self.service_name, self.service_version)

    async def _get_metric_client(self) -> MetricServiceAsyncClient:
        """Get or create metric client."""
        if self._metric_client is None:
            self._metric_client = MetricServiceAsyncClient()
        return self._metric_client

    async def _get_trace_client(self) -> TraceServiceAsyncClient:
        """Get or create trace client."""
        if self._trace_client is None:
            self._trace_client = TraceServiceAsyncClient()
        return self._trace_client

    async def _get_alert_client(self) -> AlertPolicyServiceAsyncClient:
        """Get or create alert policy client."""
        if self._alert_client is None:
            self._alert_client = AlertPolicyServiceAsyncClient()
        return self._alert_client

    def _get_error_client(self) -> error_reporting.Client:
        """Get or create error reporting client."""
        if self._error_client is None:
            self._error_client = error_reporting.Client(project=self.project_id)
        return self._error_client

    async def close(self):
        """Close all clients."""
        if self._metric_client:
            await self._metric_client.transport.close()
            self._metric_client = None

        if self._trace_client:
            await self._trace_client.transport.close()
            self._trace_client = None

        if self._alert_client:
            await self._alert_client.transport.close()
            self._alert_client = None

        if self._tracer_provider:
            self._tracer_provider.shutdown()

        if self._meter_provider:
            self._meter_provider.shutdown()

    # Distributed Tracing

    @asynccontextmanager
    async def trace_span(
        self,
        name: str,
        attributes: Optional[Dict[str, Any]] = None,
        kind: trace.SpanKind = trace.SpanKind.INTERNAL,
    ):
        """Create a trace span context manager."""
        with self._tracer.start_as_current_span(
            name, kind=kind, attributes=attributes or {}
        ) as span:
            try:
                yield span
            except Exception as e:
                span.record_exception(e)
                span.set_status(Status(StatusCode.ERROR, str(e)))
                raise
            finally:
                # Add additional attributes
                span.set_attribute("service.name", self.service_name)
                span.set_attribute("service.version", self.service_version)

    def trace_async(self, name: Optional[str] = None):
        """Decorator for tracing async functions."""

        def decorator(func):
            span_name = name or f"{func.__module__}.{func.__name__}"

            @wraps(func)
            async def wrapper(*args, **kwargs):
                async with self.trace_span(span_name):
                    return await func(*args, **kwargs)

            return wrapper

        return decorator

    def trace_sync(self, name: Optional[str] = None):
        """Decorator for tracing sync functions."""

        def decorator(func):
            span_name = name or f"{func.__module__}.{func.__name__}"

            @wraps(func)
            def wrapper(*args, **kwargs):
                with self._tracer.start_as_current_span(span_name):
                    return func(*args, **kwargs)

            return wrapper

        return decorator

    # Custom Metrics

    async def create_metric_descriptor(
        self,
        metric_type: str,
        display_name: str,
        description: str,
        value_type: ValueType = ValueType.INT64,
        metric_kind: MetricKind = MetricKind.GAUGE,
        unit: str = "1",
        labels: Optional[List[Dict[str, str]]] = None,
    ) -> MetricDescriptor:
        """Create a custom metric descriptor."""
        try:
            client = await self._get_metric_client()

            # Build metric type name
            type_name = f"custom.googleapis.com/{self.service_name}/{metric_type}"

            # Check if already exists
            if type_name in self._metric_descriptors:
                return self._metric_descriptors[type_name]

            # Build label descriptors
            label_descriptors = []
            if labels:
                for label in labels:
                    label_descriptors.append(
                        LabelDescriptor(
                            key=label["key"],
                            value_type=label.get("value_type", "STRING"),
                            description=label.get("description", ""),
                        )
                    )

            # Create descriptor
            descriptor = MetricDescriptor(
                type=type_name,
                display_name=display_name,
                description=description,
                metric_kind=metric_kind,
                value_type=value_type,
                unit=unit,
                labels=label_descriptors,
            )

            response = await client.create_metric_descriptor(
                name=self.project_name, metric_descriptor=descriptor, retry=self._retry
            )

            self._metric_descriptors[type_name] = response
            logger.info(f"Created metric descriptor: {type_name}")

            return response

        except exceptions.AlreadyExists:
            logger.info(f"Metric descriptor already exists: {type_name}")
            return self._metric_descriptors.get(type_name)
        except Exception as e:
            logger.error(f"Failed to create metric descriptor: {str(e)}")
            raise

    async def write_metric(
        self,
        metric_type: str,
        value: Union[int, float],
        labels: Optional[Dict[str, str]] = None,
        resource_type: str = "global",
        resource_labels: Optional[Dict[str, str]] = None,
    ):
        """Write a custom metric value."""
        try:
            client = await self._get_metric_client()

            # Build metric name
            type_name = f"custom.googleapis.com/{self.service_name}/{metric_type}"

            # Create time interval (now)
            now = time.time()
            interval = TimeInterval(end_time=timestamp_pb2.Timestamp(seconds=int(now)))

            # Create typed value
            if isinstance(value, int):
                typed_value = TypedValue(int64_value=value)
            else:
                typed_value = TypedValue(double_value=value)

            # Create point
            point = Point(interval=interval, value=typed_value)

            # Create time series
            series = TimeSeries(
                metric={"type": type_name, "labels": labels or {}},
                resource={"type": resource_type, "labels": resource_labels or {}},
                points=[point],
            )

            # Write time series
            await client.create_time_series(
                name=self.project_name, time_series=[series], retry=self._retry
            )

        except Exception as e:
            logger.error(f"Failed to write metric: {str(e)}")
            # Don't raise - metrics shouldn't break the application

    # Error Reporting

    def report_error(
        self,
        error: Exception,
        user: Optional[str] = None,
        http_context: Optional[Dict[str, Any]] = None,
    ):
        """Report an error to Cloud Error Reporting."""
        try:
            client = self._get_error_client()

            # Add context
            context = (
                error_reporting.HTTPContext(
                    method=http_context.get("method", ""),
                    url=http_context.get("url", ""),
                    user_agent=http_context.get("user_agent", ""),
                    remote_ip=http_context.get("remote_ip", ""),
                    referrer=http_context.get("referrer", ""),
                )
                if http_context
                else None
            )

            client.report_exception(http_context=context, user=user)

        except Exception as e:
            logger.error(f"Failed to report error: {str(e)}")
            # Don't raise - error reporting shouldn't break the application

    # Performance Metrics

    def create_counter(
        self, name: str, description: str, unit: str = "1"
    ) -> metrics.Counter:
        """Create a counter metric."""
        return self._meter.create_counter(
            name=f"{self.service_name}.{name}", description=description, unit=unit
        )

    def create_histogram(
        self, name: str, description: str, unit: str = "ms"
    ) -> metrics.Histogram:
        """Create a histogram metric."""
        return self._meter.create_histogram(
            name=f"{self.service_name}.{name}", description=description, unit=unit
        )

    def create_gauge(self, name: str, description: str, unit: str = "1"):
        """Create a gauge metric."""
        return self._meter.create_observable_gauge(
            name=f"{self.service_name}.{name}", description=description, unit=unit
        )

    # Alert Policies

    async def create_alert_policy(
        self,
        display_name: str,
        condition_filter: str,
        threshold_value: float,
        duration_seconds: int = 60,
        notification_channels: Optional[List[str]] = None,
    ):
        """Create an alert policy."""
        try:
            client = await self._get_alert_client()

            # Build condition
            condition = {
                "display_name": f"{display_name} condition",
                "condition_threshold": {
                    "filter": condition_filter,
                    "comparison": "COMPARISON_GT",
                    "threshold_value": threshold_value,
                    "duration": {"seconds": duration_seconds},
                    "aggregations": [
                        {
                            "alignment_period": {"seconds": 60},
                            "per_series_aligner": "ALIGN_MEAN",
                        }
                    ],
                },
            }

            # Build alert policy
            alert_policy = {
                "display_name": display_name,
                "conditions": [condition],
                "combiner": "OR",
                "enabled": True,
                "notification_channels": notification_channels or [],
            }

            response = await client.create_alert_policy(
                name=self.project_name, alert_policy=alert_policy, retry=self._retry
            )

            logger.info(f"Created alert policy: {display_name}")
            return response

        except Exception as e:
            logger.error(f"Failed to create alert policy: {str(e)}")
            raise

    # Monitoring Dashboard

    async def create_dashboard(self, display_name: str, widgets: List[Dict[str, Any]]):
        """Create a monitoring dashboard."""
        try:
            # Dashboard creation would typically use the Dashboard API
            # This is a placeholder for the implementation
            logger.info(f"Dashboard creation not implemented: {display_name}")

        except Exception as e:
            logger.error(f"Failed to create dashboard: {str(e)}")
            raise

    # Common Metrics for Virtuoso API

    async def track_command_execution(
        self,
        command: str,
        duration_ms: float,
        success: bool,
        user_id: Optional[str] = None,
    ):
        """Track command execution metrics."""
        labels = {"command": command, "status": "success" if success else "failure"}

        if user_id:
            labels["user_id"] = user_id

        # Write execution count
        await self.write_metric("command_executions", 1, labels=labels)

        # Write duration
        await self.write_metric(
            "command_duration_ms", duration_ms, labels={"command": command}
        )

    async def track_api_request(
        self, endpoint: str, method: str, status_code: int, duration_ms: float
    ):
        """Track API request metrics."""
        labels = {"endpoint": endpoint, "method": method, "status": str(status_code)}

        # Write request count
        await self.write_metric("api_requests", 1, labels=labels)

        # Write duration
        await self.write_metric(
            "api_request_duration_ms",
            duration_ms,
            labels={"endpoint": endpoint, "method": method},
        )

    async def track_session_activity(
        self,
        action: str,  # created, expired, active
        user_id: Optional[str] = None,
    ):
        """Track session activity metrics."""
        labels = {"action": action}

        if user_id:
            labels["user_id"] = user_id

        await self.write_metric("session_activity", 1, labels=labels)

    # Health Check

    async def health_check(self) -> Dict[str, Any]:
        """Check monitoring service connectivity."""
        try:
            # Try to list metric descriptors
            client = await self._get_metric_client()

            descriptors = []
            async for descriptor in client.list_metric_descriptors(
                name=self.project_name,
                filter=f'metric.type = starts_with("custom.googleapis.com/{self.service_name}")',
                page_size=10,
            ):
                descriptors.append(descriptor.type)

            return {
                "healthy": True,
                "project_id": self.project_id,
                "custom_metrics": len(descriptors),
                "tracing_enabled": self._tracer is not None,
                "metrics_enabled": self._meter is not None,
            }

        except Exception as e:
            return {"healthy": False, "error": str(e), "project_id": self.project_id}
