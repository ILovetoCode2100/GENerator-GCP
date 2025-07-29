"""
Cloud Storage client for file storage and archival.

This module provides async Cloud Storage operations for storing test results,
command logs, and serving static assets with proper lifecycle management.
"""

import asyncio
import json
import mimetypes
from datetime import datetime, timedelta, timezone
from typing import Any, Dict, List, Optional, Union
from pathlib import Path

from google.api_core import retry, exceptions
from google.cloud import storage
from google.cloud.storage import Blob, Bucket
import aiofiles

from app.config import settings
from app.utils.logger import setup_logger

logger = setup_logger(__name__)


class StorageClass(str):
    """Cloud Storage classes."""

    STANDARD = "STANDARD"
    NEARLINE = "NEARLINE"
    COLDLINE = "COLDLINE"
    ARCHIVE = "ARCHIVE"


class CloudStorageClient:
    """
    Async Cloud Storage client for file operations.
    """

    def __init__(self):
        """Initialize Cloud Storage client."""
        self.project_id = settings.GCP_PROJECT_ID

        # Bucket names
        self.TEST_RESULTS_BUCKET = f"{self.project_id}-virtuoso-test-results"
        self.COMMAND_LOGS_BUCKET = f"{self.project_id}-virtuoso-command-logs"
        self.STATIC_ASSETS_BUCKET = f"{self.project_id}-virtuoso-static-assets"
        self.ARCHIVE_BUCKET = f"{self.project_id}-virtuoso-archive"

        # Initialize client
        self._client: Optional[storage.Client] = None

        # Upload settings
        self.chunk_size = 1024 * 1024  # 1MB chunks
        self.max_file_size = 100 * 1024 * 1024  # 100MB

        # Retry configuration
        self._retry = retry.Retry(
            initial=0.1,
            maximum=60.0,
            multiplier=2.0,
            timeout=300.0,
            predicate=retry.if_exception_type(
                exceptions.GoogleAPIError,
                asyncio.TimeoutError,
            ),
        )

        logger.info(f"Initialized Cloud Storage client for project: {self.project_id}")

    def _get_client(self) -> storage.Client:
        """Get or create Cloud Storage client."""
        if self._client is None:
            self._client = storage.Client(project=self.project_id)
        return self._client

    def close(self):
        """Close Cloud Storage client connection."""
        if self._client:
            self._client.close()
            self._client = None

    # Bucket Management

    async def create_bucket(
        self,
        bucket_name: str,
        location: str = "US",
        storage_class: StorageClass = StorageClass.STANDARD,
        lifecycle_rules: Optional[List[Dict[str, Any]]] = None,
        versioning: bool = False,
        public: bool = False,
    ) -> Bucket:
        """Create a new storage bucket."""
        try:
            client = self._get_client()

            bucket = client.bucket(bucket_name)
            bucket.location = location
            bucket.storage_class = storage_class

            # Set versioning
            if versioning:
                bucket.versioning_enabled = True

            # Create bucket
            bucket = await asyncio.to_thread(
                client.create_bucket, bucket, retry=self._retry
            )

            # Set lifecycle rules
            if lifecycle_rules:
                bucket.lifecycle_rules = lifecycle_rules
                await asyncio.to_thread(bucket.patch)

            # Make public if requested
            if public:
                await self._make_bucket_public(bucket)

            logger.info(f"Created bucket: {bucket_name}")
            return bucket

        except exceptions.Conflict:
            logger.info(f"Bucket already exists: {bucket_name}")
            return await self.get_bucket(bucket_name)
        except Exception as e:
            logger.error(f"Failed to create bucket: {str(e)}")
            raise

    async def get_bucket(self, bucket_name: str) -> Optional[Bucket]:
        """Get bucket object."""
        try:
            client = self._get_client()
            bucket = await asyncio.to_thread(
                client.get_bucket, bucket_name, retry=self._retry
            )
            return bucket

        except exceptions.NotFound:
            return None
        except Exception as e:
            logger.error(f"Failed to get bucket: {str(e)}")
            raise

    async def delete_bucket(self, bucket_name: str, force: bool = False) -> bool:
        """Delete a bucket."""
        try:
            bucket = await self.get_bucket(bucket_name)
            if not bucket:
                return False

            if force:
                # Delete all objects first
                await self.delete_all_objects(bucket_name)

            await asyncio.to_thread(bucket.delete, retry=self._retry)
            logger.info(f"Deleted bucket: {bucket_name}")
            return True

        except Exception as e:
            logger.error(f"Failed to delete bucket: {str(e)}")
            raise

    async def _make_bucket_public(self, bucket: Bucket):
        """Make a bucket publicly readable."""
        policy = bucket.get_iam_policy(requested_policy_version=3)
        policy.bindings.append(
            {
                "role": "roles/storage.objectViewer",
                "members": {"allUsers"},
            }
        )
        await asyncio.to_thread(bucket.set_iam_policy, policy)

    # File Operations

    async def upload_file(
        self,
        bucket_name: str,
        source_file_path: Union[str, Path],
        destination_blob_name: Optional[str] = None,
        content_type: Optional[str] = None,
        metadata: Optional[Dict[str, str]] = None,
        public: bool = False,
    ) -> str:
        """Upload a file to Cloud Storage."""
        try:
            bucket = await self.get_bucket(bucket_name)
            if not bucket:
                raise ValueError(f"Bucket not found: {bucket_name}")

            source_path = Path(source_file_path)
            if not source_path.exists():
                raise FileNotFoundError(f"File not found: {source_file_path}")

            # Check file size
            file_size = source_path.stat().st_size
            if file_size > self.max_file_size:
                raise ValueError(f"File too large: {file_size} bytes")

            # Determine blob name
            blob_name = destination_blob_name or source_path.name

            # Determine content type
            if not content_type:
                content_type, _ = mimetypes.guess_type(str(source_path))
                content_type = content_type or "application/octet-stream"

            # Create blob
            blob = bucket.blob(blob_name)

            # Set metadata
            if metadata:
                blob.metadata = metadata

            # Upload file
            async with aiofiles.open(source_path, "rb") as f:
                content = await f.read()
                await asyncio.to_thread(
                    blob.upload_from_string,
                    content,
                    content_type=content_type,
                    retry=self._retry,
                )

            # Make public if requested
            if public:
                await asyncio.to_thread(blob.make_public)

            logger.info(f"Uploaded file: {blob_name} to {bucket_name}")
            return blob.public_url if public else f"gs://{bucket_name}/{blob_name}"

        except Exception as e:
            logger.error(f"Failed to upload file: {str(e)}")
            raise

    async def upload_from_memory(
        self,
        bucket_name: str,
        content: Union[str, bytes],
        blob_name: str,
        content_type: str = "text/plain",
        metadata: Optional[Dict[str, str]] = None,
        public: bool = False,
    ) -> str:
        """Upload content from memory."""
        try:
            bucket = await self.get_bucket(bucket_name)
            if not bucket:
                raise ValueError(f"Bucket not found: {bucket_name}")

            # Convert string to bytes if needed
            if isinstance(content, str):
                content = content.encode("utf-8")

            # Check size
            if len(content) > self.max_file_size:
                raise ValueError(f"Content too large: {len(content)} bytes")

            # Create blob
            blob = bucket.blob(blob_name)

            # Set metadata
            if metadata:
                blob.metadata = metadata

            # Upload content
            await asyncio.to_thread(
                blob.upload_from_string,
                content,
                content_type=content_type,
                retry=self._retry,
            )

            # Make public if requested
            if public:
                await asyncio.to_thread(blob.make_public)

            logger.info(f"Uploaded content to: {blob_name}")
            return blob.public_url if public else f"gs://{bucket_name}/{blob_name}"

        except Exception as e:
            logger.error(f"Failed to upload content: {str(e)}")
            raise

    async def download_file(
        self, bucket_name: str, blob_name: str, destination_file_path: Union[str, Path]
    ) -> Path:
        """Download a file from Cloud Storage."""
        try:
            bucket = await self.get_bucket(bucket_name)
            if not bucket:
                raise ValueError(f"Bucket not found: {bucket_name}")

            blob = bucket.blob(blob_name)
            if not await asyncio.to_thread(blob.exists):
                raise FileNotFoundError(f"Blob not found: {blob_name}")

            destination_path = Path(destination_file_path)
            destination_path.parent.mkdir(parents=True, exist_ok=True)

            # Download file
            content = await asyncio.to_thread(blob.download_as_bytes, retry=self._retry)

            async with aiofiles.open(destination_path, "wb") as f:
                await f.write(content)

            logger.info(f"Downloaded file: {blob_name} to {destination_path}")
            return destination_path

        except Exception as e:
            logger.error(f"Failed to download file: {str(e)}")
            raise

    async def download_to_memory(self, bucket_name: str, blob_name: str) -> bytes:
        """Download file content to memory."""
        try:
            bucket = await self.get_bucket(bucket_name)
            if not bucket:
                raise ValueError(f"Bucket not found: {bucket_name}")

            blob = bucket.blob(blob_name)
            if not await asyncio.to_thread(blob.exists):
                raise FileNotFoundError(f"Blob not found: {blob_name}")

            content = await asyncio.to_thread(blob.download_as_bytes, retry=self._retry)

            logger.info(f"Downloaded {len(content)} bytes from: {blob_name}")
            return content

        except Exception as e:
            logger.error(f"Failed to download to memory: {str(e)}")
            raise

    async def delete_file(self, bucket_name: str, blob_name: str) -> bool:
        """Delete a file from Cloud Storage."""
        try:
            bucket = await self.get_bucket(bucket_name)
            if not bucket:
                return False

            blob = bucket.blob(blob_name)
            await asyncio.to_thread(blob.delete, retry=self._retry)

            logger.info(f"Deleted file: {blob_name}")
            return True

        except exceptions.NotFound:
            return False
        except Exception as e:
            logger.error(f"Failed to delete file: {str(e)}")
            raise

    async def list_files(
        self,
        bucket_name: str,
        prefix: Optional[str] = None,
        delimiter: Optional[str] = None,
        max_results: Optional[int] = None,
    ) -> List[Dict[str, Any]]:
        """List files in a bucket."""
        try:
            bucket = await self.get_bucket(bucket_name)
            if not bucket:
                return []

            blobs = await asyncio.to_thread(
                bucket.list_blobs,
                prefix=prefix,
                delimiter=delimiter,
                max_results=max_results,
                retry=self._retry,
            )

            files = []
            for blob in blobs:
                files.append(
                    {
                        "name": blob.name,
                        "size": blob.size,
                        "content_type": blob.content_type,
                        "created": blob.time_created,
                        "updated": blob.updated,
                        "metadata": blob.metadata,
                    }
                )

            return files

        except Exception as e:
            logger.error(f"Failed to list files: {str(e)}")
            raise

    # Specialized Operations

    async def store_test_result(
        self,
        test_id: str,
        result_data: Dict[str, Any],
        user_id: Optional[str] = None,
        project_id: Optional[str] = None,
    ) -> str:
        """Store test execution result."""
        timestamp = datetime.now(timezone.utc).strftime("%Y%m%d_%H%M%S")

        # Build path
        path_parts = ["test-results"]
        if project_id:
            path_parts.append(f"project-{project_id}")
        if user_id:
            path_parts.append(f"user-{user_id}")
        path_parts.extend([f"test-{test_id}", f"{timestamp}.json"])

        blob_name = "/".join(path_parts)

        # Add metadata
        metadata = {
            "test_id": test_id,
            "timestamp": timestamp,
        }
        if user_id:
            metadata["user_id"] = user_id
        if project_id:
            metadata["project_id"] = project_id

        # Store result
        content = json.dumps(result_data, indent=2)
        return await self.upload_from_memory(
            self.TEST_RESULTS_BUCKET,
            content,
            blob_name,
            content_type="application/json",
            metadata=metadata,
        )

    async def store_command_log(
        self,
        command: str,
        log_data: str,
        user_id: Optional[str] = None,
        session_id: Optional[str] = None,
    ) -> str:
        """Store command execution log."""
        timestamp = datetime.now(timezone.utc).strftime("%Y%m%d_%H%M%S")

        # Build path
        path_parts = ["command-logs"]
        if user_id:
            path_parts.append(f"user-{user_id}")
        if session_id:
            path_parts.append(f"session-{session_id}")
        path_parts.append(f"{command}_{timestamp}.log")

        blob_name = "/".join(path_parts)

        # Add metadata
        metadata = {
            "command": command,
            "timestamp": timestamp,
        }
        if user_id:
            metadata["user_id"] = user_id
        if session_id:
            metadata["session_id"] = session_id

        # Store log
        return await self.upload_from_memory(
            self.COMMAND_LOGS_BUCKET,
            log_data,
            blob_name,
            content_type="text/plain",
            metadata=metadata,
        )

    async def archive_old_data(
        self, source_bucket: str, days_old: int = 30, delete_after_archive: bool = True
    ) -> int:
        """Archive old data to archive bucket."""
        try:
            cutoff_date = datetime.now(timezone.utc) - timedelta(days=days_old)
            archived_count = 0

            # List old files
            files = await self.list_files(source_bucket)

            for file_info in files:
                if file_info["created"] < cutoff_date:
                    # Copy to archive
                    source_blob_name = file_info["name"]
                    dest_blob_name = f"{source_bucket}/{source_blob_name}"

                    # Download and re-upload (no direct copy in async)
                    content = await self.download_to_memory(
                        source_bucket, source_blob_name
                    )

                    await self.upload_from_memory(
                        self.ARCHIVE_BUCKET,
                        content,
                        dest_blob_name,
                        content_type=file_info["content_type"],
                        metadata=file_info["metadata"],
                    )

                    # Delete from source if requested
                    if delete_after_archive:
                        await self.delete_file(source_bucket, source_blob_name)

                    archived_count += 1

            logger.info(f"Archived {archived_count} files from {source_bucket}")
            return archived_count

        except Exception as e:
            logger.error(f"Failed to archive data: {str(e)}")
            raise

    async def generate_signed_url(
        self,
        bucket_name: str,
        blob_name: str,
        expiration_minutes: int = 60,
        method: str = "GET",
    ) -> str:
        """Generate a signed URL for temporary access."""
        try:
            bucket = await self.get_bucket(bucket_name)
            if not bucket:
                raise ValueError(f"Bucket not found: {bucket_name}")

            blob = bucket.blob(blob_name)

            url = await asyncio.to_thread(
                blob.generate_signed_url,
                expiration=timedelta(minutes=expiration_minutes),
                method=method,
                version="v4",
            )

            return url

        except Exception as e:
            logger.error(f"Failed to generate signed URL: {str(e)}")
            raise

    async def delete_all_objects(self, bucket_name: str) -> int:
        """Delete all objects in a bucket."""
        try:
            files = await self.list_files(bucket_name)

            for file_info in files:
                await self.delete_file(bucket_name, file_info["name"])

            return len(files)

        except Exception as e:
            logger.error(f"Failed to delete all objects: {str(e)}")
            raise

    # Lifecycle Management

    async def set_lifecycle_rules(self, bucket_name: str, rules: List[Dict[str, Any]]):
        """Set lifecycle rules for automatic data management."""
        try:
            bucket = await self.get_bucket(bucket_name)
            if not bucket:
                raise ValueError(f"Bucket not found: {bucket_name}")

            bucket.lifecycle_rules = rules
            await asyncio.to_thread(bucket.patch)

            logger.info(f"Set lifecycle rules for: {bucket_name}")

        except Exception as e:
            logger.error(f"Failed to set lifecycle rules: {str(e)}")
            raise

    async def create_default_buckets(self):
        """Create all default buckets with appropriate settings."""
        # Test results bucket
        await self.create_bucket(
            self.TEST_RESULTS_BUCKET,
            lifecycle_rules=[
                {
                    "action": {
                        "type": "SetStorageClass",
                        "storageClass": StorageClass.NEARLINE,
                    },
                    "condition": {"age": 30},
                }
            ],
        )

        # Command logs bucket
        await self.create_bucket(
            self.COMMAND_LOGS_BUCKET,
            lifecycle_rules=[{"action": {"type": "Delete"}, "condition": {"age": 90}}],
        )

        # Static assets bucket (public)
        await self.create_bucket(self.STATIC_ASSETS_BUCKET, public=True)

        # Archive bucket
        await self.create_bucket(
            self.ARCHIVE_BUCKET,
            storage_class=StorageClass.COLDLINE,
            lifecycle_rules=[
                {
                    "action": {
                        "type": "SetStorageClass",
                        "storageClass": StorageClass.ARCHIVE,
                    },
                    "condition": {"age": 365},
                }
            ],
        )

    # Additional convenience methods for API compatibility

    async def upload_json(
        self,
        bucket_name: str,
        blob_name: str,
        data: Dict[str, Any],
        metadata: Optional[Dict[str, str]] = None,
    ) -> str:
        """
        Upload JSON data to Cloud Storage.

        Args:
            bucket_name: Name of the bucket
            blob_name: Name of the blob
            data: JSON data to upload
            metadata: Optional metadata

        Returns:
            Public URL of the uploaded file
        """
        json_str = json.dumps(data, indent=2)
        return await self.upload_from_memory(
            bucket_name=bucket_name,
            blob_name=blob_name,
            data=json_str.encode(),
            content_type="application/json",
            metadata=metadata,
        )

    async def upload_bytes(
        self,
        bucket_name: str,
        blob_name: str,
        data: bytes,
        content_type: Optional[str] = None,
        metadata: Optional[Dict[str, str]] = None,
    ) -> str:
        """
        Upload raw bytes to Cloud Storage.

        Args:
            bucket_name: Name of the bucket
            blob_name: Name of the blob
            data: Bytes to upload
            content_type: Optional content type
            metadata: Optional metadata

        Returns:
            Public URL of the uploaded file
        """
        return await self.upload_from_memory(
            bucket_name=bucket_name,
            blob_name=blob_name,
            data=data,
            content_type=content_type or "application/octet-stream",
            metadata=metadata,
        )

    async def download_json(self, bucket_name: str, blob_name: str) -> Dict[str, Any]:
        """
        Download JSON data from Cloud Storage.

        Args:
            bucket_name: Name of the bucket
            blob_name: Name of the blob

        Returns:
            Parsed JSON data
        """
        data = await self.download_to_memory(bucket_name, blob_name)
        return json.loads(data.decode())

    async def download_as_text(
        self, bucket_name: str, blob_name: str, encoding: str = "utf-8"
    ) -> str:
        """
        Download file as text from Cloud Storage.

        Args:
            bucket_name: Name of the bucket
            blob_name: Name of the blob
            encoding: Text encoding

        Returns:
            File content as string
        """
        data = await self.download_to_memory(bucket_name, blob_name)
        return data.decode(encoding)

    async def list_blobs(
        self,
        bucket_name: str,
        prefix: Optional[str] = None,
        delimiter: Optional[str] = None,
        max_results: Optional[int] = None,
    ) -> List[Blob]:
        """
        List blobs in a bucket.

        Args:
            bucket_name: Name of the bucket
            prefix: Filter blobs by prefix
            delimiter: Delimiter for hierarchy
            max_results: Maximum results to return

        Returns:
            List of blob objects
        """
        try:
            client = self._get_client()
            bucket = client.bucket(bucket_name)

            blobs = []
            async for blob in self._list_blobs_async(
                bucket, prefix=prefix, delimiter=delimiter, max_results=max_results
            ):
                blobs.append(blob)

            return blobs

        except Exception as e:
            logger.error(f"Failed to list blobs: {str(e)}")
            raise

    async def _list_blobs_async(
        self,
        bucket: Bucket,
        prefix: Optional[str] = None,
        delimiter: Optional[str] = None,
        max_results: Optional[int] = None,
    ):
        """Async generator for listing blobs."""
        loop = asyncio.get_event_loop()

        # Get blob iterator
        blob_iter = bucket.list_blobs(
            prefix=prefix, delimiter=delimiter, max_results=max_results
        )

        # Convert to async iteration
        for blob in await loop.run_in_executor(None, list, blob_iter):
            yield blob
