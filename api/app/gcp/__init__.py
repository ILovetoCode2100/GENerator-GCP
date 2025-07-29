"""
Google Cloud Platform service integrations for Virtuoso API.

This module provides lazy-loaded GCP service clients with connection pooling
and proper error handling.
"""

from typing import Optional, TYPE_CHECKING

# Type checking imports
if TYPE_CHECKING:
    from .firestore_client import FirestoreClient
    from .cloud_tasks_client import CloudTasksClient
    from .pubsub_client import PubSubClient
    from .secret_manager_client import SecretManagerClient
    from .cloud_storage_client import CloudStorageClient
    from .monitoring_client import MonitoringClient

# Lazy-loaded client instances
_firestore_client: Optional["FirestoreClient"] = None
_cloud_tasks_client: Optional["CloudTasksClient"] = None
_pubsub_client: Optional["PubSubClient"] = None
_secret_manager_client: Optional["SecretManagerClient"] = None
_cloud_storage_client: Optional["CloudStorageClient"] = None
_monitoring_client: Optional["MonitoringClient"] = None


def get_firestore_client() -> "FirestoreClient":
    """Get or create Firestore client instance."""
    global _firestore_client
    if _firestore_client is None:
        from .firestore_client import FirestoreClient

        _firestore_client = FirestoreClient()
    return _firestore_client


def get_cloud_tasks_client() -> "CloudTasksClient":
    """Get or create Cloud Tasks client instance."""
    global _cloud_tasks_client
    if _cloud_tasks_client is None:
        from .cloud_tasks_client import CloudTasksClient

        _cloud_tasks_client = CloudTasksClient()
    return _cloud_tasks_client


def get_pubsub_client() -> "PubSubClient":
    """Get or create Pub/Sub client instance."""
    global _pubsub_client
    if _pubsub_client is None:
        from .pubsub_client import PubSubClient

        _pubsub_client = PubSubClient()
    return _pubsub_client


def get_secret_manager_client() -> "SecretManagerClient":
    """Get or create Secret Manager client instance."""
    global _secret_manager_client
    if _secret_manager_client is None:
        from .secret_manager_client import SecretManagerClient

        _secret_manager_client = SecretManagerClient()
    return _secret_manager_client


def get_cloud_storage_client() -> "CloudStorageClient":
    """Get or create Cloud Storage client instance."""
    global _cloud_storage_client
    if _cloud_storage_client is None:
        from .cloud_storage_client import CloudStorageClient

        _cloud_storage_client = CloudStorageClient()
    return _cloud_storage_client


def get_monitoring_client() -> "MonitoringClient":
    """Get or create Monitoring client instance."""
    global _monitoring_client
    if _monitoring_client is None:
        from .monitoring_client import MonitoringClient

        _monitoring_client = MonitoringClient()
    return _monitoring_client


# Export all client getters
__all__ = [
    "get_firestore_client",
    "get_cloud_tasks_client",
    "get_pubsub_client",
    "get_secret_manager_client",
    "get_cloud_storage_client",
    "get_monitoring_client",
]
