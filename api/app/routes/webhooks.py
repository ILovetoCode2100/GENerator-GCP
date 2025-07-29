"""
Webhook endpoints for event subscriptions and notifications.

These endpoints handle webhook management, Pub/Sub event subscriptions,
and async notification delivery.
"""

from typing import Dict, Any, List, Optional
from datetime import datetime, timezone
from uuid import uuid4
import hmac
import hashlib
import json

from fastapi import APIRouter, HTTPException, status, Depends, Request, BackgroundTasks
from pydantic import BaseModel, Field, HttpUrl

from ..config import settings
from ..utils.logger import get_logger
from ..services.auth_service import AuthUser, Permission
from ..middleware.auth import get_authenticated_user, require_permissions
from ..middleware.rate_limit import rate_limit, RateLimitStrategy
from ..models.responses import BaseResponse, ResponseStatus

# GCP imports
if settings.is_gcp_enabled:
    from ..gcp.pubsub_client import PubSubClient
    from ..gcp.firestore_client import FirestoreClient
    from ..gcp.cloud_tasks_client import CloudTasksClient

router = APIRouter()
logger = get_logger(__name__)

# Initialize GCP clients if enabled
pubsub_client = None
firestore_client = None
tasks_client = None

if settings.is_gcp_enabled:
    if settings.USE_PUBSUB:
        pubsub_client = PubSubClient()
    if settings.USE_FIRESTORE:
        firestore_client = FirestoreClient()
    if settings.USE_CLOUD_TASKS:
        tasks_client = CloudTasksClient()


class WebhookEventType(BaseModel):
    """Available webhook event types."""

    COMMAND_EXECUTED = "command.executed"
    COMMAND_FAILED = "command.failed"
    BATCH_COMPLETED = "batch.completed"
    TEST_STARTED = "test.started"
    TEST_COMPLETED = "test.completed"
    TEST_FAILED = "test.failed"
    SESSION_CREATED = "session.created"
    SESSION_ACTIVATED = "session.activated"
    SESSION_DELETED = "session.deleted"
    EXECUTION_STARTED = "execution.started"
    EXECUTION_COMPLETED = "execution.completed"
    EXECUTION_FAILED = "execution.failed"


class WebhookCreate(BaseModel):
    """Request model for creating a webhook."""

    name: str = Field(..., description="Webhook name")
    url: HttpUrl = Field(..., description="Webhook endpoint URL")
    events: List[str] = Field(..., description="List of event types to subscribe to")
    secret: Optional[str] = Field(
        None, description="Secret for HMAC signature verification"
    )
    headers: Optional[Dict[str, str]] = Field(
        None, description="Custom headers to include"
    )
    active: bool = Field(True, description="Whether webhook is active")


class WebhookUpdate(BaseModel):
    """Request model for updating a webhook."""

    name: Optional[str] = Field(None, description="Webhook name")
    url: Optional[HttpUrl] = Field(None, description="Webhook endpoint URL")
    events: Optional[List[str]] = Field(None, description="List of event types")
    secret: Optional[str] = Field(None, description="Secret for HMAC signature")
    headers: Optional[Dict[str, str]] = Field(None, description="Custom headers")
    active: Optional[bool] = Field(None, description="Whether webhook is active")


class Webhook(BaseModel):
    """Webhook model."""

    webhook_id: str = Field(..., description="Webhook ID")
    name: str = Field(..., description="Webhook name")
    url: str = Field(..., description="Webhook endpoint URL")
    events: List[str] = Field(..., description="Subscribed event types")
    secret: Optional[str] = Field(None, description="Secret for HMAC signature")
    headers: Optional[Dict[str, str]] = Field(None, description="Custom headers")
    active: bool = Field(..., description="Whether webhook is active")
    created_at: datetime = Field(..., description="Creation timestamp")
    updated_at: datetime = Field(..., description="Last update timestamp")
    last_triggered_at: Optional[datetime] = Field(
        None, description="Last trigger timestamp"
    )
    failure_count: int = Field(0, description="Consecutive failure count")


