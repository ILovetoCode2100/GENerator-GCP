"""
Analytics endpoints for usage insights and reporting.

These endpoints provide analytics data from BigQuery for monitoring
usage patterns, performance metrics, and generating reports.
"""

from typing import Dict, Any, List, Optional
from datetime import datetime, timedelta, timezone
from enum import Enum

from fastapi import APIRouter, HTTPException, status, Depends, Query
from pydantic import BaseModel, Field

from ..config import settings
from ..utils.logger import get_logger
from ..services.auth_service import AuthUser, Permission
from ..middleware.auth import get_authenticated_user, require_permissions
from ..middleware.rate_limit import rate_limit, RateLimitStrategy
from ..models.responses import BaseResponse, ResponseStatus

# GCP imports
if settings.is_gcp_enabled:
    from ..gcp.bigquery_client import BigQueryClient
    from ..gcp.firestore_client import FirestoreClient

router = APIRouter()
logger = get_logger(__name__)

# Initialize GCP clients if enabled
bigquery_client = None
firestore_client = None

if settings.is_gcp_enabled:
    if settings.USE_BIGQUERY:
        bigquery_client = BigQueryClient()
    if settings.USE_FIRESTORE:
        firestore_client = FirestoreClient()


class TimeRange(str, Enum):
    """Predefined time ranges for analytics."""

    LAST_HOUR = "last_hour"
    LAST_24_HOURS = "last_24_hours"
    LAST_7_DAYS = "last_7_days"
    LAST_30_DAYS = "last_30_days"
    CUSTOM = "custom"


class MetricType(str, Enum):
    """Types of metrics available."""

    COMMAND_COUNT = "command_count"
    SUCCESS_RATE = "success_rate"
    AVERAGE_DURATION = "average_duration"
    ERROR_RATE = "error_rate"
    USER_ACTIVITY = "user_activity"
    API_USAGE = "api_usage"


class AnalyticsQuery(BaseModel):
    """Query parameters for analytics."""

    metric: MetricType = Field(..., description="Metric type to query")
    time_range: TimeRange = Field(TimeRange.LAST_24_HOURS, description="Time range")
    start_time: Optional[datetime] = Field(None, description="Custom start time")
    end_time: Optional[datetime] = Field(None, description="Custom end time")
    group_by: Optional[List[str]] = Field(None, description="Group by fields")
    filters: Optional[Dict[str, Any]] = Field(None, description="Additional filters")


class AnalyticsReport(BaseModel):
    """Analytics report response."""

    metric: MetricType
    time_range: TimeRange
    start_time: datetime
    end_time: datetime
    data: List[Dict[str, Any]]
    summary: Dict[str, Any]
    generated_at: datetime


