"""
Firestore client for session management, API key storage, and caching.

This module provides async Firestore operations with proper error handling,
retry logic, and support for local development with the Firestore emulator.
"""

import asyncio
import os
from datetime import datetime, timedelta, timezone
from typing import Any, Dict, List, Optional
from contextlib import asynccontextmanager

from google.api_core import retry
from google.api_core.exceptions import GoogleAPIError, NotFound, AlreadyExists
from google.cloud import firestore
from google.cloud.firestore_v1 import AsyncClient
from google.cloud.firestore_v1.base_query import FieldFilter

from app.config import settings
from app.utils.logger import setup_logger

logger = setup_logger(__name__)


class FirestoreClient:
    """
    Async Firestore client with caching and retry logic.
    """

    def __init__(self):
        """Initialize Firestore client."""
        self.project_id = settings.GCP_PROJECT_ID
        self.emulator_host = os.getenv("FIRESTORE_EMULATOR_HOST")

        # Initialize client (will be created on first use)
        self._client: Optional[AsyncClient] = None
        self._cache: Dict[str, Dict[str, Any]] = {}
        self._cache_ttl = timedelta(seconds=settings.CACHE_TTL)

        # Collections
        self.SESSIONS_COLLECTION = "sessions"
        self.API_KEYS_COLLECTION = "api_keys"
        self.COMMAND_HISTORY_COLLECTION = "command_history"
        self.CACHE_COLLECTION = "cache"

        # Retry configuration
        self._retry = retry.AsyncRetry(
            initial=0.1,
            maximum=60.0,
            multiplier=2.0,
            timeout=300.0,
            predicate=retry.if_exception_type(
                GoogleAPIError,
                asyncio.TimeoutError,
            ),
        )

        if self.emulator_host:
            logger.info(f"Using Firestore emulator at: {self.emulator_host}")
        else:
            logger.info(f"Using Firestore for project: {self.project_id}")

    async def _get_client(self) -> AsyncClient:
        """Get or create Firestore client."""
        if self._client is None:
            self._client = AsyncClient(project=self.project_id)
        return self._client

    async def close(self):
        """Close Firestore client connection."""
        if self._client:
            self._client.close()
            self._client = None

    # Session Management

    async def create_session(
        self,
        session_id: str,
        user_id: str,
        checkpoint_id: Optional[str] = None,
        metadata: Optional[Dict[str, Any]] = None,
    ) -> Dict[str, Any]:
        """Create a new session."""
        try:
            client = await self._get_client()
            session_ref = client.collection(self.SESSIONS_COLLECTION).document(
                session_id
            )

            session_data = {
                "session_id": session_id,
                "user_id": user_id,
                "checkpoint_id": checkpoint_id,
                "metadata": metadata or {},
                "created_at": datetime.now(timezone.utc),
                "updated_at": datetime.now(timezone.utc),
                "expires_at": datetime.now(timezone.utc)
                + timedelta(seconds=settings.SESSION_MAX_AGE),
                "active": True,
            }

            await session_ref.set(session_data, retry=self._retry)
            logger.info(f"Created session: {session_id} for user: {user_id}")

            return session_data

        except Exception as e:
            logger.error(f"Failed to create session: {str(e)}")
            raise

    async def get_session(self, session_id: str) -> Optional[Dict[str, Any]]:
        """Get session by ID."""
        try:
            client = await self._get_client()
            session_ref = client.collection(self.SESSIONS_COLLECTION).document(
                session_id
            )

            doc = await session_ref.get(retry=self._retry)
            if not doc.exists:
                return None

            session_data = doc.to_dict()

            # Check if session is expired
            if session_data.get("expires_at") and session_data[
                "expires_at"
            ] < datetime.now(timezone.utc):
                await self.delete_session(session_id)
                return None

            return session_data

        except NotFound:
            return None
        except Exception as e:
            logger.error(f"Failed to get session: {str(e)}")
            raise

    async def update_session(
        self,
        session_id: str,
        checkpoint_id: Optional[str] = None,
        metadata: Optional[Dict[str, Any]] = None,
        extend_expiry: bool = True,
    ) -> Optional[Dict[str, Any]]:
        """Update existing session."""
        try:
            client = await self._get_client()
            session_ref = client.collection(self.SESSIONS_COLLECTION).document(
                session_id
            )

            update_data = {
                "updated_at": datetime.now(timezone.utc),
            }

            if checkpoint_id is not None:
                update_data["checkpoint_id"] = checkpoint_id

            if metadata is not None:
                update_data["metadata"] = metadata

            if extend_expiry:
                update_data["expires_at"] = datetime.now(timezone.utc) + timedelta(
                    seconds=settings.SESSION_MAX_AGE
                )

            await session_ref.update(update_data, retry=self._retry)

            # Return updated session
            return await self.get_session(session_id)

        except NotFound:
            return None
        except Exception as e:
            logger.error(f"Failed to update session: {str(e)}")
            raise

    async def delete_session(self, session_id: str) -> bool:
        """Delete session."""
        try:
            client = await self._get_client()
            session_ref = client.collection(self.SESSIONS_COLLECTION).document(
                session_id
            )

            await session_ref.delete(retry=self._retry)
            logger.info(f"Deleted session: {session_id}")

            return True

        except Exception as e:
            logger.error(f"Failed to delete session: {str(e)}")
            return False

    async def list_user_sessions(
        self, user_id: str, active_only: bool = True
    ) -> List[Dict[str, Any]]:
        """List all sessions for a user."""
        try:
            client = await self._get_client()
            query = client.collection(self.SESSIONS_COLLECTION).where(
                filter=FieldFilter("user_id", "==", user_id)
            )

            if active_only:
                query = query.where(filter=FieldFilter("active", "==", True))

            docs = await query.get(retry=self._retry)

            sessions = []
            for doc in docs:
                session_data = doc.to_dict()
                # Filter out expired sessions
                if session_data.get("expires_at") and session_data[
                    "expires_at"
                ] >= datetime.now(timezone.utc):
                    sessions.append(session_data)

            return sessions

        except Exception as e:
            logger.error(f"Failed to list user sessions: {str(e)}")
            raise

    # API Key Management

    async def validate_api_key(self, api_key: str) -> Optional[Dict[str, Any]]:
        """Validate API key and return associated metadata."""
        try:
            client = await self._get_client()

            # Query by hashed key for security
            import hashlib

            key_hash = hashlib.sha256(api_key.encode()).hexdigest()

            query = (
                client.collection(self.API_KEYS_COLLECTION)
                .where(filter=FieldFilter("key_hash", "==", key_hash))
                .where(filter=FieldFilter("active", "==", True))
            )

            docs = await query.get(retry=self._retry)

            if not docs:
                return None

            key_data = docs[0].to_dict()

            # Check expiration
            if key_data.get("expires_at") and key_data["expires_at"] < datetime.now(
                timezone.utc
            ):
                return None

            # Update last used
            await docs[0].reference.update(
                {
                    "last_used_at": datetime.now(timezone.utc),
                    "usage_count": firestore.Increment(1),
                }
            )

            return {
                "user_id": key_data.get("user_id"),
                "tenant_id": key_data.get("tenant_id"),
                "permissions": key_data.get("permissions", []),
                "metadata": key_data.get("metadata", {}),
            }

        except Exception as e:
            logger.error(f"Failed to validate API key: {str(e)}")
            return None

    async def create_api_key(
        self,
        user_id: str,
        api_key: str,
        tenant_id: Optional[str] = None,
        permissions: Optional[List[str]] = None,
        expires_in_days: Optional[int] = None,
        metadata: Optional[Dict[str, Any]] = None,
    ) -> Dict[str, Any]:
        """Create a new API key."""
        try:
            client = await self._get_client()

            import hashlib

            key_hash = hashlib.sha256(api_key.encode()).hexdigest()

            key_data = {
                "key_hash": key_hash,
                "user_id": user_id,
                "tenant_id": tenant_id,
                "permissions": permissions or [],
                "metadata": metadata or {},
                "created_at": datetime.now(timezone.utc),
                "last_used_at": None,
                "usage_count": 0,
                "active": True,
            }

            if expires_in_days:
                key_data["expires_at"] = datetime.now(timezone.utc) + timedelta(
                    days=expires_in_days
                )

            # Use key hash as document ID
            key_ref = client.collection(self.API_KEYS_COLLECTION).document(key_hash)
            await key_ref.set(key_data, retry=self._retry)

            logger.info(f"Created API key for user: {user_id}")
            return key_data

        except AlreadyExists:
            raise ValueError("API key already exists")
        except Exception as e:
            logger.error(f"Failed to create API key: {str(e)}")
            raise

    async def revoke_api_key(self, api_key: str) -> bool:
        """Revoke an API key."""
        try:
            client = await self._get_client()

            import hashlib

            key_hash = hashlib.sha256(api_key.encode()).hexdigest()

            key_ref = client.collection(self.API_KEYS_COLLECTION).document(key_hash)
            await key_ref.update(
                {
                    "active": False,
                    "revoked_at": datetime.now(timezone.utc),
                },
                retry=self._retry,
            )

            logger.info(f"Revoked API key: {key_hash[:8]}...")
            return True

        except NotFound:
            return False
        except Exception as e:
            logger.error(f"Failed to revoke API key: {str(e)}")
            raise

    # Command History

    async def log_command(
        self,
        user_id: str,
        command: str,
        args: List[str],
        checkpoint_id: Optional[str] = None,
        session_id: Optional[str] = None,
        result: Optional[Dict[str, Any]] = None,
        error: Optional[str] = None,
        duration_ms: Optional[int] = None,
    ) -> str:
        """Log a command execution."""
        try:
            client = await self._get_client()

            command_data = {
                "user_id": user_id,
                "command": command,
                "args": args,
                "checkpoint_id": checkpoint_id,
                "session_id": session_id,
                "result": result,
                "error": error,
                "duration_ms": duration_ms,
                "timestamp": datetime.now(timezone.utc),
                "success": error is None,
            }

            doc_ref = await client.collection(self.COMMAND_HISTORY_COLLECTION).add(
                command_data, retry=self._retry
            )

            return doc_ref[1].id

        except Exception as e:
            logger.error(f"Failed to log command: {str(e)}")
            # Don't raise - logging shouldn't break the command
            return ""

    async def get_command_history(
        self,
        user_id: Optional[str] = None,
        checkpoint_id: Optional[str] = None,
        session_id: Optional[str] = None,
        limit: int = 100,
        offset: int = 0,
    ) -> List[Dict[str, Any]]:
        """Get command history with filters."""
        try:
            client = await self._get_client()
            query = client.collection(self.COMMAND_HISTORY_COLLECTION)

            if user_id:
                query = query.where(filter=FieldFilter("user_id", "==", user_id))

            if checkpoint_id:
                query = query.where(
                    filter=FieldFilter("checkpoint_id", "==", checkpoint_id)
                )

            if session_id:
                query = query.where(filter=FieldFilter("session_id", "==", session_id))

            query = query.order_by("timestamp", direction=firestore.Query.DESCENDING)
            query = query.limit(limit).offset(offset)

            docs = await query.get(retry=self._retry)

            return [doc.to_dict() for doc in docs]

        except Exception as e:
            logger.error(f"Failed to get command history: {str(e)}")
            raise

    # Caching Layer

    async def cache_get(self, key: str) -> Optional[Any]:
        """Get value from cache."""
        # Check in-memory cache first
        if key in self._cache:
            cache_entry = self._cache[key]
            if cache_entry["expires_at"] > datetime.now(timezone.utc):
                return cache_entry["value"]
            else:
                del self._cache[key]

        # Check Firestore cache
        try:
            client = await self._get_client()
            cache_ref = client.collection(self.CACHE_COLLECTION).document(key)

            doc = await cache_ref.get(retry=self._retry)
            if not doc.exists:
                return None

            cache_data = doc.to_dict()

            # Check expiration
            if cache_data.get("expires_at") and cache_data["expires_at"] < datetime.now(
                timezone.utc
            ):
                await cache_ref.delete()
                return None

            # Store in memory cache
            self._cache[key] = {
                "value": cache_data["value"],
                "expires_at": cache_data["expires_at"],
            }

            return cache_data["value"]

        except Exception as e:
            logger.error(f"Cache get failed: {str(e)}")
            return None

    async def cache_set(
        self, key: str, value: Any, ttl_seconds: Optional[int] = None
    ) -> bool:
        """Set value in cache with optional TTL."""
        try:
            ttl = ttl_seconds or settings.CACHE_TTL
            expires_at = datetime.now(timezone.utc) + timedelta(seconds=ttl)

            # Store in memory cache
            self._cache[key] = {
                "value": value,
                "expires_at": expires_at,
            }

            # Store in Firestore
            client = await self._get_client()
            cache_ref = client.collection(self.CACHE_COLLECTION).document(key)

            cache_data = {
                "value": value,
                "expires_at": expires_at,
                "created_at": datetime.now(timezone.utc),
            }

            await cache_ref.set(cache_data, retry=self._retry)
            return True

        except Exception as e:
            logger.error(f"Cache set failed: {str(e)}")
            return False

    async def cache_delete(self, key: str) -> bool:
        """Delete value from cache."""
        # Remove from memory cache
        if key in self._cache:
            del self._cache[key]

        try:
            client = await self._get_client()
            cache_ref = client.collection(self.CACHE_COLLECTION).document(key)
            await cache_ref.delete(retry=self._retry)
            return True

        except Exception as e:
            logger.error(f"Cache delete failed: {str(e)}")
            return False

    async def cache_clear(self, pattern: Optional[str] = None) -> int:
        """Clear cache entries matching pattern."""
        count = 0

        # Clear memory cache
        if pattern:
            keys_to_delete = [k for k in self._cache.keys() if pattern in k]
            for key in keys_to_delete:
                del self._cache[key]
                count += 1
        else:
            count = len(self._cache)
            self._cache.clear()

        # Clear Firestore cache
        try:
            client = await self._get_client()

            if pattern:
                # This is inefficient but Firestore doesn't support wildcard deletes
                docs = await client.collection(self.CACHE_COLLECTION).get()
                batch = client.batch()

                for doc in docs:
                    if pattern in doc.id:
                        batch.delete(doc.reference)
                        count += 1

                await batch.commit()
            else:
                # Delete all cache documents
                docs = await client.collection(self.CACHE_COLLECTION).get()
                batch = client.batch()

                for doc in docs:
                    batch.delete(doc.reference)
                    count += 1

                await batch.commit()

            return count

        except Exception as e:
            logger.error(f"Cache clear failed: {str(e)}")
            return count

    # Transaction support

    @asynccontextmanager
    async def transaction(self):
        """Create a Firestore transaction context."""
        client = await self._get_client()
        transaction = client.transaction()

        try:
            yield transaction
            await transaction.commit()
        except Exception:
            # Transaction will be automatically rolled back
            raise

    # Cleanup expired data

    async def cleanup_expired_data(self):
        """Clean up expired sessions, cache entries, etc."""
        try:
            client = await self._get_client()
            now = datetime.now(timezone.utc)

            # Clean up expired sessions
            expired_sessions = client.collection(self.SESSIONS_COLLECTION).where(
                filter=FieldFilter("expires_at", "<", now)
            )

            docs = await expired_sessions.get()
            if docs:
                batch = client.batch()
                for doc in docs:
                    batch.delete(doc.reference)
                await batch.commit()
                logger.info(f"Cleaned up {len(docs)} expired sessions")

            # Clean up expired cache entries
            expired_cache = client.collection(self.CACHE_COLLECTION).where(
                filter=FieldFilter("expires_at", "<", now)
            )

            docs = await expired_cache.get()
            if docs:
                batch = client.batch()
                for doc in docs:
                    batch.delete(doc.reference)
                await batch.commit()
                logger.info(f"Cleaned up {len(docs)} expired cache entries")

        except Exception as e:
            logger.error(f"Failed to cleanup expired data: {str(e)}")

    # Test Management

    async def create_test_run(
        self,
        test_id: str,
        user_id: str,
        project_id: Optional[str] = None,
        checkpoint_id: Optional[str] = None,
        definition: Optional[Dict[str, Any]] = None,
        status: str = "created",
        task_name: Optional[str] = None,
        steps_count: Optional[int] = None,
    ) -> Dict[str, Any]:
        """Create a test run record."""
        try:
            client = await self._get_client()

            test_run_data = {
                "test_id": test_id,
                "user_id": user_id,
                "project_id": project_id,
                "checkpoint_id": checkpoint_id,
                "definition": definition,
                "status": status,
                "task_name": task_name,
                "steps_count": steps_count,
                "created_at": datetime.now(timezone.utc),
                "updated_at": datetime.now(timezone.utc),
            }

            doc_ref = (
                await client.collection("test_runs")
                .document(test_id)
                .set(test_run_data, retry=self._retry)
            )

            logger.info(f"Created test run: {test_id}")
            return test_run_data

        except Exception as e:
            logger.error(f"Failed to create test run: {str(e)}")
            raise

    async def get_test_run(self, test_id: str) -> Optional[Dict[str, Any]]:
        """Get test run by ID."""
        try:
            client = await self._get_client()
            doc_ref = client.collection("test_runs").document(test_id)

            doc = await doc_ref.get(retry=self._retry)
            if doc.exists:
                return doc.to_dict()
            return None

        except Exception as e:
            logger.error(f"Failed to get test run: {str(e)}")
            return None

    async def get_user_test_history(
        self,
        user_id: str,
        limit: int = 100,
        offset: int = 0,
        status: Optional[str] = None,
    ) -> List[Dict[str, Any]]:
        """Get test history for a user."""
        try:
            client = await self._get_client()
            query = client.collection("test_runs").where(
                filter=FieldFilter("user_id", "==", user_id)
            )

            if status:
                query = query.where(filter=FieldFilter("status", "==", status))

            query = query.order_by("created_at", direction=firestore.Query.DESCENDING)
            query = query.limit(limit).offset(offset)

            docs = await query.get(retry=self._retry)

            return [doc.to_dict() for doc in docs]

        except Exception as e:
            logger.error(f"Failed to get test history: {str(e)}")
            return []

    # Webhook Management

    async def create_webhook(
        self,
        webhook_id: str,
        user_id: str,
        tenant_id: str,
        webhook_data: Dict[str, Any],
    ) -> Dict[str, Any]:
        """Create a webhook."""
        try:
            client = await self._get_client()

            webhook_data.update(
                {
                    "webhook_id": webhook_id,
                    "user_id": user_id,
                    "tenant_id": tenant_id,
                    "created_at": datetime.now(timezone.utc),
                    "updated_at": datetime.now(timezone.utc),
                }
            )

            await (
                client.collection("webhooks")
                .document(webhook_id)
                .set(webhook_data, retry=self._retry)
            )

            logger.info(f"Created webhook: {webhook_id}")
            return webhook_data

        except Exception as e:
            logger.error(f"Failed to create webhook: {str(e)}")
            raise

    async def get_webhook(self, webhook_id: str) -> Optional[Dict[str, Any]]:
        """Get webhook by ID."""
        try:
            client = await self._get_client()
            doc_ref = client.collection("webhooks").document(webhook_id)

            doc = await doc_ref.get(retry=self._retry)
            if doc.exists:
                return doc.to_dict()
            return None

        except Exception as e:
            logger.error(f"Failed to get webhook: {str(e)}")
            return None

    async def list_user_webhooks(
        self, user_id: str, active_only: bool = True, limit: int = 100, offset: int = 0
    ) -> List[Dict[str, Any]]:
        """List webhooks for a user."""
        try:
            client = await self._get_client()
            query = client.collection("webhooks").where(
                filter=FieldFilter("user_id", "==", user_id)
            )

            if active_only:
                query = query.where(filter=FieldFilter("active", "==", True))

            query = query.order_by("created_at", direction=firestore.Query.DESCENDING)
            query = query.limit(limit).offset(offset)

            docs = await query.get(retry=self._retry)

            return [doc.to_dict() for doc in docs]

        except Exception as e:
            logger.error(f"Failed to list webhooks: {str(e)}")
            return []

    async def update_webhook(
        self, webhook_id: str, update_data: Dict[str, Any]
    ) -> Optional[Dict[str, Any]]:
        """Update webhook."""
        try:
            client = await self._get_client()
            webhook_ref = client.collection("webhooks").document(webhook_id)

            update_data["updated_at"] = datetime.now(timezone.utc)

            await webhook_ref.update(update_data, retry=self._retry)

            # Return updated webhook
            doc = await webhook_ref.get(retry=self._retry)
            if doc.exists:
                return doc.to_dict()
            return None

        except Exception as e:
            logger.error(f"Failed to update webhook: {str(e)}")
            return None

    async def delete_webhook(self, webhook_id: str) -> bool:
        """Delete webhook."""
        try:
            client = await self._get_client()
            webhook_ref = client.collection("webhooks").document(webhook_id)

            await webhook_ref.delete(retry=self._retry)
            logger.info(f"Deleted webhook: {webhook_id}")

            return True

        except Exception as e:
            logger.error(f"Failed to delete webhook: {str(e)}")
            return False

    async def get_webhooks_for_event(self, event_type: str) -> List[Dict[str, Any]]:
        """Get all webhooks subscribed to an event type."""
        try:
            client = await self._get_client()

            # Query webhooks that have this event in their events array
            query = (
                client.collection("webhooks")
                .where(filter=FieldFilter("events", "array_contains", event_type))
                .where(filter=FieldFilter("active", "==", True))
            )

            docs = await query.get(retry=self._retry)

            return [doc.to_dict() for doc in docs]

        except Exception as e:
            logger.error(f"Failed to get webhooks for event: {str(e)}")
            return []

    async def get_webhook_deliveries(
        self, webhook_id: str, limit: int = 100, offset: int = 0
    ) -> List[Dict[str, Any]]:
        """Get webhook delivery history."""
        try:
            client = await self._get_client()
            query = client.collection("webhook_deliveries").where(
                filter=FieldFilter("webhook_id", "==", webhook_id)
            )

            query = query.order_by("created_at", direction=firestore.Query.DESCENDING)
            query = query.limit(limit).offset(offset)

            docs = await query.get(retry=self._retry)

            return [doc.to_dict() for doc in docs]

        except Exception as e:
            logger.error(f"Failed to get webhook deliveries: {str(e)}")
            return []

    # Report Management

    async def create_report_request(
        self,
        report_id: str,
        user_id: str,
        tenant_id: str,
        report_type: str,
        time_range: str,
        format: str,
        status: str = "pending",
    ) -> Dict[str, Any]:
        """Create a report request."""
        try:
            client = await self._get_client()

            report_data = {
                "report_id": report_id,
                "user_id": user_id,
                "tenant_id": tenant_id,
                "report_type": report_type,
                "time_range": time_range,
                "format": format,
                "status": status,
                "created_at": datetime.now(timezone.utc),
                "updated_at": datetime.now(timezone.utc),
            }

            await (
                client.collection("report_requests")
                .document(report_id)
                .set(report_data, retry=self._retry)
            )

            logger.info(f"Created report request: {report_id}")
            return report_data

        except Exception as e:
            logger.error(f"Failed to create report request: {str(e)}")
            raise

    async def get_report_status(self, report_id: str) -> Optional[Dict[str, Any]]:
        """Get report status."""
        try:
            client = await self._get_client()
            doc_ref = client.collection("report_requests").document(report_id)

            doc = await doc_ref.get(retry=self._retry)
            if doc.exists:
                return doc.to_dict()
            return None

        except Exception as e:
            logger.error(f"Failed to get report status: {str(e)}")
            return None

    # API Request Logging

    async def log_api_request(
        self,
        user_id: str,
        path: str,
        method: str,
        client_host: Optional[str] = None,
        user_agent: Optional[str] = None,
    ) -> str:
        """Log an API request for analytics."""
        try:
            client = await self._get_client()

            request_data = {
                "user_id": user_id,
                "path": path,
                "method": method,
                "client_host": client_host,
                "user_agent": user_agent,
                "timestamp": datetime.now(timezone.utc),
            }

            doc_ref = await client.collection("api_requests").add(
                request_data, retry=self._retry
            )

            return doc_ref[1].id

        except Exception as e:
            logger.error(f"Failed to log API request: {str(e)}")
            return ""