class WebhookDelivery(BaseModel):
    """Webhook delivery attempt model."""

    delivery_id: str = Field(..., description="Delivery ID")
    webhook_id: str = Field(..., description="Webhook ID")
    event_type: str = Field(..., description="Event type")
    event_id: str = Field(..., description="Event ID")
    payload: Dict[str, Any] = Field(..., description="Event payload")
    status: str = Field(..., description="Delivery status")
    attempts: int = Field(..., description="Number of attempts")
    response_status: Optional[int] = Field(None, description="HTTP response status")
    response_body: Optional[str] = Field(None, description="Response body")
    error: Optional[str] = Field(None, description="Error message if failed")
    created_at: datetime = Field(..., description="Creation timestamp")
    delivered_at: Optional[datetime] = Field(
        None, description="Successful delivery timestamp"
    )


@router.post(
    "/",
    response_model=Webhook,
    dependencies=[Depends(rate_limit(10, 60, RateLimitStrategy.PER_USER))],
)
async def create_webhook(
    request: WebhookCreate,
    user: AuthUser = Depends(get_authenticated_user),
    background_tasks: BackgroundTasks = BackgroundTasks(),
) -> Webhook:
    """
    Create a new webhook subscription.

    Args:
        request: Webhook creation request
        user: Authenticated user
        background_tasks: Background tasks

    Returns:
        Created webhook
    """
    webhook_id = f"wh_{uuid4().hex[:12]}"
    now = datetime.now(timezone.utc)

    # Generate secret if not provided
    if not request.secret:
        request.secret = f"whsec_{uuid4().hex}"

    webhook = Webhook(
        webhook_id=webhook_id,
        name=request.name,
        url=str(request.url),
        events=request.events,
        secret=request.secret,
        headers=request.headers,
        active=request.active,
        created_at=now,
        updated_at=now,
        failure_count=0,
    )

    # Store in Firestore if available
    if settings.USE_FIRESTORE and firestore_client:
        try:
            await firestore_client.create_webhook(
                webhook_id=webhook_id,
                user_id=user.user_id,
                tenant_id=user.tenant_id,
                webhook_data=webhook.dict(),
            )

            # Subscribe to Pub/Sub topics for events
            if settings.USE_PUBSUB and pubsub_client:
                for event_type in request.events:
                    topic_name = get_topic_for_event(event_type)
                    if topic_name:
                        background_tasks.add_task(
                            subscribe_to_topic,
                            webhook_id=webhook_id,
                            topic_name=topic_name,
                            event_type=event_type,
                        )

        except Exception as e:
            logger.error(f"Failed to create webhook: {e}")
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail=f"Failed to create webhook: {str(e)}",
            )

    return webhook


@router.get(
    "/",
    response_model=List[Webhook],
    dependencies=[Depends(require_permissions(Permission.READ_WEBHOOKS))],
)
async def list_webhooks(
    active_only: bool = True,
    limit: int = 100,
    offset: int = 0,
    user: AuthUser = Depends(get_authenticated_user),
) -> List[Webhook]:
    """
    List all webhooks for the user.

    Args:
        active_only: Filter only active webhooks
        limit: Maximum results
        offset: Pagination offset
        user: Authenticated user

    Returns:
        List of webhooks
    """
    webhooks = []

    if settings.USE_FIRESTORE and firestore_client:
        try:
            webhook_data_list = await firestore_client.list_user_webhooks(
                user_id=user.user_id,
                active_only=active_only,
                limit=limit,
                offset=offset,
            )

            for data in webhook_data_list:
                webhooks.append(Webhook(**data))

        except Exception as e:
            logger.error(f"Failed to list webhooks: {e}")

    return webhooks


@router.get(
    "/{webhook_id}",
    response_model=Webhook,
    dependencies=[Depends(require_permissions(Permission.READ_WEBHOOKS))],
)
async def get_webhook(
    webhook_id: str, user: AuthUser = Depends(get_authenticated_user)
) -> Webhook:
    """
    Get a specific webhook.

    Args:
        webhook_id: Webhook ID
        user: Authenticated user

    Returns:
        Webhook details
    """
    if settings.USE_FIRESTORE and firestore_client:
        try:
            webhook_data = await firestore_client.get_webhook(webhook_id)

            if webhook_data and webhook_data.get("user_id") == user.user_id:
                return Webhook(**webhook_data)

        except Exception as e:
            logger.error(f"Failed to get webhook: {e}")

    raise HTTPException(
        status_code=status.HTTP_404_NOT_FOUND, detail=f"Webhook not found: {webhook_id}"
    )


