"""
Cloud Function for processing usage analytics and generating reports.
"""

import json
import asyncio
import os
from typing import Dict, Any, List, Optional
from datetime import datetime, timedelta
from collections import defaultdict
from google.cloud import firestore, bigquery, storage
from flask import Request, Response
import logging

# Configure structured logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Environment variables
PROJECT_ID = os.environ.get("GCP_PROJECT", "virtuoso-api")
BIGQUERY_DATASET = os.environ.get("BIGQUERY_DATASET", "virtuoso_analytics")
ANALYTICS_BUCKET = os.environ.get("ANALYTICS_BUCKET", "virtuoso-api-analytics")
REPORT_RETENTION_DAYS = int(os.environ.get("REPORT_RETENTION_DAYS", "90"))


class AnalyticsProcessor:
    """Processes usage analytics and generates reports."""

    def __init__(self):
        self.db = firestore.AsyncClient()
        self.bq_client = bigquery.Client()
        self.storage_client = storage.Client()
        self.start_time = datetime.utcnow()

    async def aggregate_command_usage(
        self, start_date: Optional[datetime] = None, end_date: Optional[datetime] = None
    ) -> Dict[str, Any]:
        """Aggregate command usage statistics."""
        try:
            # Default to last 24 hours
            if not end_date:
                end_date = datetime.utcnow()
            if not start_date:
                start_date = end_date - timedelta(days=1)

            logger.info(f"Aggregating command usage from {start_date} to {end_date}")

            # Query command logs
            logs_ref = self.db.collection("command_logs")
            query = logs_ref.where("timestamp", ">=", start_date).where(
                "timestamp", "<=", end_date
            )

            # Aggregate data
            command_stats = defaultdict(
                lambda: {
                    "count": 0,
                    "success": 0,
                    "failure": 0,
                    "total_duration_ms": 0,
                    "users": set(),
                }
            )

            total_commands = 0

            async for doc in query.stream():
                data = doc.to_dict()
                command = data.get("command", "unknown")

                stats = command_stats[command]
                stats["count"] += 1
                total_commands += 1

                if data.get("success", False):
                    stats["success"] += 1
                else:
                    stats["failure"] += 1

                if duration := data.get("duration_ms"):
                    stats["total_duration_ms"] += duration

                if user_id := data.get("user_id"):
                    stats["users"].add(user_id)

            # Calculate final statistics
            final_stats = {}
            for command, stats in command_stats.items():
                final_stats[command] = {
                    "count": stats["count"],
                    "success_count": stats["success"],
                    "failure_count": stats["failure"],
                    "success_rate": round(stats["success"] / stats["count"] * 100, 2)
                    if stats["count"] > 0
                    else 0,
                    "avg_duration_ms": round(
                        stats["total_duration_ms"] / stats["count"]
                    )
                    if stats["count"] > 0
                    else 0,
                    "unique_users": len(stats["users"]),
                }

            return {
                "period": {
                    "start": start_date.isoformat(),
                    "end": end_date.isoformat(),
                },
                "total_commands": total_commands,
                "unique_commands": len(final_stats),
                "command_stats": final_stats,
                "top_commands": sorted(
                    final_stats.items(), key=lambda x: x[1]["count"], reverse=True
                )[:10],
            }

        except Exception as e:
            logger.error(f"Command usage aggregation failed: {str(e)}")
            raise

    async def calculate_api_metrics(
        self, start_date: Optional[datetime] = None, end_date: Optional[datetime] = None
    ) -> Dict[str, Any]:
        """Calculate API performance metrics."""
        try:
            # Default to last 24 hours
            if not end_date:
                end_date = datetime.utcnow()
            if not start_date:
                start_date = end_date - timedelta(days=1)

            logger.info(f"Calculating API metrics from {start_date} to {end_date}")

            # Query API logs
            api_logs_ref = self.db.collection("api_logs")
            query = api_logs_ref.where("timestamp", ">=", start_date).where(
                "timestamp", "<=", end_date
            )

            # Aggregate metrics
            endpoint_metrics = defaultdict(
                lambda: {
                    "requests": 0,
                    "success": 0,
                    "errors": defaultdict(int),
                    "response_times": [],
                    "methods": defaultdict(int),
                }
            )

            total_requests = 0
            total_errors = 0

            async for doc in query.stream():
                data = doc.to_dict()
                endpoint = data.get("endpoint", "unknown")

                metrics = endpoint_metrics[endpoint]
                metrics["requests"] += 1
                total_requests += 1

                if data.get("status_code", 500) < 400:
                    metrics["success"] += 1
                else:
                    total_errors += 1
                    error_code = data.get("status_code", "unknown")
                    metrics["errors"][str(error_code)] += 1

                if response_time := data.get("response_time_ms"):
                    metrics["response_times"].append(response_time)

                method = data.get("method", "GET")
                metrics["methods"][method] += 1

            # Calculate final metrics
            final_metrics = {}
            for endpoint, metrics in endpoint_metrics.items():
                response_times = metrics["response_times"]
                final_metrics[endpoint] = {
                    "total_requests": metrics["requests"],
                    "success_count": metrics["success"],
                    "error_count": metrics["requests"] - metrics["success"],
                    "success_rate": round(
                        metrics["success"] / metrics["requests"] * 100, 2
                    )
                    if metrics["requests"] > 0
                    else 0,
                    "avg_response_time_ms": round(
                        sum(response_times) / len(response_times)
                    )
                    if response_times
                    else 0,
                    "p95_response_time_ms": self._calculate_percentile(
                        response_times, 95
                    )
                    if response_times
                    else 0,
                    "p99_response_time_ms": self._calculate_percentile(
                        response_times, 99
                    )
                    if response_times
                    else 0,
                    "errors_by_code": dict(metrics["errors"]),
                    "methods": dict(metrics["methods"]),
                }

            # Overall metrics
            all_response_times = []
            for metrics in endpoint_metrics.values():
                all_response_times.extend(metrics["response_times"])

            return {
                "period": {
                    "start": start_date.isoformat(),
                    "end": end_date.isoformat(),
                },
                "overall": {
                    "total_requests": total_requests,
                    "total_errors": total_errors,
                    "error_rate": round(total_errors / total_requests * 100, 2)
                    if total_requests > 0
                    else 0,
                    "avg_response_time_ms": round(
                        sum(all_response_times) / len(all_response_times)
                    )
                    if all_response_times
                    else 0,
                    "p95_response_time_ms": self._calculate_percentile(
                        all_response_times, 95
                    )
                    if all_response_times
                    else 0,
                    "p99_response_time_ms": self._calculate_percentile(
                        all_response_times, 99
                    )
                    if all_response_times
                    else 0,
                },
                "endpoints": final_metrics,
                "top_endpoints": sorted(
                    final_metrics.items(),
                    key=lambda x: x[1]["total_requests"],
                    reverse=True,
                )[:10],
            }

        except Exception as e:
            logger.error(f"API metrics calculation failed: {str(e)}")
            raise

    def _calculate_percentile(self, values: List[float], percentile: int) -> float:
        """Calculate percentile value."""
        if not values:
            return 0

        sorted_values = sorted(values)
        index = int(len(sorted_values) * percentile / 100)
        return sorted_values[min(index, len(sorted_values) - 1)]

    async def generate_user_analytics(self) -> Dict[str, Any]:
        """Generate user activity analytics."""
        try:
            logger.info("Generating user analytics...")

            # Get active users in last 30 days
            cutoff_date = datetime.utcnow() - timedelta(days=30)

            users_ref = self.db.collection("users")
            active_query = users_ref.where("last_active", ">=", cutoff_date)

            user_stats = {
                "total_users": 0,
                "active_users_30d": 0,
                "new_users_7d": 0,
                "user_activity": defaultdict(
                    lambda: {"commands": 0, "projects": 0, "last_active": None}
                ),
            }

            # Count total users
            total_count_query = users_ref.count()
            user_stats["total_users"] = (await total_count_query.get())[0][0].value

            # Process active users
            async for doc in active_query.stream():
                user_data = doc.to_dict()
                user_id = doc.id

                user_stats["active_users_30d"] += 1

                # Check if new user (created in last 7 days)
                if created_at := user_data.get("created_at"):
                    if created_at >= datetime.utcnow() - timedelta(days=7):
                        user_stats["new_users_7d"] += 1

                # Get user activity
                activity = user_stats["user_activity"][user_id]
                activity["last_active"] = user_data.get("last_active")

                # Count user's commands
                command_count = (
                    await self.db.collection("command_logs")
                    .where("user_id", "==", user_id)
                    .where("timestamp", ">=", cutoff_date)
                    .count()
                    .get()
                )
                activity["commands"] = command_count[0][0].value if command_count else 0

                # Count user's projects
                project_count = (
                    await self.db.collection("projects")
                    .where("user_id", "==", user_id)
                    .count()
                    .get()
                )
                activity["projects"] = project_count[0][0].value if project_count else 0

            # Find most active users
            sorted_users = sorted(
                user_stats["user_activity"].items(),
                key=lambda x: x[1]["commands"],
                reverse=True,
            )

            return {
                "timestamp": datetime.utcnow().isoformat(),
                "summary": {
                    "total_users": user_stats["total_users"],
                    "active_users_30d": user_stats["active_users_30d"],
                    "new_users_7d": user_stats["new_users_7d"],
                    "activity_rate": round(
                        user_stats["active_users_30d"]
                        / user_stats["total_users"]
                        * 100,
                        2,
                    )
                    if user_stats["total_users"] > 0
                    else 0,
                },
                "top_users": [
                    {
                        "user_id": user_id,
                        "commands": stats["commands"],
                        "projects": stats["projects"],
                        "last_active": stats["last_active"].isoformat()
                        if stats["last_active"]
                        else None,
                    }
                    for user_id, stats in sorted_users[:10]
                ],
            }

        except Exception as e:
            logger.error(f"User analytics generation failed: {str(e)}")
            raise

    async def export_to_bigquery(self, data: Dict[str, Any], table_name: str) -> None:
        """Export analytics data to BigQuery."""
        try:
            dataset_id = f"{PROJECT_ID}.{BIGQUERY_DATASET}"
            table_id = f"{dataset_id}.{table_name}"

            # Prepare data for BigQuery
            rows_to_insert = [{**data, "processed_at": datetime.utcnow().isoformat()}]

            # Insert data
            errors = self.bq_client.insert_rows_json(table_id, rows_to_insert)

            if errors:
                logger.error(f"BigQuery insert errors: {errors}")
                raise Exception(f"Failed to insert data to BigQuery: {errors}")

            logger.info(f"Successfully exported data to {table_id}")

        except Exception as e:
            logger.error(f"BigQuery export failed: {str(e)}")
            # Don't raise - we don't want to fail the entire process

    async def save_report(self, report: Dict[str, Any], report_type: str) -> str:
        """Save analytics report to Cloud Storage."""
        try:
            # Generate filename
            timestamp = datetime.utcnow().strftime("%Y%m%d_%H%M%S")
            filename = f"reports/{report_type}/{timestamp}.json"

            # Get bucket
            bucket = self.storage_client.bucket(ANALYTICS_BUCKET)
            blob = bucket.blob(filename)

            # Upload report
            blob.upload_from_string(
                json.dumps(report, default=str, indent=2),
                content_type="application/json",
            )

            logger.info(f"Report saved to gs://{ANALYTICS_BUCKET}/{filename}")
            return f"gs://{ANALYTICS_BUCKET}/{filename}"

        except Exception as e:
            logger.error(f"Report save failed: {str(e)}")
            raise

    async def generate_comprehensive_report(
        self, period: str = "daily"
    ) -> Dict[str, Any]:
        """Generate a comprehensive analytics report."""
        # Determine time period
        end_date = datetime.utcnow()
        if period == "daily":
            start_date = end_date - timedelta(days=1)
        elif period == "weekly":
            start_date = end_date - timedelta(days=7)
        elif period == "monthly":
            start_date = end_date - timedelta(days=30)
        else:
            start_date = end_date - timedelta(days=1)

        logger.info(f"Generating {period} analytics report...")

        # Run all analytics
        tasks = [
            self.aggregate_command_usage(start_date, end_date),
            self.calculate_api_metrics(start_date, end_date),
            self.generate_user_analytics(),
        ]

        results = await asyncio.gather(*tasks, return_exceptions=True)

        report = {
            "report_type": f"{period}_analytics",
            "generated_at": datetime.utcnow().isoformat(),
            "period": {
                "type": period,
                "start": start_date.isoformat(),
                "end": end_date.isoformat(),
            },
        }

        # Process results
        if not isinstance(results[0], Exception):
            report["command_usage"] = results[0]
        else:
            report["command_usage"] = {"error": str(results[0])}

        if not isinstance(results[1], Exception):
            report["api_metrics"] = results[1]
        else:
            report["api_metrics"] = {"error": str(results[1])}

        if not isinstance(results[2], Exception):
            report["user_analytics"] = results[2]
        else:
            report["user_analytics"] = {"error": str(results[2])}

        # Save report
        report_url = await self.save_report(report, period)
        report["report_url"] = report_url

        # Export to BigQuery (optional)
        if not any(isinstance(r, Exception) for r in results):
            await self.export_to_bigquery(report, f"{period}_reports")

        return report


