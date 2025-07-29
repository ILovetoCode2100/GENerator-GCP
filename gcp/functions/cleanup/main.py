"""
Cloud Function for scheduled maintenance and cleanup tasks.
"""

import json
import asyncio
import os
from typing import Dict, Any, List
from datetime import datetime, timedelta
from google.cloud import firestore, storage, logging as cloud_logging
from flask import Request, Response
import logging

# Configure structured logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Set up Cloud Logging
cloud_logger = cloud_logging.Client().logger("cleanup-function")

# Environment variables
PROJECT_ID = os.environ.get("GCP_PROJECT", "virtuoso-api")
CLEANUP_BUCKET = os.environ.get("CLEANUP_BUCKET", "virtuoso-api-archives")
SESSION_TTL_HOURS = int(os.environ.get("SESSION_TTL_HOURS", "24"))
LOG_RETENTION_DAYS = int(os.environ.get("LOG_RETENTION_DAYS", "30"))
TEMP_FILE_TTL_HOURS = int(os.environ.get("TEMP_FILE_TTL_HOURS", "6"))


class CleanupService:
    """Performs various cleanup operations."""

    def __init__(self):
        self.db = firestore.AsyncClient()
        self.storage_client = storage.Client()
        self.stats = {
            "sessions_cleaned": 0,
            "logs_archived": 0,
            "temp_files_deleted": 0,
            "errors": [],
        }
        self.start_time = datetime.utcnow()

    async def clean_expired_sessions(self) -> Dict[str, Any]:
        """Clean up expired Firestore sessions."""
        try:
            logger.info("Starting session cleanup...")

            # Calculate expiration time
            expiry_time = datetime.utcnow() - timedelta(hours=SESSION_TTL_HOURS)

            # Query expired sessions
            sessions_ref = self.db.collection("sessions")
            query = sessions_ref.where("last_accessed", "<", expiry_time).limit(500)

            # Get and delete expired sessions
            batch = self.db.batch()
            batch_count = 0

            async for doc in query.stream():
                batch.delete(doc.reference)
                batch_count += 1
                self.stats["sessions_cleaned"] += 1

                # Commit batch every 100 documents
                if batch_count >= 100:
                    await batch.commit()
                    batch = self.db.batch()
                    batch_count = 0

            # Commit remaining batch
            if batch_count > 0:
                await batch.commit()

            logger.info(f"Cleaned {self.stats['sessions_cleaned']} expired sessions")

            return {
                "status": "success",
                "cleaned": self.stats["sessions_cleaned"],
                "threshold": f"{SESSION_TTL_HOURS} hours",
            }

        except Exception as e:
            error_msg = f"Session cleanup failed: {str(e)}"
            logger.error(error_msg)
            self.stats["errors"].append(error_msg)
            return {"status": "error", "error": str(e)}

    async def archive_old_logs(self) -> Dict[str, Any]:
        """Archive old logs to Cloud Storage."""
        try:
            logger.info("Starting log archival...")

            # Calculate cutoff date
            cutoff_date = datetime.utcnow() - timedelta(days=LOG_RETENTION_DAYS)

            # Query old logs
            logs_ref = self.db.collection("logs")
            query = logs_ref.where("timestamp", "<", cutoff_date).limit(1000)

            # Prepare archive
            archive_data = []
            batch = self.db.batch()
            batch_count = 0

            async for doc in query.stream():
                log_data = doc.to_dict()
                log_data["_id"] = doc.id
                archive_data.append(log_data)

                # Delete from Firestore
                batch.delete(doc.reference)
                batch_count += 1

                # Archive every 100 logs
                if len(archive_data) >= 100:
                    await self._save_archive(archive_data)
                    archive_data = []

                # Commit deletion batch
                if batch_count >= 100:
                    await batch.commit()
                    batch = self.db.batch()
                    batch_count = 0

            # Save remaining archive data
            if archive_data:
                await self._save_archive(archive_data)

            # Commit remaining deletions
            if batch_count > 0:
                await batch.commit()

            logger.info(f"Archived {self.stats['logs_archived']} logs")

            return {
                "status": "success",
                "archived": self.stats["logs_archived"],
                "retention_days": LOG_RETENTION_DAYS,
            }

        except Exception as e:
            error_msg = f"Log archival failed: {str(e)}"
            logger.error(error_msg)
            self.stats["errors"].append(error_msg)
            return {"status": "error", "error": str(e)}

    async def _save_archive(self, logs: List[Dict[str, Any]]) -> None:
        """Save logs to Cloud Storage archive."""
        if not logs:
            return

        # Create archive filename with timestamp
        timestamp = datetime.utcnow().strftime("%Y%m%d_%H%M%S")
        filename = f"logs/archive_{timestamp}_{len(logs)}.json"

        # Get bucket
        bucket = self.storage_client.bucket(CLEANUP_BUCKET)
        blob = bucket.blob(filename)

        # Upload as JSON
        blob.upload_from_string(
            json.dumps(logs, default=str, indent=2), content_type="application/json"
        )

        self.stats["logs_archived"] += len(logs)
        logger.info(f"Archived {len(logs)} logs to {filename}")

    async def delete_temp_files(self) -> Dict[str, Any]:
        """Delete temporary files from Cloud Storage."""
        try:
            logger.info("Starting temporary file cleanup...")

            # Get temp bucket
            bucket = self.storage_client.bucket(f"{PROJECT_ID}-temp")

            # Calculate expiry time
            expiry_time = datetime.utcnow() - timedelta(hours=TEMP_FILE_TTL_HOURS)

            # List and delete old files
            deleted_files = []

            for blob in bucket.list_blobs(prefix="temp/"):
                if blob.time_created < expiry_time.replace(
                    tzinfo=blob.time_created.tzinfo
                ):
                    blob.delete()
                    deleted_files.append(blob.name)
                    self.stats["temp_files_deleted"] += 1

                    # Log every 100 files
                    if len(deleted_files) % 100 == 0:
                        logger.info(f"Deleted {len(deleted_files)} temp files...")

            logger.info(f"Deleted {self.stats['temp_files_deleted']} temporary files")

            return {
                "status": "success",
                "deleted": self.stats["temp_files_deleted"],
                "ttl_hours": TEMP_FILE_TTL_HOURS,
                "sample_files": deleted_files[:10],  # Show first 10 as sample
            }

        except Exception as e:
            error_msg = f"Temp file cleanup failed: {str(e)}"
            logger.error(error_msg)
            self.stats["errors"].append(error_msg)
            return {"status": "error", "error": str(e)}

    async def generate_cleanup_report(self) -> Dict[str, Any]:
        """Generate a comprehensive cleanup report."""
        report = {
            "timestamp": self.start_time.isoformat(),
            "duration_seconds": int(
                (datetime.utcnow() - self.start_time).total_seconds()
            ),
            "summary": {
                "sessions_cleaned": self.stats["sessions_cleaned"],
                "logs_archived": self.stats["logs_archived"],
                "temp_files_deleted": self.stats["temp_files_deleted"],
                "total_operations": sum(
                    [
                        self.stats["sessions_cleaned"],
                        self.stats["logs_archived"],
                        self.stats["temp_files_deleted"],
                    ]
                ),
                "errors_count": len(self.stats["errors"]),
            },
            "errors": self.stats["errors"],
            "status": "success" if not self.stats["errors"] else "partial_success",
        }

        # Save report to Firestore
        await self.db.collection("cleanup_reports").add(
            {**report, "created_at": firestore.SERVER_TIMESTAMP}
        )

        # Log to Cloud Logging
        cloud_logger.log_struct(report, severity="INFO")

        return report

    async def run_all_cleanup_tasks(self) -> Dict[str, Any]:
        """Run all cleanup tasks."""
        results = {}

        # Run cleanup tasks
        tasks = [
            ("sessions", self.clean_expired_sessions()),
            ("logs", self.archive_old_logs()),
            ("temp_files", self.delete_temp_files()),
        ]

        for task_name, task in tasks:
            try:
                result = await task
                results[task_name] = result
            except Exception as e:
                logger.error(f"Task {task_name} failed: {str(e)}")
                results[task_name] = {"status": "error", "error": str(e)}

        # Generate report
        report = await self.generate_cleanup_report()

        return {"report": report, "details": results}


