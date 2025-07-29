"""
BigQuery client for analytics queries and data warehouse operations.

This module provides async BigQuery operations for running analytics queries,
exporting data, and managing datasets for the Virtuoso API CLI.
"""

import asyncio
from datetime import datetime
from typing import Any, Dict, List, Optional
from google.cloud import bigquery
from google.cloud.bigquery import Client, QueryJobConfig
from google.api_core import retry, exceptions

from app.config import settings
from app.utils.logger import setup_logger

logger = setup_logger(__name__)


class BigQueryClient:
    """
    Async BigQuery client for analytics and data warehouse operations.
    """

    def __init__(self):
        """Initialize BigQuery client."""
        self.project_id = settings.GCP_PROJECT_ID
        self.location = settings.GCP_LOCATION

        # Initialize client (will be created on first use)
        self._client: Optional[Client] = None

        # Dataset names
        self.ANALYTICS_DATASET = "analytics"
        self.WAREHOUSE_DATASET = "warehouse"

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

        logger.info(f"Initialized BigQuery client for project: {self.project_id}")

    def _get_client(self) -> Client:
        """Get or create BigQuery client."""
        if self._client is None:
            self._client = Client(project=self.project_id, location=self.location)
        return self._client

    async def execute_query(
        self,
        query: str,
        parameters: Optional[List[Dict[str, Any]]] = None,
        use_query_cache: bool = True,
        timeout: float = 300.0,
    ) -> List[Dict[str, Any]]:
        """
        Execute a BigQuery SQL query.

        Args:
            query: SQL query string
            parameters: Query parameters for parameterized queries
            use_query_cache: Whether to use query cache
            timeout: Query timeout in seconds

        Returns:
            List of result rows as dictionaries
        """
        try:
            client = self._get_client()

            # Configure query
            job_config = QueryJobConfig(
                use_query_cache=use_query_cache, query_parameters=parameters or []
            )

            # Run query
            logger.info(f"Executing BigQuery query: {query[:100]}...")
            query_job = client.query(query, job_config=job_config)

            # Wait for results with timeout
            results = await asyncio.wait_for(
                asyncio.get_event_loop().run_in_executor(None, query_job.result),
                timeout=timeout,
            )

            # Convert results to list of dicts
            rows = []
            for row in results:
                rows.append(dict(row))

            logger.info(f"Query returned {len(rows)} rows")
            return rows

        except asyncio.TimeoutError:
            logger.error(f"Query timed out after {timeout} seconds")
            raise
        except Exception as e:
            logger.error(f"Failed to execute query: {str(e)}")
            raise

    async def insert_rows(
        self,
        dataset_id: str,
        table_id: str,
        rows: List[Dict[str, Any]],
        auto_create_table: bool = True,
    ) -> bool:
        """
        Insert rows into a BigQuery table.

        Args:
            dataset_id: Dataset ID
            table_id: Table ID
            rows: List of row dictionaries
            auto_create_table: Whether to auto-create table if it doesn't exist

        Returns:
            True if successful
        """
        try:
            client = self._get_client()
            table_ref = client.dataset(dataset_id).table(table_id)

            # Check if table exists
            try:
                table = client.get_table(table_ref)
            except exceptions.NotFound:
                if auto_create_table and rows:
                    # Infer schema from first row
                    table = await self.create_table_from_rows(
                        dataset_id, table_id, rows[0]
                    )
                else:
                    raise ValueError(f"Table {dataset_id}.{table_id} does not exist")

            # Insert rows
            errors = await asyncio.get_event_loop().run_in_executor(
                None, client.insert_rows_json, table, rows
            )

            if errors:
                logger.error(f"Failed to insert rows: {errors}")
                return False

            logger.info(f"Inserted {len(rows)} rows into {dataset_id}.{table_id}")
            return True

        except Exception as e:
            logger.error(f"Failed to insert rows: {str(e)}")
            raise

    async def create_table_from_rows(
        self, dataset_id: str, table_id: str, sample_row: Dict[str, Any]
    ) -> bigquery.Table:
        """
        Create a table with schema inferred from sample row.

        Args:
            dataset_id: Dataset ID
            table_id: Table ID
            sample_row: Sample row to infer schema

        Returns:
            Created table
        """
        try:
            client = self._get_client()

            # Infer schema
            schema = []
            for key, value in sample_row.items():
                field_type = "STRING"  # Default

                if isinstance(value, bool):
                    field_type = "BOOLEAN"
                elif isinstance(value, int):
                    field_type = "INTEGER"
                elif isinstance(value, float):
                    field_type = "FLOAT"
                elif isinstance(value, datetime):
                    field_type = "TIMESTAMP"
                elif isinstance(value, dict):
                    field_type = "RECORD"
                elif isinstance(value, list):
                    field_type = "REPEATED"

                schema.append(bigquery.SchemaField(key, field_type))

            # Create table
            table_ref = client.dataset(dataset_id).table(table_id)
            table = bigquery.Table(table_ref, schema=schema)

            table = await asyncio.get_event_loop().run_in_executor(
                None, client.create_table, table
            )

            logger.info(f"Created table {dataset_id}.{table_id}")
            return table

        except Exception as e:
            logger.error(f"Failed to create table: {str(e)}")
            raise

    async def export_to_gcs(
        self, query: str, gcs_uri: str, format: str = "CSV", compression: str = "GZIP"
    ) -> str:
        """
        Export query results to Google Cloud Storage.

        Args:
            query: SQL query to export
            gcs_uri: GCS destination URI (gs://bucket/path)
            format: Export format (CSV, JSON, AVRO, PARQUET)
            compression: Compression type (NONE, GZIP)

        Returns:
            Job ID
        """
        try:
            client = self._get_client()

            # Configure export job
            job_config = bigquery.ExtractJobConfig(
                destination_format=format, compression=compression
            )

            # Create temporary table from query
            temp_table_id = f"temp_export_{datetime.now().strftime('%Y%m%d_%H%M%S')}"
            temp_table_ref = client.dataset(self.WAREHOUSE_DATASET).table(temp_table_id)

            # Run query to create temp table
            query_job_config = bigquery.QueryJobConfig(
                destination=temp_table_ref, write_disposition="WRITE_TRUNCATE"
            )

            query_job = client.query(query, job_config=query_job_config)
            await asyncio.get_event_loop().run_in_executor(None, query_job.result)

            # Export temp table to GCS
            extract_job = client.extract_table(
                temp_table_ref, gcs_uri, job_config=job_config
            )

            # Wait for export to complete
            await asyncio.get_event_loop().run_in_executor(None, extract_job.result)

            # Clean up temp table
            client.delete_table(temp_table_ref)

            logger.info(f"Exported query results to {gcs_uri}")
            return extract_job.job_id

        except Exception as e:
            logger.error(f"Failed to export to GCS: {str(e)}")
            raise

    async def create_dataset_if_not_exists(
        self, dataset_id: str, description: Optional[str] = None
    ) -> bool:
        """
        Create a dataset if it doesn't exist.

        Args:
            dataset_id: Dataset ID
            description: Dataset description

        Returns:
            True if created, False if already exists
        """
        try:
            client = self._get_client()
            dataset_ref = client.dataset(dataset_id)

            try:
                client.get_dataset(dataset_ref)
                return False  # Already exists
            except exceptions.NotFound:
                # Create dataset
                dataset = bigquery.Dataset(dataset_ref)
                dataset.location = self.location
                if description:
                    dataset.description = description

                await asyncio.get_event_loop().run_in_executor(
                    None, client.create_dataset, dataset
                )

                logger.info(f"Created dataset {dataset_id}")
                return True

        except Exception as e:
            logger.error(f"Failed to create dataset: {str(e)}")
            raise

    async def get_table_schema(
        self, dataset_id: str, table_id: str
    ) -> Optional[List[Dict[str, Any]]]:
        """
        Get table schema.

        Args:
            dataset_id: Dataset ID
            table_id: Table ID

        Returns:
            Schema as list of field dictionaries
        """
        try:
            client = self._get_client()
            table_ref = client.dataset(dataset_id).table(table_id)

            table = await asyncio.get_event_loop().run_in_executor(
                None, client.get_table, table_ref
            )

            schema = []
            for field in table.schema:
                schema.append(
                    {
                        "name": field.name,
                        "type": field.field_type,
                        "mode": field.mode,
                        "description": field.description,
                    }
                )

            return schema

        except exceptions.NotFound:
            return None
        except Exception as e:
            logger.error(f"Failed to get table schema: {str(e)}")
            raise
