"""
Pub/Sub client for event publishing and subscriptions.

This module provides async Pub/Sub operations for event-driven architecture,
webhook notifications, and dead letter queue handling.
"""

import asyncio
import json
from datetime import datetime, timezone
from typing import Any, Callable, Dict, List, Optional
from enum import Enum

from google.api_core import retry, exceptions
from google.cloud import pubsub_v1
from google.cloud.pubsub_v1.types import (
    ReceivedMessage,
    DeadLetterPolicy,
    RetryPolicy,
    ExpirationPolicy,
)
from google.pubsub_v1.services.subscriber import SubscriberAsyncClient
from google.pubsub_v1.services.publisher import PublisherAsyncClient

from app.config import settings
from app.utils.logger import setup_logger

logger = setup_logger(__name__)


class EventType(str, Enum):
    """Supported event types."""

    COMMAND_EXECUTED = "command.executed"
    COMMAND_FAILED = "command.failed"
    TEST_STARTED = "test.started"
    TEST_COMPLETED = "test.completed"
    TEST_FAILED = "test.failed"
    SESSION_CREATED = "session.created"
    SESSION_EXPIRED = "session.expired"
    BATCH_STARTED = "batch.started"
    BATCH_COMPLETED = "batch.completed"
    BATCH_FAILED = "batch.failed"
    WEBHOOK_RECEIVED = "webhook.received"
    API_KEY_CREATED = "api_key.created"
    API_KEY_REVOKED = "api_key.revoked"