def cleanup(request: Request) -> Response:
    """
    Cloud Function entry point for cleanup operations.

    Args:
        request: The Flask request object

    Returns:
        Response with cleanup results
    """
    # Handle CORS
    if request.method == "OPTIONS":
        headers = {
            "Access-Control-Allow-Origin": "*",
            "Access-Control-Allow-Methods": "POST, GET",
            "Access-Control-Allow-Headers": "Content-Type",
            "Access-Control-Max-Age": "3600",
        }
        return Response("", 204, headers)

    headers = {"Access-Control-Allow-Origin": "*", "Content-Type": "application/json"}

    try:
        # Check for specific task
        task = request.args.get("task")

        # Run cleanup
        service = CleanupService()
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)

        try:
            if task == "sessions":
                result = loop.run_until_complete(service.clean_expired_sessions())
            elif task == "logs":
                result = loop.run_until_complete(service.archive_old_logs())
            elif task == "temp_files":
                result = loop.run_until_complete(service.delete_temp_files())
            else:
                # Run all tasks
                result = loop.run_until_complete(service.run_all_cleanup_tasks())

            return Response(json.dumps(result, indent=2), 200, headers)

        finally:
            loop.close()

    except Exception as e:
        logger.error(f"Cleanup function failed: {str(e)}")
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
        return cleanup(flask_request)

    app.run(debug=True, port=8080)