@router.patch(
    "/{webhook_id}",
    response_model=Webhook,
    dependencies=[Depends(rate_limit(20, 60, RateLimitStrategy.PER_USER))],
)
async def update_webhook(
    webhook_id: str,
    request: WebhookUpdate,
    user: AuthUser = Depends(get_authenticated_user),
    background_tasks: BackgroundTasks = BackgroundTasks(),
) -> Webhook:
    """
    Update a webhook.

    Args:
        webhook_id: Webhook ID
        request: Update request
        user: Authenticated user
        background_tasks: Background tasks

    Returns:
        Updated webhook
    """
    if settings.USE_FIRESTORE and firestore_client:
        try:
            # Get existing webhook
            webhook_data = await firestore_client.get_webhook(webhook_id)

            if not webhook_data or webhook_data.get("user_id") != user.user_id:
                raise HTTPException(
                    status_code=status.HTTP_403_FORBIDDEN,
                    detail="Access denied to this webhook",
                )

            # Update fields
            update_data = request.dict(exclude_unset=True)
            update_data["updated_at"] = datetime.now(timezone.utc)

            # Update webhook
            updated_webhook = await firestore_client.update_webhook(
                webhook_id=webhook_id, update_data=update_data
            )

            if updated_webhook:
                # Update Pub/Sub subscriptions if events changed
                if request.events is not None and settings.USE_PUBSUB and pubsub_client:
                    old_events = set(webhook_data.get("events", []))
                    new_events = set(request.events)

                    # Unsubscribe from removed events
                    for event_type in old_events - new_events:
                        topic_name = get_topic_for_event(event_type)
                        if topic_name:
                            background_tasks.add_task(
                                unsubscribe_from_topic,
                                webhook_id=webhook_id,
                                topic_name=topic_name,
                                event_type=event_type,
                            )

                    # Subscribe to new events
                    for event_type in new_events - old_events:
                        topic_name = get_topic_for_event(event_type)
                        if topic_name:
                            background_tasks.add_task(
                                subscribe_to_topic,
                                webhook_id=webhook_id,
                                topic_name=topic_name,
                                event_type=event_type,
                            )

                return Webhook(**updated_webhook)

        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Failed to update webhook: {e}")

    raise HTTPException(
        status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
        detail="Failed to update webhook",
    )


@router.delete(
    "/{webhook_id}",
    status_code=status.HTTP_204_NO_CONTENT,
    dependencies=[Depends(rate_limit(10, 60, RateLimitStrategy.PER_USER))],
)
async def delete_webhook(
    webhook_id: str,
    user: AuthUser = Depends(get_authenticated_user),
    background_tasks: BackgroundTasks = BackgroundTasks(),
) -> None:
    """
    Delete a webhook.

    Args:
        webhook_id: Webhook ID
        user: Authenticated user
        background_tasks: Background tasks
    """
    if settings.USE_FIRESTORE and firestore_client:
        try:
            # Get webhook to verify ownership
            webhook_data = await firestore_client.get_webhook(webhook_id)

            if not webhook_data or webhook_data.get("user_id") != user.user_id:
                raise HTTPException(
                    status_code=status.HTTP_403_FORBIDDEN,
                    detail="Access denied to this webhook",
                )

            # Delete webhook
            success = await firestore_client.delete_webhook(webhook_id)

            if success:
                # Unsubscribe from all events
                if settings.USE_PUBSUB and pubsub_client:
                    for event_type in webhook_data.get("events", []):
                        topic_name = get_topic_for_event(event_type)
                        if topic_name:
                            background_tasks.add_task(
                                unsubscribe_from_topic,
                                webhook_id=webhook_id,
                                topic_name=topic_name,
                                event_type=event_type,
                            )
                return

        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Failed to delete webhook: {e}")

    raise HTTPException(
        status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
        detail="Failed to delete webhook",
    )


