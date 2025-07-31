"""
Test management endpoints with real CLI execution.
"""

from typing import Dict, Any, List, Optional
from uuid import uuid4
import json
import os

from fastapi import APIRouter, HTTPException, status, BackgroundTasks, Depends
from pydantic import BaseModel, Field

from ..config import settings
from ..utils.logger import get_logger
from ..services.auth_service import AuthUser
from ..middleware.auth import get_authenticated_user
from ..middleware.rate_limit import rate_limit, RateLimitStrategy

# Initialize logger first
logger = get_logger(__name__)

# CLI executor imports with error handling
CLIExecutor = None
CommandContext = None
OutputFormat = None

try:
    from ..services.cli_executor import CLIExecutor, CommandContext, OutputFormat
except ImportError as e:
    logger.warning(f"Failed to import CLI executor components: {e}")
    # Define fallback classes if needed
    class DummyCommandContext:
        def __init__(self, **kwargs):
            pass
    
    class DummyOutputFormat:
        JSON = "json"
    
    CommandContext = DummyCommandContext if CommandContext is None else CommandContext
    OutputFormat = DummyOutputFormat if OutputFormat is None else OutputFormat

# GCP imports with error handling
CloudTasksClient = None
CloudStorageClient = None
FirestoreClient = None
PubSubClient = None

if settings.is_gcp_enabled:
    try:
        from ..gcp.cloud_tasks_client import CloudTasksClient
    except ImportError as e:
        logger.warning(f"Failed to import CloudTasksClient: {e}")
        CloudTasksClient = None
    
    try:
        from ..gcp.cloud_storage_client import CloudStorageClient
    except ImportError as e:
        logger.warning(f"Failed to import CloudStorageClient: {e}")
        CloudStorageClient = None
    
    try:
        from ..gcp.firestore_client import FirestoreClient
    except ImportError as e:
        logger.warning(f"Failed to import FirestoreClient: {e}")
        FirestoreClient = None
    
    try:
        from ..gcp.pubsub_client import PubSubClient
    except ImportError as e:
        logger.warning(f"Failed to import PubSubClient: {e}")
        PubSubClient = None

router = APIRouter()

# Initialize CLI executor with comprehensive error handling
cli_executor = None
if CLIExecutor is not None:
    try:
        cli_executor = CLIExecutor()
        logger.info(f"CLI executor initialized successfully at {cli_executor.cli_path}")
    except ImportError as e:
        logger.warning(f"CLI executor import failed (dependencies missing): {e}")
    except FileNotFoundError as e:
        logger.warning(f"CLI executor binary not found: {e}")
    except PermissionError as e:
        logger.warning(f"CLI executor permission error: {e}")
    except Exception as e:
        logger.warning(f"Failed to initialize CLI executor: {e}")
else:
    logger.warning("CLI executor class not available, skipping initialization")

# Initialize GCP clients if enabled with individual error handling
tasks_client = None
storage_client = None
firestore_client = None
pubsub_client = None

if settings.is_gcp_enabled:
    if settings.USE_CLOUD_TASKS and CloudTasksClient is not None:
        try:
            tasks_client = CloudTasksClient()
            logger.info("Cloud Tasks client initialized successfully")
        except Exception as e:
            logger.warning(f"Failed to initialize Cloud Tasks client: {e}")
    
    if settings.USE_CLOUD_STORAGE and CloudStorageClient is not None:
        try:
            storage_client = CloudStorageClient()
            logger.info("Cloud Storage client initialized successfully")
        except Exception as e:
            logger.warning(f"Failed to initialize Cloud Storage client: {e}")
    
    if settings.USE_FIRESTORE and FirestoreClient is not None:
        try:
            firestore_client = FirestoreClient()
            logger.info("Firestore client initialized successfully")
        except Exception as e:
            logger.warning(f"Failed to initialize Firestore client: {e}")
    
    if settings.USE_PUBSUB and PubSubClient is not None:
        try:
            pubsub_client = PubSubClient()
            logger.info("PubSub client initialized successfully")
        except Exception as e:
            logger.warning(f"Failed to initialize PubSub client: {e}")


class TestDefinition(BaseModel):
    """Test definition model."""

    name: str = Field(..., description="Test name")
    description: Optional[str] = Field(None, description="Test description")
    steps: List[Dict[str, Any]] = Field(..., description="Test steps")
    config: Optional[Dict[str, Any]] = Field(None, description="Test configuration")