def analytics(request: Request) -> Response:
    """
    Cloud Function entry point for analytics processing.

    Args:
        request: The Flask request object

    Returns:
        Response with analytics results
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
        # Get parameters
        report_type = request.args.get("type", "command_usage")
        period = request.args.get("period", "daily")

        # Initialize processor
        processor = AnalyticsProcessor()
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)

        try:
            if report_type == "comprehensive":
                result = loop.run_until_complete(
                    processor.generate_comprehensive_report(period)
                )
            elif report_type == "command_usage":
                result = loop.run_until_complete(processor.aggregate_command_usage())
            elif report_type == "api_metrics":
                result = loop.run_until_complete(processor.calculate_api_metrics())
            elif report_type == "user_analytics":
                result = loop.run_until_complete(processor.generate_user_analytics())
            else:
                return Response(
                    json.dumps(
                        {
                            "error": "Invalid report type",
                            "valid_types": [
                                "comprehensive",
                                "command_usage",
                                "api_metrics",
                                "user_analytics",
                            ],
                        }
                    ),
                    400,
                    headers,
                )

            return Response(json.dumps(result, indent=2), 200, headers)

        finally:
            loop.close()

    except Exception as e:
        logger.error(f"Analytics function failed: {str(e)}")
        error_response = {
            "status": "error",
            "error": str(e),
            "type": type(e).__name__,
            "timestamp": datetime.utcnow().isoformat(),
        }
        return Response(json.dumps(error_response), 500, headers)


# For local testing
if __name__ == "__main__":
    from flask import Flask, request as flask_request

    app = Flask(__name__)

    @app.route("/", methods=["GET", "POST", "OPTIONS"])
    def main():
        return analytics(flask_request)

    app.run(debug=True, port=8080)