@router.post(
    "/query",
    response_model=BaseResponse[AnalyticsReport],
    dependencies=[
        Depends(require_permissions(Permission.READ_ANALYTICS)),
        Depends(rate_limit(10, 60, RateLimitStrategy.PER_USER)),
    ],
)
async def query_analytics(
    query: AnalyticsQuery, user: AuthUser = Depends(get_authenticated_user)
) -> BaseResponse[AnalyticsReport]:
    """
    Query analytics data based on specified metrics and filters.

    Args:
        query: Analytics query parameters
        user: Authenticated user

    Returns:
        Analytics report
    """
    if not settings.USE_BIGQUERY or not bigquery_client:
        raise HTTPException(
            status_code=status.HTTP_501_NOT_IMPLEMENTED,
            detail="Analytics not available. BigQuery not enabled.",
        )

    try:
        # Calculate time range
        end_time = query.end_time or datetime.now(timezone.utc)

        if query.time_range == TimeRange.CUSTOM:
            if not query.start_time:
                raise HTTPException(
                    status_code=status.HTTP_400_BAD_REQUEST,
                    detail="start_time required for custom time range",
                )
            start_time = query.start_time
        else:
            # Calculate start time based on range
            time_deltas = {
                TimeRange.LAST_HOUR: timedelta(hours=1),
                TimeRange.LAST_24_HOURS: timedelta(days=1),
                TimeRange.LAST_7_DAYS: timedelta(days=7),
                TimeRange.LAST_30_DAYS: timedelta(days=30),
            }
            start_time = end_time - time_deltas[query.time_range]

        # Build BigQuery query based on metric type
        bq_query = build_analytics_query(
            metric=query.metric,
            start_time=start_time,
            end_time=end_time,
            group_by=query.group_by,
            filters=query.filters,
            user_id=user.user_id if not user.is_admin else None,
        )

        # Execute query
        results = await bigquery_client.execute_query(bq_query)

        # Process results
        data = []
        summary = {}

        for row in results:
            data.append(dict(row))

        # Calculate summary statistics
        if data:
            if query.metric == MetricType.COMMAND_COUNT:
                summary["total"] = sum(row.get("count", 0) for row in data)
                summary["unique_commands"] = len(
                    set(row.get("command") for row in data if row.get("command"))
                )
            elif query.metric == MetricType.SUCCESS_RATE:
                total = sum(row.get("total", 0) for row in data)
                successful = sum(row.get("successful", 0) for row in data)
                summary["overall_success_rate"] = (
                    (successful / total * 100) if total > 0 else 0
                )
            elif query.metric == MetricType.AVERAGE_DURATION:
                durations = [
                    row.get("avg_duration_ms", 0)
                    for row in data
                    if row.get("avg_duration_ms")
                ]
                summary["overall_average"] = (
                    sum(durations) / len(durations) if durations else 0
                )

        report = AnalyticsReport(
            metric=query.metric,
            time_range=query.time_range,
            start_time=start_time,
            end_time=end_time,
            data=data,
            summary=summary,
            generated_at=datetime.now(timezone.utc),
        )

        return BaseResponse(
            status=ResponseStatus.SUCCESS,
            data=report,
            message="Analytics query executed successfully",
        )

    except Exception as e:
        logger.error(f"Analytics query failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Analytics query failed: {str(e)}",
        )


