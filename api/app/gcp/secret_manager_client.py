"""
Secret Manager client for secure credential management.

This module provides async Secret Manager operations for loading API credentials,
managing secrets, and implementing secret rotation with caching.
"""

import asyncio
import json
from datetime import datetime, timedelta, timezone
from typing import Any, Dict, List, Optional, Union
from enum import Enum

from google.api_core import retry, exceptions
from google.cloud.secretmanager_v1 import SecretManagerServiceAsyncClient
from google.cloud.secretmanager_v1.types import (
    Secret,
    SecretVersion,
    SecretPayload,
    Replication,
)

from app.config import settings
from app.utils.logger import setup_logger

logger = setup_logger(__name__)


class SecretType(str, Enum):
    """Types of secrets managed."""

    VIRTUOSO_API_KEY = "virtuoso-api-key"
    DATABASE_PASSWORD = "database-password"
    JWT_SECRET = "jwt-secret"
    ENCRYPTION_KEY = "encryption-key"
    OAUTH_CLIENT_SECRET = "oauth-client-secret"
    WEBHOOK_SECRET = "webhook-secret"
    SERVICE_ACCOUNT_KEY = "service-account-key"


class SecretManagerClient:
    """
    Async Secret Manager client for credential management.
    """

    def __init__(self):
        """Initialize Secret Manager client."""
        self.project_id = settings.GCP_PROJECT_ID

        # Secret naming convention
        self.SECRET_PREFIX = "virtuoso"

        # Initialize client
        self._client: Optional[SecretManagerServiceAsyncClient] = None

        # Cache for secrets
        self._cache: Dict[str, Dict[str, Any]] = {}
        self._cache_ttl = timedelta(minutes=settings.SECRET_CACHE_TTL_MINUTES)

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

        logger.info(f"Initialized Secret Manager client for project: {self.project_id}")

    async def _get_client(self) -> SecretManagerServiceAsyncClient:
        """Get or create Secret Manager client."""
        if self._client is None:
            self._client = SecretManagerServiceAsyncClient()
        return self._client

    async def close(self):
        """Close Secret Manager client connection."""
        if self._client:
            await self._client.transport.close()
            self._client = None

    def _get_secret_id(
        self, secret_type: SecretType, suffix: Optional[str] = None
    ) -> str:
        """Generate secret ID from type and optional suffix."""
        parts = [self.SECRET_PREFIX, secret_type.value]
        if suffix:
            parts.append(suffix)
        return "-".join(parts)

    def _get_secret_name(self, secret_id: str) -> str:
        """Get full secret resource name."""
        return f"projects/{self.project_id}/secrets/{secret_id}"

    def _get_version_name(
        self, secret_id: str, version: Union[str, int] = "latest"
    ) -> str:
        """Get full secret version resource name."""
        return f"projects/{self.project_id}/secrets/{secret_id}/versions/{version}"

    # Secret Management

    async def create_secret(
        self,
        secret_type: SecretType,
        secret_data: Union[str, Dict[str, Any]],
        suffix: Optional[str] = None,
        labels: Optional[Dict[str, str]] = None,
        auto_rotation_days: Optional[int] = None,
    ) -> Secret:
        """Create a new secret."""
        try:
            client = await self._get_client()
            secret_id = self._get_secret_id(secret_type, suffix)

            # Convert dict to JSON string if needed
            if isinstance(secret_data, dict):
                secret_data = json.dumps(secret_data)

            # Create secret metadata
            secret = Secret(
                replication=Replication(automatic=Replication.Automatic()),
                labels={
                    "type": secret_type.value,
                    "managed-by": "virtuoso-api",
                    **(labels or {}),
                },
            )

            # Create secret
            parent = f"projects/{self.project_id}"
            created_secret = await client.create_secret(
                parent=parent,
                secret_id=secret_id,
                secret=secret,
                retry=self._retry,
            )

            # Add initial version
            await self.add_secret_version(secret_type, secret_data, suffix)

            logger.info(f"Created secret: {secret_id}")

            # Set up rotation if specified
            if auto_rotation_days:
                await self._setup_rotation(secret_id, auto_rotation_days)

            return created_secret

        except exceptions.AlreadyExists:
            logger.warning(f"Secret already exists: {secret_id}")
            return await self.get_secret(secret_type, suffix)
        except Exception as e:
            logger.error(f"Failed to create secret: {str(e)}")
            raise

    async def get_secret(
        self, secret_type: SecretType, suffix: Optional[str] = None
    ) -> Optional[Secret]:
        """Get secret metadata."""
        try:
            client = await self._get_client()
            secret_id = self._get_secret_id(secret_type, suffix)
            secret_name = self._get_secret_name(secret_id)

            secret = await client.get_secret(
                name=secret_name,
                retry=self._retry,
            )

            return secret

        except exceptions.NotFound:
            return None
        except Exception as e:
            logger.error(f"Failed to get secret: {str(e)}")
            raise

    async def get_secret_value(
        self,
        secret_type: SecretType,
        suffix: Optional[str] = None,
        version: Union[str, int] = "latest",
        use_cache: bool = True,
    ) -> Optional[Union[str, Dict[str, Any]]]:
        """Get secret value with caching."""
        secret_id = self._get_secret_id(secret_type, suffix)
        cache_key = f"{secret_id}:{version}"

        # Check cache
        if use_cache and cache_key in self._cache:
            cache_entry = self._cache[cache_key]
            if cache_entry["expires_at"] > datetime.now(timezone.utc):
                return cache_entry["value"]
            else:
                del self._cache[cache_key]

        try:
            client = await self._get_client()
            version_name = self._get_version_name(secret_id, version)

            response = await client.access_secret_version(
                name=version_name,
                retry=self._retry,
            )

            # Decode payload
            payload = response.payload.data.decode("UTF-8")

            # Try to parse as JSON
            try:
                value = json.loads(payload)
            except json.JSONDecodeError:
                value = payload

            # Cache the value
            if use_cache:
                self._cache[cache_key] = {
                    "value": value,
                    "expires_at": datetime.now(timezone.utc) + self._cache_ttl,
                }

            return value

        except exceptions.NotFound:
            logger.warning(f"Secret not found: {secret_id}")
            return None
        except Exception as e:
            logger.error(f"Failed to get secret value: {str(e)}")
            raise

    async def add_secret_version(
        self,
        secret_type: SecretType,
        secret_data: Union[str, Dict[str, Any]],
        suffix: Optional[str] = None,
    ) -> SecretVersion:
        """Add a new version to an existing secret."""
        try:
            client = await self._get_client()
            secret_id = self._get_secret_id(secret_type, suffix)
            secret_name = self._get_secret_name(secret_id)

            # Convert dict to JSON string if needed
            if isinstance(secret_data, dict):
                secret_data = json.dumps(secret_data)

            # Create payload
            payload = SecretPayload(data=secret_data.encode("UTF-8"))

            # Add version
            version = await client.add_secret_version(
                parent=secret_name,
                payload=payload,
                retry=self._retry,
            )

            # Clear cache for this secret
            self._clear_cache_for_secret(secret_id)

            logger.info(f"Added version to secret: {secret_id}")
            return version

        except Exception as e:
            logger.error(f"Failed to add secret version: {str(e)}")
            raise

    async def delete_secret(
        self, secret_type: SecretType, suffix: Optional[str] = None
    ) -> bool:
        """Delete a secret and all its versions."""
        try:
            client = await self._get_client()
            secret_id = self._get_secret_id(secret_type, suffix)
            secret_name = self._get_secret_name(secret_id)

            await client.delete_secret(
                name=secret_name,
                retry=self._retry,
            )

            # Clear cache
            self._clear_cache_for_secret(secret_id)

            logger.info(f"Deleted secret: {secret_id}")
            return True

        except exceptions.NotFound:
            return False
        except Exception as e:
            logger.error(f"Failed to delete secret: {str(e)}")
            raise

    # Specific Secret Loaders

    async def load_virtuoso_credentials(self) -> Dict[str, Any]:
        """Load Virtuoso API credentials."""
        creds = await self.get_secret_value(SecretType.VIRTUOSO_API_KEY, use_cache=True)

        if not creds:
            # Fallback to environment variables
            creds = {
                "api_key": settings.VIRTUOSO_API_KEY,
                "base_url": settings.VIRTUOSO_BASE_URL,
                "org_id": settings.VIRTUOSO_ORG_ID,
            }

            # Store in Secret Manager for future use
            if creds["api_key"]:
                await self.create_secret(SecretType.VIRTUOSO_API_KEY, creds)

        return creds

    async def load_api_keys(self) -> List[str]:
        """Load valid API keys for authentication."""
        keys_data = await self.get_secret_value(
            SecretType.OAUTH_CLIENT_SECRET, suffix="api-keys", use_cache=True
        )

        if not keys_data:
            # Fallback to environment variables
            return settings.API_KEYS

        if isinstance(keys_data, dict):
            return keys_data.get("keys", [])
        elif isinstance(keys_data, str):
            return [k.strip() for k in keys_data.split(",") if k.strip()]
        else:
            return []

    async def load_jwt_secret(self) -> str:
        """Load JWT secret for token signing."""
        secret = await self.get_secret_value(SecretType.JWT_SECRET, use_cache=True)

        if not secret:
            # Fallback to settings
            secret = settings.SECRET_KEY

            # Store for future use
            if secret and secret != "your-secret-key-here":
                await self.create_secret(SecretType.JWT_SECRET, secret)

        return str(secret)

    # Secret Rotation

    async def rotate_secret(
        self,
        secret_type: SecretType,
        new_value: Union[str, Dict[str, Any]],
        suffix: Optional[str] = None,
        disable_old_versions: bool = True,
    ) -> SecretVersion:
        """Rotate a secret by adding a new version."""
        # Add new version
        new_version = await self.add_secret_version(secret_type, new_value, suffix)

        if disable_old_versions:
            await self._disable_old_versions(secret_type, suffix)

        logger.info(f"Rotated secret: {self._get_secret_id(secret_type, suffix)}")
        return new_version

    async def _disable_old_versions(
        self,
        secret_type: SecretType,
        suffix: Optional[str] = None,
        keep_versions: int = 2,
    ):
        """Disable old secret versions, keeping the latest N versions."""
        try:
            client = await self._get_client()
            secret_id = self._get_secret_id(secret_type, suffix)
            secret_name = self._get_secret_name(secret_id)

            # List all versions
            versions = []
            async for version in client.list_secret_versions(
                parent=secret_name,
                retry=self._retry,
            ):
                if version.state == SecretVersion.State.ENABLED:
                    versions.append(version)

            # Sort by create time (newest first)
            versions.sort(key=lambda v: v.create_time.seconds, reverse=True)

            # Disable old versions
            for version in versions[keep_versions:]:
                await client.disable_secret_version(
                    name=version.name,
                    retry=self._retry,
                )
                logger.info(f"Disabled secret version: {version.name}")

        except Exception as e:
            logger.error(f"Failed to disable old versions: {str(e)}")

    async def _setup_rotation(self, secret_id: str, rotation_days: int):
        """Set up automatic rotation for a secret."""
        # This would typically integrate with Cloud Scheduler or Cloud Functions
        # For now, we'll log the intent
        logger.info(
            f"Rotation setup requested for {secret_id} every {rotation_days} days. "
            "Implement Cloud Scheduler integration for automatic rotation."
        )

    # Cache Management

    def _clear_cache_for_secret(self, secret_id: str):
        """Clear all cached versions of a secret."""
        keys_to_remove = [
            key for key in self._cache.keys() if key.startswith(f"{secret_id}:")
        ]

        for key in keys_to_remove:
            del self._cache[key]

    def clear_cache(self):
        """Clear all cached secrets."""
        self._cache.clear()
        logger.info("Cleared secret cache")

    # Bulk Operations

    async def load_all_secrets(
        self, secret_types: Optional[List[SecretType]] = None
    ) -> Dict[str, Any]:
        """Load multiple secrets at once."""
        if secret_types is None:
            secret_types = list(SecretType)

        secrets = {}

        # Load secrets concurrently
        tasks = []
        for secret_type in secret_types:
            task = self.get_secret_value(secret_type, use_cache=True)
            tasks.append((secret_type, task))

        for secret_type, task in tasks:
            try:
                value = await task
                if value is not None:
                    secrets[secret_type.value] = value
            except Exception as e:
                logger.error(f"Failed to load {secret_type}: {str(e)}")

        return secrets

    # Health Check

    async def health_check(self) -> Dict[str, Any]:
        """Check Secret Manager connectivity and permissions."""
        try:
            client = await self._get_client()

            # Try to list secrets (tests read permission)
            parent = f"projects/{self.project_id}"
            secrets_count = 0

            async for _ in client.list_secrets(
                parent=parent,
                page_size=1,
                retry=self._retry,
            ):
                secrets_count += 1
                break

            return {
                "healthy": True,
                "project_id": self.project_id,
                "can_list_secrets": secrets_count > 0,
                "cache_size": len(self._cache),
            }

        except Exception as e:
            return {
                "healthy": False,
                "error": str(e),
                "project_id": self.project_id,
            }
