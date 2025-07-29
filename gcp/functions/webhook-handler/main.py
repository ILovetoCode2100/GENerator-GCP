"""
Cloud Function for processing external webhooks from GitHub and Virtuoso.
"""

import json
import hmac
import hashlib
import os
from typing import Dict, Any
from datetime import datetime
import asyncio
from google.cloud import pubsub_v1
from flask import Request, Response
import logging

# Configure structured logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Environment variables
PROJECT_ID = os.environ.get("GCP_PROJECT", "virtuoso-api")
GITHUB_SECRET = os.environ.get("GITHUB_WEBHOOK_SECRET", "")
VIRTUOSO_SECRET = os.environ.get("VIRTUOSO_WEBHOOK_SECRET", "")
PUBSUB_TOPIC_GITHUB = os.environ.get("PUBSUB_TOPIC_GITHUB", "github-webhooks")
PUBSUB_TOPIC_VIRTUOSO = os.environ.get("PUBSUB_TOPIC_VIRTUOSO", "virtuoso-webhooks")

# Initialize Pub/Sub publisher
publisher = pubsub_v1.PublisherClient()


class WebhookProcessor:
    """Processes and validates webhooks."""

    @staticmethod
    def verify_github_signature(payload: bytes, signature: str, secret: str) -> bool:
        """Verify GitHub webhook signature."""
        if not signature or not secret:
            return False

        try:
            # GitHub sends sha256 signatures
            expected_signature = (
                "sha256="
                + hmac.new(secret.encode("utf-8"), payload, hashlib.sha256).hexdigest()
            )

            return hmac.compare_digest(expected_signature, signature)
        except Exception as e:
            logger.error(f"Error verifying GitHub signature: {str(e)}")
            return False

    @staticmethod
    def verify_virtuoso_signature(payload: bytes, signature: str, secret: str) -> bool:
        """Verify Virtuoso webhook signature."""
        if not signature or not secret:
            return False

        try:
            # Virtuoso uses sha1 signatures
            expected_signature = hmac.new(
                secret.encode("utf-8"), payload, hashlib.sha1
            ).hexdigest()

            return hmac.compare_digest(expected_signature, signature)
        except Exception as e:
            logger.error(f"Error verifying Virtuoso signature: {str(e)}")
            return False

    async def process_github_webhook(
        self, event_type: str, payload: Dict[str, Any]
    ) -> Dict[str, Any]:
        """Process GitHub webhook events."""
        processed_data = {
            "source": "github",
            "event_type": event_type,
            "timestamp": datetime.utcnow().isoformat(),
            "payload": payload,
        }

        # Extract relevant information based on event type
        if event_type == "push":
            processed_data["summary"] = {
                "repository": payload.get("repository", {}).get("full_name"),
                "ref": payload.get("ref"),
                "commits": len(payload.get("commits", [])),
                "pusher": payload.get("pusher", {}).get("name"),
            }
        elif event_type == "pull_request":
            pr = payload.get("pull_request", {})
            processed_data["summary"] = {
                "action": payload.get("action"),
                "number": pr.get("number"),
                "title": pr.get("title"),
                "user": pr.get("user", {}).get("login"),
                "base": pr.get("base", {}).get("ref"),
                "head": pr.get("head", {}).get("ref"),
            }
        elif event_type == "issues":
            issue = payload.get("issue", {})
            processed_data["summary"] = {
                "action": payload.get("action"),
                "number": issue.get("number"),
                "title": issue.get("title"),
                "user": issue.get("user", {}).get("login"),
                "state": issue.get("state"),
            }
        elif event_type == "workflow_run":
            workflow = payload.get("workflow_run", {})
            processed_data["summary"] = {
                "action": payload.get("action"),
                "name": workflow.get("name"),
                "status": workflow.get("status"),
                "conclusion": workflow.get("conclusion"),
                "run_number": workflow.get("run_number"),
            }

        return processed_data

    async def process_virtuoso_webhook(self, payload: Dict[str, Any]) -> Dict[str, Any]:
        """Process Virtuoso webhook events."""
        processed_data = {
            "source": "virtuoso",
            "event_type": payload.get("event_type", "unknown"),
            "timestamp": datetime.utcnow().isoformat(),
            "payload": payload,
        }

        # Extract relevant information based on Virtuoso event types
        event_type = payload.get("event_type")

        if event_type == "test_execution_completed":
            execution = payload.get("execution", {})
            processed_data["summary"] = {
                "execution_id": execution.get("id"),
                "status": execution.get("status"),
                "duration_ms": execution.get("duration"),
                "passed_steps": execution.get("passed_steps"),
                "failed_steps": execution.get("failed_steps"),
                "project_id": execution.get("project_id"),
            }
        elif event_type == "test_failure":
            processed_data["summary"] = {
                "test_id": payload.get("test_id"),
                "checkpoint_id": payload.get("checkpoint_id"),
                "error_message": payload.get("error_message"),
                "step_number": payload.get("step_number"),
            }
        elif event_type == "goal_completed":
            processed_data["summary"] = {
                "goal_id": payload.get("goal_id"),
                "project_id": payload.get("project_id"),
                "status": payload.get("status"),
                "completed_at": payload.get("completed_at"),
            }

        return processed_data

    async def publish_to_pubsub(self, topic_name: str, data: Dict[str, Any]) -> str:
        """Publish processed webhook data to Pub/Sub."""
        topic_path = publisher.topic_path(PROJECT_ID, topic_name)

        # Convert data to JSON bytes
        message_data = json.dumps(data).encode("utf-8")

        # Add attributes for filtering
        attributes = {
            "source": data.get("source", "unknown"),
            "event_type": data.get("event_type", "unknown"),
            "timestamp": data.get("timestamp", ""),
        }

        # Publish message
        future = publisher.publish(topic_path, message_data, **attributes)

        # Wait for publish to complete
        message_id = future.result()
        logger.info(f"Published message {message_id} to {topic_name}")

        return message_id


