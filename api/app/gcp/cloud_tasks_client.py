"""
Cloud Tasks client for async command execution and batch processing.

This module provides async Cloud Tasks operations for handling long-running
test suites and batch command processing with proper retry configuration.
"""

import asyncio
import json
from datetime import datetime, timedelta, timezone
from typing import Any, Dict, List, Optional
from enum import Enum

from google.api_core import retry, exceptions
from google.cloud import tasks_v2
from google.cloud.tasks_v2 import CloudTasksAsyncClient
from google.cloud.tasks_v2.types import (
    Task,
    HttpRequest,
    HttpMethod,
    OidcToken,
    RetryConfig,
    RateLimits,
    Queue,
)
from google.protobuf import duration_pb2, timestamp_pb2

from app.config import settings
from app.utils.logger import setup_logger

logger = setup_logger(__name__)


class TaskPriority(str, Enum):
    """Task priority levels."""

    HIGH = "high"
    NORMAL = "normal"
    LOW = "low"


class CloudTasksClient:
    """
    Async Cloud Tasks client for command execution and batch processing.
    """

    def __init__(self):
        """Initialize Cloud Tasks client."""
        self.project_id = settings.GCP_PROJECT_ID
        self.location = settings.GCP_LOCATION
        self.service_account_email = settings.GCP_SERVICE_ACCOUNT_EMAIL

        # Queue names by priority
        self.QUEUES = {
            TaskPriority.HIGH: "virtuoso-commands-high",
            TaskPriority.NORMAL: "virtuoso-commands",
            TaskPriority.LOW: "virtuoso-commands-low",
        }

        self.BATCH_QUEUE = "virtuoso-batch-processing"
        self.TEST_SUITE_QUEUE = "virtuoso-test-suites"

        # Initialize client
        self._client: Optional[CloudTasksAsyncClient] = None

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

        # Task configuration
        self.max_retry_attempts = 5
        self.max_retry_duration = timedelta(hours=2)
        self.min_backoff = timedelta(seconds=10)
        self.max_backoff = timedelta(minutes=5)

        logger.info(f"Initialized Cloud Tasks client for project: {self.project_id}")

    async def _get_client(self) -> CloudTasksAsyncClient:
        """Get or create Cloud Tasks client."""
        if self._client is None:
            self._client = CloudTasksAsyncClient()
        return self._client

    async def close(self):
        """Close Cloud Tasks client connection."""
        if self._client:
            await self._client.transport.close()
            self._client = None

    def _get_queue_path(self, queue_name: str) -> str:
        """Get full queue path."""
        return (
            f"projects/{self.project_id}/locations/{self.location}/queues/{queue_name}"
        )

    def _create_retry_config(self) -> RetryConfig:
        """Create retry configuration for tasks."""
        return RetryConfig(
            max_attempts=self.max_retry_attempts,
            max_retry_duration=duration_pb2.Duration(
                seconds=int(self.max_retry_duration.total_seconds())
            ),
            min_backoff=duration_pb2.Duration(
                seconds=int(self.min_backoff.total_seconds())
            ),
            max_backoff=duration_pb2.Duration(
                seconds=int(self.max_backoff.total_seconds())
            ),
            max_doublings=3,
        )

    # Queue Management

    async def create_queue(
        self,
        queue_name: str,
        max_concurrent_dispatches: int = 10,
        max_dispatches_per_second: float = 10.0,
        max_retry_attempts: Optional[int] = None,
    ) -> Queue:
        """Create a new task queue."""
        try:
            client = await self._get_client()
            parent = f"projects/{self.project_id}/locations/{self.location}"

            queue = Queue(
                name=self._get_queue_path(queue_name),
                rate_limits=RateLimits(
                    max_concurrent_dispatches=max_concurrent_dispatches,
                    max_dispatches_per_second=max_dispatches_per_second,
                ),
                retry_config=RetryConfig(
                    max_attempts=max_retry_attempts or self.max_retry_attempts,
                ),
            )

            response = await client.create_queue(
                parent=parent,
                queue=queue,
                retry=self._retry,
            )

            logger.info(f"Created queue: {queue_name}")
            return response

        except exceptions.AlreadyExists:
            logger.info(f"Queue already exists: {queue_name}")
            return await self.get_queue(queue_name)
        except Exception as e:
            logger.error(f"Failed to create queue: {str(e)}")
            raise

    async def get_queue(self, queue_name: str) -> Optional[Queue]:
        """Get queue information."""
        try:
            client = await self._get_client()
            queue_path = self._get_queue_path(queue_name)

            response = await client.get_queue(
                name=queue_path,
                retry=self._retry,
            )

            return response

        except exceptions.NotFound:
            return None
        except Exception as e:
            logger.error(f"Failed to get queue: {str(e)}")
            raise

    async def pause_queue(self, queue_name: str) -> Queue:
        """Pause a queue."""
        try:
            client = await self._get_client()
            queue_path = self._get_queue_path(queue_name)

            response = await client.pause_queue(
                name=queue_path,
                retry=self._retry,
            )

            logger.info(f"Paused queue: {queue_name}")
            return response

        except Exception as e:
            logger.error(f"Failed to pause queue: {str(e)}")
            raise

    async def resume_queue(self, queue_name: str) -> Queue:
        """Resume a paused queue."""
        try:
            client = await self._get_client()
            queue_path = self._get_queue_path(queue_name)

            response = await client.resume_queue(
                name=queue_path,
                retry=self._retry,
            )

            logger.info(f"Resumed queue: {queue_name}")
            return response

        except Exception as e:
            logger.error(f"Failed to resume queue: {str(e)}")
            raise

    # Task Creation

    async def create_command_task(
        self,
        command: str,
        args: List[str],
        checkpoint_id: Optional[str] = None,
        user_id: Optional[str] = None,
        session_id: Optional[str] = None,
        priority: TaskPriority = TaskPriority.NORMAL,
        delay_seconds: Optional[int] = None,
        task_id: Optional[str] = None,
        callback_url: Optional[str] = None,
    ) -> Task:
        """Create a task for async command execution."""
        try:
            client = await self._get_client()
            queue_name = self.QUEUES[priority]
            parent = self._get_queue_path(queue_name)

            # Build task payload
            payload = {
                "command": command,
                "args": args,
                "checkpoint_id": checkpoint_id,
                "user_id": user_id,
                "session_id": session_id,
                "callback_url": callback_url,
                "timestamp": datetime.now(timezone.utc).isoformat(),
            }

            # Build HTTP request
            http_request = HttpRequest(
                http_method=HttpMethod.POST,
                url=f"{settings.API_BASE_URL}/api/v1/commands/execute",
                headers={
                    "Content-Type": "application/json",
                    "X-CloudTasks-QueueName": queue_name,
                    "X-CloudTasks-TaskName": task_id or "",
                },
                body=json.dumps(payload).encode(),
            )

            # Add OIDC token for authentication
            if self.service_account_email:
                http_request.oidc_token = OidcToken(
                    service_account_email=self.service_account_email,
                    audience=settings.API_BASE_URL,
                )

            # Build task
            task = Task(
                http_request=http_request,
                retry_config=self._create_retry_config(),
            )

            # Set schedule time if delayed
            if delay_seconds:
                schedule_time = timestamp_pb2.Timestamp()
                schedule_time.FromDatetime(
                    datetime.now(timezone.utc) + timedelta(seconds=delay_seconds)
                )
                task.schedule_time = schedule_time

            # Create task
            request = tasks_v2.CreateTaskRequest(
                parent=parent,
                task=task,
            )

            if task_id:
                request.task.name = f"{parent}/tasks/{task_id}"

            response = await client.create_task(
                request=request,
                retry=self._retry,
            )

            logger.info(f"Created command task: {response.name}")
            return response

        except Exception as e:
            logger.error(f"Failed to create command task: {str(e)}")
            raise

    async def create_batch_task(
        self,
        commands: List[Dict[str, Any]],
        batch_id: str,
        user_id: Optional[str] = None,
        parallel: bool = True,
        max_parallel: int = 10,
        stop_on_error: bool = False,
        callback_url: Optional[str] = None,
        delay_seconds: Optional[int] = None,
    ) -> Task:
        """Create a task for batch command processing."""
        try:
            client = await self._get_client()
            parent = self._get_queue_path(self.BATCH_QUEUE)

            # Build batch payload
            payload = {
                "batch_id": batch_id,
                "commands": commands,
                "user_id": user_id,
                "parallel": parallel,
                "max_parallel": max_parallel,
                "stop_on_error": stop_on_error,
                "callback_url": callback_url,
                "timestamp": datetime.now(timezone.utc).isoformat(),
            }

            # Build HTTP request
            http_request = HttpRequest(
                http_method=HttpMethod.POST,
                url=f"{settings.API_BASE_URL}/api/v1/commands/batch",
                headers={
                    "Content-Type": "application/json",
                    "X-CloudTasks-QueueName": self.BATCH_QUEUE,
                    "X-CloudTasks-BatchId": batch_id,
                },
                body=json.dumps(payload).encode(),
            )

            # Add OIDC token
            if self.service_account_email:
                http_request.oidc_token = OidcToken(
                    service_account_email=self.service_account_email,
                    audience=settings.API_BASE_URL,
                )

            # Build task
            task = Task(
                http_request=http_request,
                retry_config=self._create_retry_config(),
            )

            # Set schedule time if delayed
            if delay_seconds:
                schedule_time = timestamp_pb2.Timestamp()
                schedule_time.FromDatetime(
                    datetime.now(timezone.utc) + timedelta(seconds=delay_seconds)
                )
                task.schedule_time = schedule_time

            # Create task
            request = tasks_v2.CreateTaskRequest(
                parent=parent,
                task=task,
            )

            response = await client.create_task(
                request=request,
                retry=self._retry,
            )

            logger.info(f"Created batch task: {batch_id}")
            return response

        except Exception as e:
            logger.error(f"Failed to create batch task: {str(e)}")
            raise

    async def create_test_suite_task(
        self,
        test_suite_path: str,
        project_id: str,
        user_id: Optional[str] = None,
        environment: Optional[str] = None,
        variables: Optional[Dict[str, str]] = None,
        callback_url: Optional[str] = None,
        delay_seconds: Optional[int] = None,
    ) -> Task:
        """Create a task for test suite execution."""
        try:
            client = await self._get_client()
            parent = self._get_queue_path(self.TEST_SUITE_QUEUE)

            # Build test suite payload
            payload = {
                "test_suite_path": test_suite_path,
                "project_id": project_id,
                "user_id": user_id,
                "environment": environment,
                "variables": variables or {},
                "callback_url": callback_url,
                "timestamp": datetime.now(timezone.utc).isoformat(),
            }

            # Build HTTP request
            http_request = HttpRequest(
                http_method=HttpMethod.POST,
                url=f"{settings.API_BASE_URL}/api/v1/tests/run",
                headers={
                    "Content-Type": "application/json",
                    "X-CloudTasks-QueueName": self.TEST_SUITE_QUEUE,
                    "X-CloudTasks-ProjectId": project_id,
                },
                body=json.dumps(payload).encode(),
            )

            # Add OIDC token
            if self.service_account_email:
                http_request.oidc_token = OidcToken(
                    service_account_email=self.service_account_email,
                    audience=settings.API_BASE_URL,
                )

            # Build task with longer timeout for test suites
            task = Task(
                http_request=http_request,
                retry_config=RetryConfig(
                    max_attempts=3,  # Fewer retries for test suites
                    max_retry_duration=duration_pb2.Duration(seconds=3600),  # 1 hour
                ),
            )

            # Set schedule time if delayed
            if delay_seconds:
                schedule_time = timestamp_pb2.Timestamp()
                schedule_time.FromDatetime(
                    datetime.now(timezone.utc) + timedelta(seconds=delay_seconds)
                )
                task.schedule_time = schedule_time

            # Create task
            request = tasks_v2.CreateTaskRequest(
                parent=parent,
                task=task,
            )

            response = await client.create_task(
                request=request,
                retry=self._retry,
            )

            logger.info(f"Created test suite task: {test_suite_path}")
            return response

        except Exception as e:
            logger.error(f"Failed to create test suite task: {str(e)}")
            raise

    # Task Management

    async def get_task(self, queue_name: str, task_id: str) -> Optional[Task]:
        """Get task information."""
        try:
            client = await self._get_client()
            task_name = f"{self._get_queue_path(queue_name)}/tasks/{task_id}"

            response = await client.get_task(
                name=task_name,
                retry=self._retry,
            )

            return response

        except exceptions.NotFound:
            return None
        except Exception as e:
            logger.error(f"Failed to get task: {str(e)}")
            raise

    async def delete_task(self, queue_name: str, task_id: str) -> bool:
        """Delete a task."""
        try:
            client = await self._get_client()
            task_name = f"{self._get_queue_path(queue_name)}/tasks/{task_id}"

            await client.delete_task(
                name=task_name,
                retry=self._retry,
            )

            logger.info(f"Deleted task: {task_id}")
            return True

        except exceptions.NotFound:
            return False
        except Exception as e:
            logger.error(f"Failed to delete task: {str(e)}")
            raise

    async def list_tasks(
        self, queue_name: str, page_size: int = 100, page_token: Optional[str] = None
    ) -> Dict[str, Any]:
        """List tasks in a queue."""
        try:
            client = await self._get_client()
            parent = self._get_queue_path(queue_name)

            response = await client.list_tasks(
                parent=parent,
                page_size=page_size,
                page_token=page_token,
                retry=self._retry,
            )

            tasks = []
            async for task in response:
                tasks.append(
                    {
                        "name": task.name,
                        "schedule_time": task.schedule_time.ToDatetime()
                        if task.schedule_time
                        else None,
                        "create_time": task.create_time.ToDatetime()
                        if task.create_time
                        else None,
                        "dispatch_count": task.dispatch_count,
                        "response_count": task.response_count,
                        "first_attempt": task.first_attempt,
                        "last_attempt": task.last_attempt,
                    }
                )

            return {
                "tasks": tasks,
                "next_page_token": response.next_page_token,
            }

        except Exception as e:
            logger.error(f"Failed to list tasks: {str(e)}")
            raise

    # Bulk Operations

    async def create_command_tasks_bulk(
        self,
        commands: List[Dict[str, Any]],
        priority: TaskPriority = TaskPriority.NORMAL,
        batch_size: int = 500,
    ) -> List[Task]:
        """Create multiple command tasks in bulk."""
        tasks_created = []

        try:
            # Process in batches to avoid API limits
            for i in range(0, len(commands), batch_size):
                batch = commands[i : i + batch_size]

                # Create tasks concurrently within batch
                tasks = []
                for cmd in batch:
                    task = self.create_command_task(
                        command=cmd["command"],
                        args=cmd.get("args", []),
                        checkpoint_id=cmd.get("checkpoint_id"),
                        user_id=cmd.get("user_id"),
                        session_id=cmd.get("session_id"),
                        priority=priority,
                        delay_seconds=cmd.get("delay_seconds"),
                        task_id=cmd.get("task_id"),
                        callback_url=cmd.get("callback_url"),
                    )
                    tasks.append(task)

                # Execute batch concurrently
                results = await asyncio.gather(*tasks, return_exceptions=True)

                for result in results:
                    if isinstance(result, Exception):
                        logger.error(f"Failed to create task: {result}")
                    else:
                        tasks_created.append(result)

                # Add small delay between batches
                if i + batch_size < len(commands):
                    await asyncio.sleep(0.1)

            logger.info(f"Created {len(tasks_created)} tasks in bulk")
            return tasks_created

        except Exception as e:
            logger.error(f"Bulk task creation failed: {str(e)}")
            raise

    # Queue Statistics

    async def get_queue_stats(self, queue_name: str) -> Dict[str, Any]:
        """Get queue statistics."""
        try:
            queue = await self.get_queue(queue_name)
            if not queue:
                return {}

            # Get approximate task count
            client = await self._get_client()
            parent = self._get_queue_path(queue_name)

            task_count = 0
            async for _ in client.list_tasks(parent=parent, page_size=1000):
                task_count += 1

            return {
                "name": queue.name,
                "state": queue.state.name,
                "rate_limits": {
                    "max_concurrent": queue.rate_limits.max_concurrent_dispatches,
                    "max_per_second": queue.rate_limits.max_dispatches_per_second,
                },
                "retry_config": {
                    "max_attempts": queue.retry_config.max_attempts,
                },
                "task_count": task_count,
                "purge_time": queue.purge_time.ToDatetime()
                if queue.purge_time
                else None,
            }

        except Exception as e:
            logger.error(f"Failed to get queue stats: {str(e)}")
            raise

    async def create_webhook_delivery_task(
        self,
        delivery_id: str,
        webhook_id: str,
        webhook_url: str,
        payload: Dict[str, Any],
        secret: Optional[str] = None,
        headers: Optional[Dict[str, str]] = None,
        delay_seconds: Optional[int] = None,
    ) -> Task:
        """Create a task for webhook delivery."""
        try:
            client = await self._get_client()
            parent = self._get_queue_path("virtuoso-webhooks")

            # Build webhook delivery payload
            delivery_payload = {
                "delivery_id": delivery_id,
                "webhook_id": webhook_id,
                "webhook_url": webhook_url,
                "payload": payload,
                "secret": secret,
                "custom_headers": headers or {},
                "timestamp": datetime.now(timezone.utc).isoformat(),
            }

            # Build HTTP request
            http_request = HttpRequest(
                http_method=HttpMethod.POST,
                url=f"{settings.API_BASE_URL}/api/v1/webhooks/deliver",
                headers={
                    "Content-Type": "application/json",
                    "X-CloudTasks-QueueName": "virtuoso-webhooks",
                    "X-CloudTasks-DeliveryId": delivery_id,
                },
                body=json.dumps(delivery_payload).encode(),
            )

            # Add OIDC token for authentication
            if self.service_account_email:
                http_request.oidc_token = OidcToken(
                    service_account_email=self.service_account_email,
                    audience=settings.API_BASE_URL,
                )

            # Build task with webhook-specific retry config
            task = Task(
                http_request=http_request,
                retry_config=RetryConfig(
                    max_attempts=3,
                    max_retry_duration=duration_pb2.Duration(seconds=3600),  # 1 hour
                    min_backoff=duration_pb2.Duration(seconds=30),
                    max_backoff=duration_pb2.Duration(seconds=300),
                ),
            )

            # Set schedule time if delayed
            if delay_seconds:
                schedule_time = timestamp_pb2.Timestamp()
                schedule_time.FromDatetime(
                    datetime.now(timezone.utc) + timedelta(seconds=delay_seconds)
                )
                task.schedule_time = schedule_time

            # Create task
            request = tasks_v2.CreateTaskRequest(
                parent=parent,
                task=task,
            )

            response = await client.create_task(
                request=request,
                retry=self._retry,
            )

            logger.info(f"Created webhook delivery task: {delivery_id}")
            return response

        except Exception as e:
            logger.error(f"Failed to create webhook delivery task: {str(e)}")
            raise