@router.post(
    "/{webhook_id}/test",
    response_model=BaseResponse[Dict[str, Any]],
    dependencies=[Depends(rate_limit(5, 60, RateLimitStrategy.PER_USER))],
)
async def test_webhook(
    webhook_id: str, user: AuthUser = Depends(get_authenticated_user)
) -> BaseResponse[Dict[str, Any]]:
    """
    Test a webhook with a sample payload.

    Args:
        webhook_id: Webhook ID
        user: Authenticated user

    Returns:
        Test result
    """
    if settings.USE_FIRESTORE and firestore_client:
        try:
            # Get webhook
            webhook_data = await firestore_client.get_webhook(webhook_id)

            if not webhook_data or webhook_data.get("user_id") != user.user_id:
                raise HTTPException(
                    status_code=status.HTTP_403_FORBIDDEN,
                    detail="Access denied to this webhook",
                )

            # Create test payload
            test_payload = {
                "event_id": f"test_{uuid4().hex[:8]}",
                "event_type": "webhook.test",
                "timestamp": datetime.now(timezone.utc).isoformat(),
                "data": {
                    "message": "This is a test webhook delivery",
                    "webhook_id": webhook_id,
                    "user_id": user.user_id,
                },
            }

            # Deliver webhook
            if settings.USE_CLOUD_TASKS and tasks_client:
                delivery_id = await deliver_webhook_async(
                    webhook_data=webhook_data, payload=test_payload
                )

                return BaseResponse(
                    status=ResponseStatus.SUCCESS,
                    data={
                        "delivery_id": delivery_id,
                        "status": "queued",
                        "message": "Test webhook queued for delivery",
                    },
                    message="Webhook test initiated",
                )
            else:
                # Synchronous delivery for testing
                result = await deliver_webhook_sync(
                    webhook_data=webhook_data, payload=test_payload
                )

                return BaseResponse(
                    status=ResponseStatus.SUCCESS,
                    data=result,
                    message="Webhook test completed",
                )

        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Failed to test webhook: {e}")

    raise HTTPException(
        status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
        detail="Failed to test webhook",
    )


@router.get(
    "/{webhook_id}/deliveries",
    response_model=BaseResponse[List[WebhookDelivery]],
    dependencies=[Depends(require_permissions(Permission.READ_WEBHOOKS))],
)
async def get_webhook_deliveries(
    webhook_id: str,
    limit: int = 50,
    offset: int = 0,
    user: AuthUser = Depends(get_authenticated_user),
) -> BaseResponse[List[WebhookDelivery]]:
    """
    Get delivery history for a webhook.

    Args:
        webhook_id: Webhook ID
        limit: Maximum results
        offset: Pagination offset
        user: Authenticated user

    Returns:
        Webhook delivery history
    """
    deliveries = []

    if settings.USE_FIRESTORE and firestore_client:
        try:
            # Verify webhook ownership
            webhook_data = await firestore_client.get_webhook(webhook_id)

            if not webhook_data or webhook_data.get("user_id") != user.user_id:
                raise HTTPException(
                    status_code=status.HTTP_403_FORBIDDEN,
                    detail="Access denied to this webhook",
                )

            # Get delivery history
            delivery_data_list = await firestore_client.get_webhook_deliveries(
                webhook_id=webhook_id, limit=limit, offset=offset
            )

            for data in delivery_data_list:
                deliveries.append(WebhookDelivery(**data))

        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Failed to get webhook deliveries: {e}")

    return BaseResponse(
        status=ResponseStatus.SUCCESS,
        data=deliveries,
        message=f"Found {len(deliveries)} deliveries",
    )


@router.post(
    "/pubsub/receive",
    response_model=BaseResponse[Dict[str, str]],
    include_in_schema=False,  # Internal endpoint
)
async def receive_pubsub_message(
    request: Request, background_tasks: BackgroundTasks = BackgroundTasks()
) -> BaseResponse[Dict[str, str]]:
    """
    Receive Pub/Sub push messages for webhook delivery.

    This is an internal endpoint called by Pub/Sub push subscriptions.

    Args:
        request: HTTP request containing Pub/Sub message
        background_tasks: Background tasks

    Returns:
        Acknowledgment
    """
    try:
        # Verify Pub/Sub push token
        token = request.headers.get("X-Pubsub-Token")
        if token != settings.PUBSUB_PUSH_TOKEN:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid Pub/Sub token"
            )

        # Parse Pub/Sub message
        body = await request.json()
        message = body.get("message", {})

        if not message:
            return BaseResponse(
                status=ResponseStatus.SUCCESS,
                data={"status": "ignored"},
                message="No message to process",
            )

        # Decode message data
        import base64

        data = json.loads(base64.b64decode(message["data"]).decode())

        # Get event details
        event_type = data.get("event_type")
        event_data = data.get("data", {})

        # Find webhooks subscribed to this event
        if settings.USE_FIRESTORE and firestore_client:
            webhooks = await firestore_client.get_webhooks_for_event(event_type)

            # Queue webhook deliveries
            for webhook_data in webhooks:
                if webhook_data.get("active"):
                    background_tasks.add_task(
                        process_webhook_delivery,
                        webhook_data=webhook_data,
                        event_type=event_type,
                        event_data=event_data,
                        event_id=data.get("event_id", str(uuid4())),
                    )

        return BaseResponse(
            status=ResponseStatus.SUCCESS,
            data={"status": "accepted"},
            message="Message processed",
        )

    except Exception as e:
        logger.error(f"Failed to process Pub/Sub message: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to process message",
        )