@router.get(
    "/dashboard",
    response_model=BaseResponse[Dict[str, Any]],
    dependencies=[Depends(require_permissions(Permission.READ_ANALYTICS))],
)
async def get_dashboard_stats(
    time_range: TimeRange = Query(TimeRange.LAST_24_HOURS),
    user: AuthUser = Depends(get_authenticated_user),
) -> BaseResponse[Dict[str, Any]]:
    """
    Get dashboard statistics for quick overview.

    Args:
        time_range: Time range for statistics
        user: Authenticated user

    Returns:
        Dashboard statistics
    """
    stats = {
        "total_commands": 0,
        "success_rate": 0.0,
        "active_users": 0,
        "top_commands": [],
        "error_summary": {},
        "recent_activity": [],
    }

    # Try cache first
    if settings.USE_FIRESTORE and firestore_client:
        cache_key = f"dashboard_stats:{user.tenant_id}:{time_range.value}"
        cached_stats = await firestore_client.cache_get(cache_key)
        if cached_stats:
            return BaseResponse(
                status=ResponseStatus.SUCCESS,
                data=cached_stats,
                message="Dashboard statistics (cached)",
            )

    # If BigQuery available, get real-time stats
    if settings.USE_BIGQUERY and bigquery_client:
        try:
            # Calculate time window
            end_time = datetime.now(timezone.utc)
            time_deltas = {
                TimeRange.LAST_HOUR: timedelta(hours=1),
                TimeRange.LAST_24_HOURS: timedelta(days=1),
                TimeRange.LAST_7_DAYS: timedelta(days=7),
                TimeRange.LAST_30_DAYS: timedelta(days=30),
            }
            start_time = end_time - time_deltas.get(time_range, timedelta(days=1))

            # Run parallel queries for different metrics
            queries = {
                "command_stats": f"""
                    SELECT
                        COUNT(*) as total_commands,
                        COUNTIF(success) as successful_commands,
                        COUNT(DISTINCT user_id) as active_users
                    FROM `{settings.GCP_PROJECT_ID}.analytics.command_executions`
                    WHERE timestamp BETWEEN '{start_time.isoformat()}' AND '{end_time.isoformat()}'
                    {f"AND tenant_id = '{user.tenant_id}'" if not user.is_admin else ""}
                """,
                "top_commands": f"""
                    SELECT
                        command,
                        COUNT(*) as count,
                        AVG(duration_ms) as avg_duration
                    FROM `{settings.GCP_PROJECT_ID}.analytics.command_executions`
                    WHERE timestamp BETWEEN '{start_time.isoformat()}' AND '{end_time.isoformat()}'
                    {f"AND tenant_id = '{user.tenant_id}'" if not user.is_admin else ""}
                    GROUP BY command
                    ORDER BY count DESC
                    LIMIT 10
                """,
                "error_summary": f"""
                    SELECT
                        error_type,
                        COUNT(*) as count
                    FROM `{settings.GCP_PROJECT_ID}.analytics.command_executions`
                    WHERE timestamp BETWEEN '{start_time.isoformat()}' AND '{end_time.isoformat()}'
                    AND success = FALSE
                    {f"AND tenant_id = '{user.tenant_id}'" if not user.is_admin else ""}
                    GROUP BY error_type
                    ORDER BY count DESC
                    LIMIT 10
                """,
            }

            # Execute queries
            results = {}
            for key, query in queries.items():
                try:
                    results[key] = await bigquery_client.execute_query(query)
                except Exception as e:
                    logger.error(f"Dashboard query {key} failed: {e}")
                    results[key] = []

            # Process results
            if results["command_stats"]:
                row = results["command_stats"][0]
                stats["total_commands"] = row.get("total_commands", 0)
                stats["active_users"] = row.get("active_users", 0)
                successful = row.get("successful_commands", 0)
                total = row.get("total_commands", 0)
                stats["success_rate"] = (successful / total * 100) if total > 0 else 0

            if results["top_commands"]:
                stats["top_commands"] = [
                    {
                        "command": row["command"],
                        "count": row["count"],
                        "avg_duration_ms": row.get("avg_duration", 0),
                    }
                    for row in results["top_commands"]
                ]

            if results["error_summary"]:
                stats["error_summary"] = {
                    row["error_type"]: row["count"] for row in results["error_summary"]
                }

            # Cache the results
            if settings.USE_FIRESTORE and firestore_client:
                await firestore_client.cache_set(
                    key=cache_key,
                    value=stats,
                    ttl_seconds=300,  # 5 minutes
                )

        except Exception as e:
            logger.error(f"Failed to get dashboard stats from BigQuery: {e}")

    return BaseResponse(
        status=ResponseStatus.SUCCESS,
        data=stats,
        message="Dashboard statistics retrieved",
    )


@router.post(
    "/report/generate",
    response_model=BaseResponse[Dict[str, str]],
    dependencies=[
        Depends(require_permissions(Permission.GENERATE_REPORTS)),
        Depends(rate_limit(5, 3600, RateLimitStrategy.PER_USER)),
    ],
)
async def generate_report(
    report_type: str,
    time_range: TimeRange = TimeRange.LAST_30_DAYS,
    format: str = "pdf",
    user: AuthUser = Depends(get_authenticated_user),
) -> BaseResponse[Dict[str, str]]:
    """
    Generate an analytics report.

    Args:
        report_type: Type of report to generate
        time_range: Time range for the report
        format: Output format (pdf, csv, excel)
        user: Authenticated user

    Returns:
        Report generation status and download URL
    """
    # This would integrate with Cloud Functions or App Engine
    # to generate reports asynchronously

    report_id = f"report_{user.tenant_id}_{datetime.now().strftime('%Y%m%d_%H%M%S')}"

    # Store report request
    if settings.USE_FIRESTORE and firestore_client:
        await firestore_client.create_report_request(
            report_id=report_id,
            user_id=user.user_id,
            tenant_id=user.tenant_id,
            report_type=report_type,
            time_range=time_range.value,
            format=format,
            status="pending",
        )

    return BaseResponse(
        status=ResponseStatus.SUCCESS,
        data={
            "report_id": report_id,
            "status": "generating",
            "estimated_time": "5-10 minutes",
            "download_url": f"/api/v1/analytics/report/{report_id}/download",
        },
        message="Report generation started",
    )