def webhook_handler(request: Request) -> Response:
    """
    Cloud Function entry point for webhook processing.

    Args:
        request: The Flask request object

    Returns:
        Response with processing results
    """
    # Handle CORS
    if request.method == "OPTIONS":
        headers = {
            "Access-Control-Allow-Origin": "*",
            "Access-Control-Allow-Methods": "POST",
            "Access-Control-Allow-Headers": "Content-Type, X-Hub-Signature-256, X-Virtuoso-Signature",
            "Access-Control-Max-Age": "3600",
        }
        return Response("", 204, headers)

    headers = {"Content-Type": "application/json"}

    try:
        # Get request data
        request_data = request.get_data()

        # Initialize processor
        processor = WebhookProcessor()

        # Check for GitHub webhook
        github_signature = request.headers.get("X-Hub-Signature-256")
        if github_signature:
            # Verify GitHub signature
            if not processor.verify_github_signature(
                request_data, github_signature, GITHUB_SECRET
            ):
                logger.warning("Invalid GitHub webhook signature")
                return Response(
                    json.dumps({"error": "Invalid signature"}), 401, headers
                )

            # Get event type
            event_type = request.headers.get("X-GitHub-Event", "unknown")

            # Process webhook
            loop = asyncio.new_event_loop()
            asyncio.set_event_loop(loop)

            try:
                payload = request.get_json()
                processed_data = loop.run_until_complete(
                    processor.process_github_webhook(event_type, payload)
                )

                # Publish to Pub/Sub
                message_id = loop.run_until_complete(
                    processor.publish_to_pubsub(PUBSUB_TOPIC_GITHUB, processed_data)
                )

                response_data = {
                    "status": "success",
                    "source": "github",
                    "event_type": event_type,
                    "message_id": message_id,
                    "timestamp": datetime.utcnow().isoformat(),
                }

                return Response(json.dumps(response_data), 200, headers)

            finally:
                loop.close()

        # Check for Virtuoso webhook
        virtuoso_signature = request.headers.get("X-Virtuoso-Signature")
        if virtuoso_signature:
            # Verify Virtuoso signature
            if not processor.verify_virtuoso_signature(
                request_data, virtuoso_signature, VIRTUOSO_SECRET
            ):
                logger.warning("Invalid Virtuoso webhook signature")
                return Response(
                    json.dumps({"error": "Invalid signature"}), 401, headers
                )

            # Process webhook
            loop = asyncio.new_event_loop()
            asyncio.set_event_loop(loop)

            try:
                payload = request.get_json()
                processed_data = loop.run_until_complete(
                    processor.process_virtuoso_webhook(payload)
                )

                # Publish to Pub/Sub
                message_id = loop.run_until_complete(
                    processor.publish_to_pubsub(PUBSUB_TOPIC_VIRTUOSO, processed_data)
                )

                response_data = {
                    "status": "success",
                    "source": "virtuoso",
                    "event_type": processed_data.get("event_type"),
                    "message_id": message_id,
                    "timestamp": datetime.utcnow().isoformat(),
                }

                return Response(json.dumps(response_data), 200, headers)

            finally:
                loop.close()

        # No recognized webhook source
        logger.warning("Webhook received with no recognized signature")
        return Response(
            json.dumps(
                {
                    "error": "No valid webhook signature found",
                    "hint": "Expected X-Hub-Signature-256 or X-Virtuoso-Signature header",
                }
            ),
            400,
            headers,
        )

    except Exception as e:
        logger.error(f"Webhook processing failed: {str(e)}")
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

    @app.route("/", methods=["POST", "OPTIONS"])
    def main():
        return webhook_handler(flask_request)

    app.run(debug=True, port=8080)