class PubSubClient:
    """
    Async Pub/Sub client for event publishing and subscriptions.
    """

    def __init__(self):
        """Initialize Pub/Sub client."""
        self.project_id = settings.GCP_PROJECT_ID

        # Topic names
        self.TOPICS = {
            EventType.COMMAND_EXECUTED: "virtuoso-command-events",
            EventType.COMMAND_FAILED: "virtuoso-command-events",
            EventType.TEST_STARTED: "virtuoso-test-events",
            EventType.TEST_COMPLETED: "virtuoso-test-events",
            EventType.TEST_FAILED: "virtuoso-test-events",
            EventType.SESSION_CREATED: "virtuoso-session-events",
            EventType.SESSION_EXPIRED: "virtuoso-session-events",
            EventType.BATCH_STARTED: "virtuoso-batch-events",
            EventType.BATCH_COMPLETED: "virtuoso-batch-events",
            EventType.BATCH_FAILED: "virtuoso-batch-events",
            EventType.WEBHOOK_RECEIVED: "virtuoso-webhook-events",
            EventType.API_KEY_CREATED: "virtuoso-security-events",
            EventType.API_KEY_REVOKED: "virtuoso-security-events",
        }

        # Dead letter topic
        self.DEAD_LETTER_TOPIC = "virtuoso-dead-letter"

        # Initialize clients
        self._publisher: Optional[PublisherAsyncClient] = None
        self._subscriber: Optional[SubscriberAsyncClient] = None

        # Active subscriptions
        self._subscriptions: Dict[str, asyncio.Task] = {}
        self._handlers: Dict[str, List[Callable]] = {}

        # Publisher settings
        self.batch_settings = pubsub_v1.types.BatchSettings(
            max_messages=100,
            max_bytes=1024 * 1024,  # 1MB
            max_latency=0.1,  # 100ms
        )

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

        logger.info(f"Initialized Pub/Sub client for project: {self.project_id}")

    async def _get_publisher(self) -> PublisherAsyncClient:
        """Get or create publisher client."""
        if self._publisher is None:
            self._publisher = PublisherAsyncClient()
        return self._publisher

    async def _get_subscriber(self) -> SubscriberAsyncClient:
        """Get or create subscriber client."""
        if self._subscriber is None:
            self._subscriber = SubscriberAsyncClient()
        return self._subscriber

    async def close(self):
        """Close Pub/Sub client connections."""
        # Cancel all active subscriptions
        for subscription_name, task in self._subscriptions.items():
            task.cancel()
            try:
                await task
            except asyncio.CancelledError:
                pass

        self._subscriptions.clear()

        # Close clients
        if self._publisher:
            await self._publisher.transport.close()
            self._publisher = None

        if self._subscriber:
            await self._subscriber.transport.close()
            self._subscriber = None

    def _get_topic_path(self, topic_name: str) -> str:
        """Get full topic path."""
        return f"projects/{self.project_id}/topics/{topic_name}"

    def _get_subscription_path(self, subscription_name: str) -> str:
        """Get full subscription path."""
        return f"projects/{self.project_id}/subscriptions/{subscription_name}"

    # Topic Management

    async def create_topic(self, topic_name: str) -> str:
        """Create a new topic."""
        try:
            publisher = await self._get_publisher()
            topic_path = self._get_topic_path(topic_name)

            response = await publisher.create_topic(
                name=topic_path,
                retry=self._retry,
            )

            logger.info(f"Created topic: {topic_name}")
            return response.name

        except exceptions.AlreadyExists:
            logger.info(f"Topic already exists: {topic_name}")
            return self._get_topic_path(topic_name)
        except Exception as e:
            logger.error(f"Failed to create topic: {str(e)}")
            raise

    async def delete_topic(self, topic_name: str) -> bool:
        """Delete a topic."""
        try:
            publisher = await self._get_publisher()
            topic_path = self._get_topic_path(topic_name)

            await publisher.delete_topic(
                topic=topic_path,
                retry=self._retry,
            )

            logger.info(f"Deleted topic: {topic_name}")
            return True

        except exceptions.NotFound:
            return False
        except Exception as e:
            logger.error(f"Failed to delete topic: {str(e)}")
            raise

    # Publishing

    async def publish_event(
        self,
        event_type: EventType,
        data: Dict[str, Any],
        attributes: Optional[Dict[str, str]] = None,
        ordering_key: Optional[str] = None,
    ) -> str:
        """Publish an event to the appropriate topic."""
        try:
            publisher = await self._get_publisher()
            topic_name = self.TOPICS.get(event_type)

            if not topic_name:
                raise ValueError(f"Unknown event type: {event_type}")

            topic_path = self._get_topic_path(topic_name)

            # Prepare message
            message_data = {
                "event_type": event_type.value,
                "timestamp": datetime.now(timezone.utc).isoformat(),
                "data": data,
            }

            # Add attributes
            message_attributes = {
                "event_type": event_type.value,
                **(attributes or {}),
            }

            # Publish message
            future = await publisher.publish(
                topic=topic_path,
                data=json.dumps(message_data).encode("utf-8"),
                ordering_key=ordering_key or "",
                retry=self._retry,
                **{k: v for k, v in message_attributes.items()},
            )

            message_id = future.result()
            logger.info(f"Published {event_type} event: {message_id}")

            return message_id

        except Exception as e:
            logger.error(f"Failed to publish event: {str(e)}")
            raise

    async def publish_batch(
        self, events: List[Dict[str, Any]], topic_name: str
    ) -> List[str]:
        """Publish multiple events in a batch."""
        try:
            publisher = await self._get_publisher()
            topic_path = self._get_topic_path(topic_name)

            futures = []

            for event in events:
                message_data = json.dumps(event).encode("utf-8")

                future = publisher.publish(
                    topic=topic_path,
                    data=message_data,
                    retry=self._retry,
                )
                futures.append(future)

            # Wait for all messages to be published
            message_ids = []
            for future in futures:
                try:
                    message_id = await future
                    message_ids.append(message_id)
                except Exception as e:
                    logger.error(f"Failed to publish message in batch: {str(e)}")

            logger.info(f"Published batch of {len(message_ids)} messages")
            return message_ids

        except Exception as e:
            logger.error(f"Failed to publish batch: {str(e)}")
            raise

    # Subscription Management

    async def create_subscription(
        self,
        subscription_name: str,
        topic_name: str,
        ack_deadline_seconds: int = 600,
        enable_message_ordering: bool = False,
        enable_dead_letter: bool = True,
        max_delivery_attempts: int = 5,
        filter_expression: Optional[str] = None,
    ) -> str:
        """Create a new subscription."""
        try:
            subscriber = await self._get_subscriber()
            subscription_path = self._get_subscription_path(subscription_name)
            topic_path = self._get_topic_path(topic_name)

            # Configure dead letter policy
            dead_letter_policy = None
            if enable_dead_letter:
                dead_letter_topic_path = self._get_topic_path(self.DEAD_LETTER_TOPIC)
                dead_letter_policy = DeadLetterPolicy(
                    dead_letter_topic=dead_letter_topic_path,
                    max_delivery_attempts=max_delivery_attempts,
                )

            # Configure retry policy
            retry_policy = RetryPolicy(
                minimum_backoff={"seconds": 10},
                maximum_backoff={"seconds": 300},
            )

            # Configure expiration policy (delete subscription if inactive)
            expiration_policy = ExpirationPolicy(
                ttl={"seconds": 86400 * 31},  # 31 days
            )

            # Create subscription request
            request = {
                "name": subscription_path,
                "topic": topic_path,
                "ack_deadline_seconds": ack_deadline_seconds,
                "enable_message_ordering": enable_message_ordering,
                "retry_policy": retry_policy,
                "expiration_policy": expiration_policy,
            }

            if dead_letter_policy:
                request["dead_letter_policy"] = dead_letter_policy

            if filter_expression:
                request["filter"] = filter_expression

            response = await subscriber.create_subscription(
                request=request,
                retry=self._retry,
            )

            logger.info(f"Created subscription: {subscription_name}")
            return response.name

        except exceptions.AlreadyExists:
            logger.info(f"Subscription already exists: {subscription_name}")
            return self._get_subscription_path(subscription_name)
        except Exception as e:
            logger.error(f"Failed to create subscription: {str(e)}")
            raise

    async def delete_subscription(self, subscription_name: str) -> bool:
        """Delete a subscription."""
        try:
            subscriber = await self._get_subscriber()
            subscription_path = self._get_subscription_path(subscription_name)

            await subscriber.delete_subscription(
                subscription=subscription_path,
                retry=self._retry,
            )

            logger.info(f"Deleted subscription: {subscription_name}")
            return True

        except exceptions.NotFound:
            return False
        except Exception as e:
            logger.error(f"Failed to delete subscription: {str(e)}")
            raise

    # Message Handling

    def register_handler(self, event_type: EventType, handler: Callable):
        """Register a handler for an event type."""
        if event_type not in self._handlers:
            self._handlers[event_type] = []

        self._handlers[event_type].append(handler)
        logger.info(f"Registered handler for {event_type}")

    async def _process_message(self, message: ReceivedMessage):
        """Process a received message."""
        try:
            # Parse message
            data = json.loads(message.message.data.decode("utf-8"))
            event_type = EventType(data.get("event_type"))

            # Get handlers for this event type
            handlers = self._handlers.get(event_type, [])

            if not handlers:
                logger.warning(f"No handlers registered for {event_type}")
                return

            # Call all registered handlers
            for handler in handlers:
                try:
                    if asyncio.iscoroutinefunction(handler):
                        await handler(data)
                    else:
                        handler(data)
                except Exception as e:
                    logger.error(f"Handler error for {event_type}: {str(e)}")

            # Acknowledge message
            await message.ack()

        except Exception as e:
            logger.error(f"Failed to process message: {str(e)}")
            # Message will be redelivered

    async def subscribe(
        self, subscription_name: str, max_messages: int = 10, auto_ack: bool = False
    ):
        """Start listening to a subscription."""
        if subscription_name in self._subscriptions:
            logger.warning(f"Already subscribed to: {subscription_name}")
            return

        async def pull_messages():
            """Pull messages from subscription."""
            subscriber = await self._get_subscriber()
            subscription_path = self._get_subscription_path(subscription_name)

            while True:
                try:
                    # Pull messages
                    response = await subscriber.pull(
                        subscription=subscription_path,
                        max_messages=max_messages,
                        retry=self._retry,
                    )

                    if not response.received_messages:
                        await asyncio.sleep(1)
                        continue

                    # Process messages
                    tasks = []
                    for message in response.received_messages:
                        if auto_ack:
                            await subscriber.acknowledge(
                                subscription=subscription_path,
                                ack_ids=[message.ack_id],
                            )
                        else:
                            task = asyncio.create_task(self._process_message(message))
                            tasks.append(task)

                    if tasks:
                        await asyncio.gather(*tasks, return_exceptions=True)

                except asyncio.CancelledError:
                    break
                except Exception as e:
                    logger.error(f"Error pulling messages: {str(e)}")
                    await asyncio.sleep(5)

        # Start subscription task
        task = asyncio.create_task(pull_messages())
        self._subscriptions[subscription_name] = task

        logger.info(f"Started subscription: {subscription_name}")

    async def unsubscribe(self, subscription_name: str):
        """Stop listening to a subscription."""
        if subscription_name not in self._subscriptions:
            return

        task = self._subscriptions[subscription_name]
        task.cancel()

        try:
            await task
        except asyncio.CancelledError:
            pass

        del self._subscriptions[subscription_name]
        logger.info(f"Stopped subscription: {subscription_name}")

    # Webhook Support

    async def publish_webhook_event(
        self,
        webhook_url: str,
        event_data: Dict[str, Any],
        headers: Optional[Dict[str, str]] = None,
        retry_count: int = 3,
    ) -> str:
        """Publish a webhook event."""
        webhook_data = {
            "webhook_url": webhook_url,
            "event_data": event_data,
            "headers": headers or {},
            "retry_count": retry_count,
            "attempt": 0,
        }

        return await self.publish_event(
            EventType.WEBHOOK_RECEIVED,
            webhook_data,
            attributes={
                "webhook_url": webhook_url,
                "retry_count": str(retry_count),
            },
        )

    # Dead Letter Queue Handling

    async def process_dead_letter_queue(
        self, handler: Callable, subscription_name: str = "virtuoso-dead-letter-sub"
    ):
        """Process messages from the dead letter queue."""
        # Create dead letter subscription if needed
        await self.create_subscription(
            subscription_name,
            self.DEAD_LETTER_TOPIC,
            enable_dead_letter=False,  # Don't create another DLQ
        )

        # Register handler
        self.register_handler(EventType.WEBHOOK_RECEIVED, handler)

        # Start subscription
        await self.subscribe(subscription_name)

    # Monitoring

    async def get_subscription_metrics(self, subscription_name: str) -> Dict[str, Any]:
        """Get subscription metrics."""
        try:
            subscriber = await self._get_subscriber()
            subscription_path = self._get_subscription_path(subscription_name)

            subscription = await subscriber.get_subscription(
                subscription=subscription_path,
                retry=self._retry,
            )

            # Get snapshot if available
            snapshot = None
            if hasattr(subscription, "snapshot") and subscription.snapshot:
                snapshot = subscription.snapshot

            return {
                "name": subscription.name,
                "topic": subscription.topic,
                "ack_deadline_seconds": subscription.ack_deadline_seconds,
                "message_retention_duration": subscription.message_retention_duration,
                "enable_message_ordering": subscription.enable_message_ordering,
                "dead_letter_policy": {
                    "topic": subscription.dead_letter_policy.dead_letter_topic,
                    "max_attempts": subscription.dead_letter_policy.max_delivery_attempts,
                }
                if subscription.dead_letter_policy
                else None,
                "snapshot": snapshot,
            }

        except Exception as e:
            logger.error(f"Failed to get subscription metrics: {str(e)}")
            raise
