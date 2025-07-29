"""
Analytics function for Virtuoso API CLI
Processes events and generates analytics reports
"""

import json
import logging
import os
from datetime import datetime, timedelta
from typing import Dict, Any

import functions_framework
from google.cloud import firestore
from google.cloud import storage
from google.cloud import bigquery
from google.cloud import pubsub_v1

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Initialize clients
db = firestore.Client()
storage_client = storage.Client()
subscriber = pubsub_v1.SubscriberClient()


def process_command_metrics(start_time: datetime, end_time: datetime) -> Dict[str, Any]:
    """
    Calculate command execution metrics for the given time period.
    """
    metrics = {
        "total_commands": 0,
        "successful_commands": 0,
        "failed_commands": 0,
        "command_types": {},
        "average_duration_ms": 0,
        "error_types": {},
    }

    # Query command history
    commands_ref = db.collection("command_history")
    query = commands_ref.where("timestamp", ">=", start_time).where(
        "timestamp", "<", end_time
    )

    total_duration = 0
    command_count = 0

    for doc in query.stream():
        data = doc.to_dict()
        metrics["total_commands"] += 1

        # Track success/failure
        if data.get("result", {}).get("success", False):
            metrics["successful_commands"] += 1
        else:
            metrics["failed_commands"] += 1
            error_type = data.get("result", {}).get("error_type", "unknown")
            metrics["error_types"][error_type] = (
                metrics["error_types"].get(error_type, 0) + 1
            )

        # Track command types
        command_type = data.get("command", "unknown")
        metrics["command_types"][command_type] = (
            metrics["command_types"].get(command_type, 0) + 1
        )

        # Track duration
        duration = data.get("duration_ms", 0)
        if duration > 0:
            total_duration += duration
            command_count += 1

    # Calculate average duration
    if command_count > 0:
        metrics["average_duration_ms"] = total_duration / command_count

    return metrics


def process_test_metrics(start_time: datetime, end_time: datetime) -> Dict[str, Any]:
    """
    Calculate test execution metrics for the given time period.
    """
    metrics = {
        "total_tests": 0,
        "passed_tests": 0,
        "failed_tests": 0,
        "average_steps": 0,
        "average_duration_s": 0,
        "failure_reasons": {},
    }

    # Query test runs
    tests_ref = db.collection("test_runs")
    query = tests_ref.where("created_at", ">=", start_time).where(
        "created_at", "<", end_time
    )

    total_steps = 0
    total_duration = 0
    test_count = 0

    for doc in query.stream():
        data = doc.to_dict()
        metrics["total_tests"] += 1

        # Track pass/fail
        status = data.get("status", "unknown")
        if status == "passed":
            metrics["passed_tests"] += 1
        elif status == "failed":
            metrics["failed_tests"] += 1
            failure_reason = data.get("failure_reason", "unknown")
            metrics["failure_reasons"][failure_reason] = (
                metrics["failure_reasons"].get(failure_reason, 0) + 1
            )

        # Track steps
        steps = len(data.get("steps", []))
        total_steps += steps

        # Track duration
        start = data.get("timestamps", {}).get("started")
        end = data.get("timestamps", {}).get("completed")
        if start and end:
            duration = (end - start).total_seconds()
            total_duration += duration
            test_count += 1

    # Calculate averages
    if metrics["total_tests"] > 0:
        metrics["average_steps"] = total_steps / metrics["total_tests"]
    if test_count > 0:
        metrics["average_duration_s"] = total_duration / test_count

    return metrics


def process_user_metrics(start_time: datetime, end_time: datetime) -> Dict[str, Any]:
    """
    Calculate user activity metrics for the given time period.
    """
    metrics = {
        "active_users": set(),
        "total_sessions": 0,
        "average_session_duration_m": 0,
        "top_users": {},
    }

    # Query sessions
    sessions_ref = db.collection("sessions")
    query = sessions_ref.where("created_at", ">=", start_time).where(
        "created_at", "<", end_time
    )

    total_duration = 0
    session_count = 0

    for doc in query.stream():
        data = doc.to_dict()
        user_id = data.get("user_id", "anonymous")

        metrics["active_users"].add(user_id)
        metrics["total_sessions"] += 1

        # Track top users
        metrics["top_users"][user_id] = metrics["top_users"].get(user_id, 0) + 1

        # Track session duration
        created = data.get("created_at")
        updated = data.get("updated_at", created)
        if created and updated:
            duration = (updated - created).total_seconds() / 60  # Convert to minutes
            total_duration += duration
            session_count += 1

    # Calculate average duration
    if session_count > 0:
        metrics["average_session_duration_m"] = total_duration / session_count

    # Convert set to count
    metrics["active_users"] = len(metrics["active_users"])

    # Sort and limit top users
    metrics["top_users"] = dict(
        sorted(metrics["top_users"].items(), key=lambda x: x[1], reverse=True)[:10]
    )

    return metrics