@router.get(
    "/report/{report_id}/status",
    response_model=BaseResponse[Dict[str, Any]],
    dependencies=[Depends(require_permissions(Permission.READ_ANALYTICS))],
)
async def get_report_status(
    report_id: str, user: AuthUser = Depends(get_authenticated_user)
) -> BaseResponse[Dict[str, Any]]:
    """
    Get status of a report generation request.

    Args:
        report_id: Report ID
        user: Authenticated user

    Returns:
        Report status
    """
    if settings.USE_FIRESTORE and firestore_client:
        report_data = await firestore_client.get_report_status(report_id)
        if report_data and report_data.get("user_id") == user.user_id:
            return BaseResponse(
                status=ResponseStatus.SUCCESS,
                data=report_data,
                message="Report status retrieved",
            )

    raise HTTPException(
        status_code=status.HTTP_404_NOT_FOUND, detail=f"Report not found: {report_id}"
    )


# Helper functions
def build_analytics_query(
    metric: MetricType,
    start_time: datetime,
    end_time: datetime,
    group_by: Optional[List[str]] = None,
    filters: Optional[Dict[str, Any]] = None,
    user_id: Optional[str] = None,
) -> str:
    """Build BigQuery SQL query for analytics."""

    # Base table based on metric
    table_map = {
        MetricType.COMMAND_COUNT: "command_executions",
        MetricType.SUCCESS_RATE: "command_executions",
        MetricType.AVERAGE_DURATION: "command_executions",
        MetricType.ERROR_RATE: "command_executions",
        MetricType.USER_ACTIVITY: "user_activity",
        MetricType.API_USAGE: "api_requests",
    }

    base_table = f"`{settings.GCP_PROJECT_ID}.analytics.{table_map[metric]}`"

    # Build SELECT clause based on metric
    select_clauses = {
        MetricType.COMMAND_COUNT: "command, COUNT(*) as count",
        MetricType.SUCCESS_RATE: "command, COUNTIF(success) as successful, COUNT(*) as total",
        MetricType.AVERAGE_DURATION: "command, AVG(duration_ms) as avg_duration_ms, MAX(duration_ms) as max_duration_ms",
        MetricType.ERROR_RATE: "error_type, COUNT(*) as count",
        MetricType.USER_ACTIVITY: "user_id, COUNT(*) as activity_count",
        MetricType.API_USAGE: "endpoint, method, COUNT(*) as request_count",
    }

    select_clause = select_clauses[metric]

    # Add group by fields
    if group_by:
        select_clause = f"{', '.join(group_by)}, {select_clause}"

    # Build WHERE clause
    where_conditions = [
        f"timestamp BETWEEN '{start_time.isoformat()}' AND '{end_time.isoformat()}'"
    ]

    if user_id:
        where_conditions.append(f"user_id = '{user_id}'")

    if filters:
        for key, value in filters.items():
            if isinstance(value, str):
                where_conditions.append(f"{key} = '{value}'")
            elif isinstance(value, list):
                values_str = "', '".join(str(v) for v in value)
                where_conditions.append(f"{key} IN ('{values_str}')")
            else:
                where_conditions.append(f"{key} = {value}")

    where_clause = " AND ".join(where_conditions)

    # Build GROUP BY clause
    group_by_fields = []
    if group_by:
        group_by_fields.extend(group_by)

    # Add metric-specific grouping
    metric_groups = {
        MetricType.COMMAND_COUNT: ["command"],
        MetricType.SUCCESS_RATE: ["command"],
        MetricType.AVERAGE_DURATION: ["command"],
        MetricType.ERROR_RATE: ["error_type"],
        MetricType.USER_ACTIVITY: ["user_id"],
        MetricType.API_USAGE: ["endpoint", "method"],
    }

    group_by_fields.extend(metric_groups[metric])
    group_by_clause = (
        f"GROUP BY {', '.join(set(group_by_fields))}" if group_by_fields else ""
    )

    # Build final query
    query = f"""
        SELECT {select_clause}
        FROM {base_table}
        WHERE {where_clause}
        {group_by_clause}
        ORDER BY count DESC
        LIMIT 1000
    """

    return query.strip()