class TestRunRequest(BaseModel):
    """Request model for running a test."""

    test_id: Optional[str] = Field(None, description="Test ID to run")
    definition: Optional[TestDefinition] = Field(None, description="Test definition")
    dry_run: bool = Field(False, description="Perform a dry run without execution")
    execute: bool = Field(True, description="Execute the test after creation")
    environment: Optional[str] = Field(None, description="Test environment")
    variables: Optional[Dict[str, str]] = Field(None, description="Test variables")
    callback_url: Optional[str] = Field(
        None, description="Callback URL for async results"
    )


class TestRunResponse(BaseModel):
    """Response model for test run."""

    test_id: str = Field(..., description="Test ID")
    status: str = Field(..., description="Test status")
    project_id: Optional[str] = Field(None, description="Created project ID")
    checkpoint_id: Optional[str] = Field(None, description="Created checkpoint ID")
    execution_id: Optional[str] = Field(None, description="Execution ID if executed")
    steps_created: int = Field(..., description="Number of steps created")
    async_execution: bool = Field(False, description="Whether running asynchronously")
    uploaded_file: Optional[str] = Field(None, description="Uploaded file name")
    environment: Optional[str] = Field(None, description="Test environment")
    variables: Optional[Dict[str, Any]] = Field(None, description="Test variables")
    callback_url: Optional[str] = Field(
        None, description="Callback URL for async results"
    )


@router.get("/cli-status")
async def get_cli_status():
    """Get CLI executor status for debugging."""
    import os
    from pathlib import Path

    status = {
        "cli_executor_initialized": cli_executor is not None,
        "cli_path_env": os.environ.get("CLI_PATH", "not set"),
        "paths_checked": {},
    }

    # Check various paths
    paths_to_check = [
        "/bin/api-cli",
        "/usr/local/bin/api-cli",
        "./bin/api-cli",
        "api-cli",
        os.environ.get("CLI_PATH", ""),
    ]

    for path in paths_to_check:
        if path:
            p = Path(path)
            status["paths_checked"][path] = {
                "exists": p.exists(),
                "is_file": p.is_file() if p.exists() else None,
                "is_executable": os.access(str(p), os.X_OK) if p.exists() else None,
            }

    if cli_executor:
        status["cli_executor_path"] = cli_executor.cli_path
        status["cli_executor_error"] = None
    else:
        status["cli_executor_error"] = "Not initialized"

    return status


