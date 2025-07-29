"""
Webhook handler function for Virtuoso API CLI
Processes incoming webhooks and publishes events to Pub/Sub
"""

import json
import logging
import os
import hmac
import hashlib
from datetime import datetime

import functions_framework
from google.cloud import pubsub_v1
from google.cloud import firestore

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Initialize clients
publisher = pubsub_v1.PublisherClient()
db = firestore.Client()


def verify_webhook_signature(request_body: bytes, signature: str, secret: str) -> bool:
    """
    Verify webhook signature using HMAC-SHA256.

    Args:
        request_body: Raw request body
        signature: Signature from webhook header
        secret: Webhook secret

    Returns:
        True if signature is valid
    """
    expected_signature = hmac.new(
        secret.encode(), request_body, hashlib.sha256
    ).hexdigest()

    return hmac.compare_digest(signature, expected_signature)


@functions_framework.http
def webhook_handler(request):
    """
    Handle incoming webhooks from Virtuoso.

    Validates the webhook, processes the data, and publishes to Pub/Sub.
    """
    try:
        # Get webhook secret
        webhook_secret = os.environ.get("WEBHOOK_SECRET", "")

        # Verify signature if secret is configured
        if webhook_secret:
            signature = request.headers.get("X-Webhook-Signature", "")
            if not signature:
                logger.warning("Missing webhook signature")
                return ({"error": "Missing signature"}, 401)

            if not verify_webhook_signature(
                request.get_data(), signature, webhook_secret
            ):
                logger.warning("Invalid webhook signature")
                return ({"error": "Invalid signature"}, 401)

        # Parse webhook data
        try:
            webhook_data = request.get_json()
            if not webhook_data:
                return ({"error": "Invalid JSON"}, 400)
        except Exception as e:
            logger.error(f"Failed to parse webhook JSON: {e}")
            return ({"error": "Invalid JSON"}, 400)

        # Extract webhook metadata
        webhook_id = webhook_data.get("id", "unknown")
        event_type = webhook_data.get("event", "unknown")
        timestamp = webhook_data.get("timestamp", datetime.utcnow().isoformat())

        # Store webhook in Firestore for processing
        webhook_ref = db.collection("webhooks").document(webhook_id)
        webhook_ref.set(
            {
                "id": webhook_id,
                "event_type": event_type,
                "timestamp": timestamp,
                "data": webhook_data,
                "processed": False,
                "received_at": firestore.SERVER_TIMESTAMP,
                "source_ip": request.remote_addr,
            }
        )

        # Prepare event for Pub/Sub
        event = {
            "eventId": webhook_id,
            "eventType": f"webhook.{event_type}",
            "timestamp": timestamp,
            "data": webhook_data,
            "metadata": {
                "source": "webhook_handler",
                "version": "1.0.0",
                "webhook_id": webhook_id,
            },
        }

        # Determine which topic to publish to
        project_id = os.environ.get("PROJECT_ID")
        if event_type.startswith("command."):
            topic_name = os.environ.get(
                "PUBSUB_TOPIC_COMMANDS", "virtuoso-command-events"
            )
        elif event_type.startswith("test."):
            topic_name = os.environ.get("PUBSUB_TOPIC_TESTS", "virtuoso-test-events")
        else:
            topic_name = os.environ.get("PUBSUB_TOPIC_EVENTS", "virtuoso-system-events")

        topic_path = publisher.topic_path(project_id, topic_name)

        # Publish to Pub/Sub
        message_data = json.dumps(event).encode("utf-8")
        future = publisher.publish(
            topic_path, message_data, event_type=event_type, webhook_id=webhook_id
        )

        # Wait for publish to complete
        message_id = future.result()

        # Update webhook as processed
        webhook_ref.update(
            {
                "processed": True,
                "processed_at": firestore.SERVER_TIMESTAMP,
                "pubsub_message_id": message_id,
            }
        )

        logger.info(f"Webhook {webhook_id} processed and published to {topic_name}")

        return (
            {"status": "accepted", "webhook_id": webhook_id, "message_id": message_id},
            202,
        )

    except Exception as e:
        logger.error(f"Webhook processing failed: {e}", exc_info=True)
        return (
            {
                "error": "Internal server error",
                "message": str(e) if os.environ.get("DEBUG") else "Processing failed",
            },
            500,
        )