# Helper functions
def get_topic_for_event(event_type: str) -> Optional[str]:
    """Get Pub/Sub topic name for an event type."""
    topic_map = {
        "command.": "command-events",
        "batch.": "command-events",
        "test.": "test-events",
        "session.": "session-events",
        "execution.": "execution-events",
    }

    for prefix, topic in topic_map.items():
        if event_type.startswith(prefix):
            return topic

    return None


async def subscribe_to_topic(webhook_id: str, topic_name: str, event_type: str):
    """Subscribe webhook to a Pub/Sub topic."""
    if settings.USE_PUBSUB and pubsub_client:
        try:
            # This would create a subscription or update subscription filters
            logger.info(
                f"Subscribed webhook {webhook_id} to {topic_name} for {event_type}"
            )
        except Exception as e:
            logger.error(f"Failed to subscribe webhook: {e}")


async def unsubscribe_from_topic(webhook_id: str, topic_name: str, event_type: str):
    """Unsubscribe webhook from a Pub/Sub topic."""
    if settings.USE_PUBSUB and pubsub_client:
        try:
            # This would remove subscription or update filters
            logger.info(
                f"Unsubscribed webhook {webhook_id} from {topic_name} for {event_type}"
            )
        except Exception as e:
            logger.error(f"Failed to unsubscribe webhook: {e}")


async def deliver_webhook_async(
    webhook_data: Dict[str, Any], payload: Dict[str, Any]
) -> str:
    """Queue webhook delivery using Cloud Tasks."""
    if not settings.USE_CLOUD_TASKS or not tasks_client:
        raise ValueError("Cloud Tasks not enabled")

    delivery_id = f"del_{uuid4().hex[:12]}"

    # Create delivery task
    await tasks_client.create_webhook_delivery_task(
        delivery_id=delivery_id,
        webhook_id=webhook_data["webhook_id"],
        webhook_url=webhook_data["url"],
        payload=payload,
        secret=webhook_data.get("secret"),
        headers=webhook_data.get("headers", {}),
    )

    return delivery_id


async def deliver_webhook_sync(
    webhook_data: Dict[str, Any], payload: Dict[str, Any]
) -> Dict[str, Any]:
    """Deliver webhook synchronously (for testing)."""
    import aiohttp

    delivery_id = f"del_{uuid4().hex[:12]}"
    url = webhook_data["url"]
    secret = webhook_data.get("secret")
    custom_headers = webhook_data.get("headers", {})

    # Prepare headers
    headers = {
        "Content-Type": "application/json",
        "X-Webhook-ID": webhook_data["webhook_id"],
        "X-Delivery-ID": delivery_id,
        **custom_headers,
    }

    # Add HMAC signature if secret provided
    if secret:
        signature = hmac.new(
            secret.encode(), json.dumps(payload).encode(), hashlib.sha256
        ).hexdigest()
        headers["X-Webhook-Signature"] = f"sha256={signature}"

    # Deliver webhook
    try:
        async with aiohttp.ClientSession() as session:
            async with session.post(
                url,
                json=payload,
                headers=headers,
                timeout=aiohttp.ClientTimeout(total=30),
            ) as response:
                return {
                    "delivery_id": delivery_id,
                    "status": "delivered" if response.status < 400 else "failed",
                    "response_status": response.status,
                    "response_body": await response.text(),
                }
    except Exception as e:
        return {"delivery_id": delivery_id, "status": "failed", "error": str(e)}


async def process_webhook_delivery(
    webhook_data: Dict[str, Any],
    event_type: str,
    event_data: Dict[str, Any],
    event_id: str,
):
    """Process webhook delivery in background."""
    try:
        # Create payload
        payload = {
            "event_id": event_id,
            "event_type": event_type,
            "timestamp": datetime.now(timezone.utc).isoformat(),
            "data": event_data,
        }

        # Queue async delivery
        if settings.USE_CLOUD_TASKS and tasks_client:
            await deliver_webhook_async(webhook_data, payload)
        else:
            # Fallback to sync delivery
            await deliver_webhook_sync(webhook_data, payload)

    except Exception as e:
        logger.error(f"Failed to process webhook delivery: {e}")