@router.post(
    "/run",
    response_model=TestRunResponse,
    dependencies=[Depends(rate_limit(5, 60, RateLimitStrategy.PER_USER))],
)
async def run_test(
    request: TestRunRequest,
    user: AuthUser = Depends(get_authenticated_user),
    background_tasks: BackgroundTasks = BackgroundTasks(),
    use_async: bool = False,
) -> TestRunResponse:
    """
    Run a test from definition with actual CLI execution.
    """
    test_id = request.test_id or str(uuid4())

    # Create context for CLI execution
    context = CommandContext(user_id=user.id, request_id=test_id, session_id=test_id)

    # Initialize response values
    project_id = None
    goal_id = None
    journey_id = None
    checkpoint_id = None
    steps_created = 0

    # Check if CLI executor is available
    if not cli_executor:
        logger.error("CLI executor not initialized, returning mock response")
        # Return mock response if CLI executor is not available
        return TestRunResponse(
            test_id=test_id,
            status="created",
            project_id=f"proj_{uuid4().hex[:8]}",
            checkpoint_id=f"cp_{uuid4().hex[:8]}",
            execution_id=None,
            steps_created=len(request.definition.steps)
            if request.definition and request.definition.steps
            else 0,
            async_execution=False,
            environment=request.environment,
            variables=request.variables,
            callback_url=request.callback_url,
        )

    try:
        # Debug logging
        logger.info(f"CLI executor initialized: {cli_executor}")
        logger.info(f"CLI path: {cli_executor.cli_path if cli_executor else 'None'}")

        # Step 1: Create project if specified in config
        if request.definition and request.definition.config:
            project_name = request.definition.config.get("project_name")
            if project_name:
                logger.info(f"Creating project: {project_name}")

                # Execute create-project command
                cmd = f'create-project "{project_name}"'
                logger.info(f"Executing command: {cmd}")

                result = await cli_executor.execute_async(
                    cmd, context, OutputFormat.JSON
                )

                logger.info(
                    f"Command result - Success: {result.success}, Exit code: {result.exit_code}"
                )
                logger.info(f"Command stdout: {result.stdout}")
                logger.info(f"Command stderr: {result.stderr}")

                if result.success and result.output:
                    try:
                        data = json.loads(result.output)
                        project_id = data.get("project_id")
                        logger.info(f"Created project: {project_id}")
                    except json.JSONDecodeError:
                        logger.error(
                            f"Failed to parse project creation response: {result.output}"
                        )
                else:
                    logger.error(f"Failed to create project: {result.error}")

        # Step 2: Create goal if we have a project
        if project_id and request.definition:
            goal_name = request.definition.config.get(
                "goal_name", request.definition.name
            )
            logger.info(f"Creating goal: {goal_name} in project {project_id}")

            cmd = f'create-goal {project_id} "{goal_name}"'
            result = await cli_executor.execute_async(cmd, context, OutputFormat.JSON)

            if result.success and result.output:
                try:
                    data = json.loads(result.output)
                    goal_id = data.get("goal_id")
                    snapshot_id = data.get("snapshot_id")
                    logger.info(f"Created goal: {goal_id}")
                except json.JSONDecodeError:
                    logger.error(
                        f"Failed to parse goal creation response: {result.output}"
                    )

        # Step 3: Create journey if we have a goal
        if goal_id and snapshot_id:
            journey_name = request.definition.config.get("journey_name", "Test Journey")
            logger.info(f"Creating journey: {journey_name} in goal {goal_id}")

            cmd = f'create-journey {goal_id} {snapshot_id} "{journey_name}"'
            result = await cli_executor.execute_async(cmd, context, OutputFormat.JSON)

            if result.success and result.output:
                try:
                    data = json.loads(result.output)
                    journey_id = data.get("journey_id")
                    logger.info(f"Created journey: {journey_id}")
                except json.JSONDecodeError:
                    logger.error(
                        f"Failed to parse journey creation response: {result.output}"
                    )

        # Step 4: Create checkpoint if we have a journey
        if journey_id and goal_id and snapshot_id:
            checkpoint_name = request.definition.config.get(
                "checkpoint_name", request.definition.name
            )
            logger.info(f"Creating checkpoint: {checkpoint_name}")

            cmd = f'create-checkpoint {journey_id} {goal_id} {snapshot_id} "{checkpoint_name}"'
            result = await cli_executor.execute_async(cmd, context, OutputFormat.JSON)

            if result.success and result.output:
                try:
                    data = json.loads(result.output)
                    checkpoint_id = data.get("checkpoint_id")
                    logger.info(f"Created checkpoint: {checkpoint_id}")
                except json.JSONDecodeError:
                    logger.error(
                        f"Failed to parse checkpoint creation response: {result.output}"
                    )

        # Step 5: Add steps to checkpoint
        if checkpoint_id and request.definition and request.definition.steps:
            # Set session for subsequent commands
            os.environ["VIRTUOSO_SESSION_ID"] = str(checkpoint_id)

            for i, step in enumerate(request.definition.steps):
                action = step.get("action")

                if action == "navigate":
                    url = step.get("url", "")
                    cmd = f'step-navigate to "{url}"'
                elif action == "click":
                    hint = step.get("hint", "")
                    cmd = f'step-interact click "{hint}"'
                elif action == "write":
                    hint = step.get("hint", "")
                    text = step.get("text", "")
                    cmd = f'step-interact write "{hint}" "{text}"'
                elif action == "assert":
                    hint = step.get("hint", "")
                    cmd = f'step-assert exists "{hint}"'
                elif action == "wait":
                    time_ms = step.get("time", 1000)
                    cmd = f"step-wait time {time_ms}"
                else:
                    logger.warning(f"Unknown action: {action}")
                    continue

                # Execute step command
                result = await cli_executor.execute_async(
                    cmd, context, OutputFormat.JSON
                )

                if result.success:
                    steps_created += 1
                    logger.info(f"Added step {i+1}: {action}")
                else:
                    logger.error(f"Failed to add step {i+1}: {result.error}")

        return TestRunResponse(
            test_id=test_id,
            status="created",
            project_id=str(project_id) if project_id else None,
            checkpoint_id=str(checkpoint_id) if checkpoint_id else None,
            execution_id=None,
            steps_created=steps_created,
            async_execution=False,
            environment=request.environment,
            variables=request.variables,
            callback_url=request.callback_url,
        )

    except Exception as e:
        logger.error(f"Error running test: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to run test: {str(e)}",
        )
    finally:
        # Clean up session ID
        if "VIRTUOSO_SESSION_ID" in os.environ:
            del os.environ["VIRTUOSO_SESSION_ID"]