@functions_framework.cloud_event
def process_analytics(cloud_event):
    """
    Process analytics based on Pub/Sub event trigger.
    """
    try:
        # Decode Pub/Sub message
        message = json.loads(cloud_event.data["message"]["data"])

        # Determine time period
        event_type = message.get("eventType", "")
        if event_type == "system.metrics.aggregate":
            period = message.get("data", {}).get("aggregation_period", "hourly")
        else:
            period = "hourly"

        # Calculate time range
        end_time = datetime.utcnow()
        if period == "hourly":
            start_time = end_time - timedelta(hours=1)
        elif period == "daily":
            start_time = end_time - timedelta(days=1)
        elif period == "weekly":
            start_time = end_time - timedelta(weeks=1)
        else:
            start_time = end_time - timedelta(hours=1)

        # Process different metrics
        analytics_report = {
            "timestamp": datetime.utcnow().isoformat(),
            "period": period,
            "start_time": start_time.isoformat(),
            "end_time": end_time.isoformat(),
            "metrics": {
                "commands": process_command_metrics(start_time, end_time),
                "tests": process_test_metrics(start_time, end_time),
                "users": process_user_metrics(start_time, end_time),
            },
        }

        # Store report in Firestore
        report_ref = db.collection("analytics_reports").document()
        report_ref.set(analytics_report)

        # Save to Cloud Storage
        bucket_name = os.environ.get("STORAGE_BUCKET")
        if bucket_name:
            bucket = storage_client.bucket(bucket_name)
            blob_name = f"artifacts/reports/{end_time.strftime('%Y/%m/%d')}/analytics_{period}_{end_time.strftime('%H%M%S')}.json"
            blob = bucket.blob(blob_name)
            blob.upload_from_string(
                json.dumps(analytics_report, indent=2, default=str),
                content_type="application/json",
            )
            logger.info(f"Analytics report saved to gs://{bucket_name}/{blob_name}")

        # Export to BigQuery if configured
        dataset_id = os.environ.get("BIGQUERY_DATASET")
        if dataset_id:
            try:
                bq_client = bigquery.Client()
                table_id = f"{os.environ['PROJECT_ID']}.{dataset_id}.analytics_reports"

                # Flatten the report for BigQuery
                bq_row = {
                    "timestamp": analytics_report["timestamp"],
                    "period": period,
                    "total_commands": analytics_report["metrics"]["commands"][
                        "total_commands"
                    ],
                    "successful_commands": analytics_report["metrics"]["commands"][
                        "successful_commands"
                    ],
                    "failed_commands": analytics_report["metrics"]["commands"][
                        "failed_commands"
                    ],
                    "total_tests": analytics_report["metrics"]["tests"]["total_tests"],
                    "passed_tests": analytics_report["metrics"]["tests"][
                        "passed_tests"
                    ],
                    "failed_tests": analytics_report["metrics"]["tests"][
                        "failed_tests"
                    ],
                    "active_users": analytics_report["metrics"]["users"][
                        "active_users"
                    ],
                    "total_sessions": analytics_report["metrics"]["users"][
                        "total_sessions"
                    ],
                }

                errors = bq_client.insert_rows_json(table_id, [bq_row])
                if errors:
                    logger.error(f"BigQuery insert errors: {errors}")
                else:
                    logger.info("Analytics exported to BigQuery")
            except Exception as e:
                logger.error(f"BigQuery export failed: {e}")

        logger.info(f"Analytics processing completed for period: {period}")

    except Exception as e:
        logger.error(f"Analytics processing failed: {e}", exc_info=True)
        raise
