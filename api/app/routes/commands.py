"""
Command execution endpoints.

These endpoints handle the execution of individual Virtuoso CLI commands
with proper authentication and rate limiting.
"""

from typing import Dict, Any, List, Optional
from datetime import datetime
from uuid import uuid4

from fastapi import APIRouter, HTTPException, status, Depends, BackgroundTasks

from ..models.requests import ExecuteCommandRequest, BatchCommandRequest
from ..models.responses import (
    CommandExecutionResponse,
    BatchCommandResponse,
    BatchExecutionResult,
    StepResult,
    ErrorDetail,
    ResponseStatus,
    BaseResponse,
)
from ..services.cli_executor import CLIExecutor, OutputFormat, CommandExecutionError
from ..services.auth_service import AuthUser, Permission
from ..middleware.auth import get_authenticated_user, require_permissions
from ..middleware.rate_limit import rate_limit, RateLimitStrategy
from ..config import settings
from ..utils.logger import get_logger
from ..utils.monitoring import (
    command_execution_count,
    command_execution_duration,
    error_count,
)

# GCP imports
if settings.is_gcp_enabled:
    from ..gcp.firestore_client import FirestoreClient
    from ..gcp.cloud_tasks_client import CloudTasksClient
    from ..gcp.pubsub_client import PubSubClient

router = APIRouter()
logger = get_logger(__name__)

# Initialize CLI executor
cli_executor = CLIExecutor()

# Initialize GCP clients if enabled
firestore_client = None
tasks_client = None
pubsub_client = None

if settings.is_gcp_enabled:
    if settings.USE_FIRESTORE:
        firestore_client = FirestoreClient()
    if settings.USE_CLOUD_TASKS:
        tasks_client = CloudTasksClient()
    if settings.USE_PUBSUB:
        pubsub_client = PubSubClient()


# Command metadata for documentation
COMMAND_METADATA = {
    "step-navigate": {
        "description": "Navigation commands",
        "subcommands": [
            "to",
            "scroll-to-top",
            "scroll-to-bottom",
            "scroll-to-element",
            "scroll-to-position",
            "scroll-by",
            "scroll-up",
            "scroll-down",
        ],
        "permission": "write:tests",
    },
    "step-interact": {
        "description": "User interaction commands",
        "subcommands": [
            "click",
            "double-click",
            "right-click",
            "hover",
            "write",
            "key",
            "mouse",
            "select",
        ],
        "permission": "write:tests",
    },
    "step-assert": {
        "description": "Assertion commands",
        "subcommands": [
            "exists",
            "not-exists",
            "equals",
            "not-equals",
            "checked",
            "selected",
            "variable",
            "gt",
            "gte",
            "lt",
            "lte",
            "matches",
        ],
        "permission": "write:tests",
    },
    "step-window": {
        "description": "Window management commands",
        "subcommands": ["resize", "maximize", "switch"],
        "permission": "write:tests",
    },
    "step-data": {
        "description": "Data management commands",
        "subcommands": ["store", "cookie"],
        "permission": "write:tests",
    },
    "step-dialog": {
        "description": "Dialog handling commands",
        "subcommands": [
            "dismiss-alert",
            "dismiss-confirm",
            "dismiss-prompt",
            "dismiss-prompt-with-text",
        ],
        "permission": "write:tests",
    },
    "step-wait": {
        "description": "Wait operation commands",
        "subcommands": ["element", "time"],
        "permission": "write:tests",
    },
    "step-file": {
        "description": "File operation commands",
        "subcommands": ["upload", "upload-url"],
        "permission": "write:tests",
    },
    "step-misc": {
        "description": "Miscellaneous commands",
        "subcommands": ["comment", "execute"],
        "permission": "write:tests",
    },
    "library": {
        "description": "Library management commands",
        "subcommands": ["add", "get", "attach", "move-step", "remove-step", "update"],
        "permission": "write:library",
    },
}


@router.post(
    "/step/{command_group}/{subcommand}",
    response_model=CommandExecutionResponse,
    dependencies=[Depends(rate_limit(200, 60, RateLimitStrategy.PER_USER))],
)
async def execute_step_command(
    command_group: str,
    subcommand: str,
    command: ExecuteCommandRequest,
    user: AuthUser = Depends(get_authenticated_user),
    background_tasks: BackgroundTasks = BackgroundTasks(),
) -> CommandExecutionResponse:
    """
    Execute a single step command.

    This endpoint executes a single Virtuoso CLI step command with the provided
    parameters. Authentication and appropriate permissions are required.

    Args:
        command_group: Command group (e.g., "navigate", "interact", "assert")
        subcommand: Specific subcommand (e.g., "to", "click", "exists")
        command: Command parameters
        user: Authenticated user
        background_tasks: Background task queue

    Returns:
        Command execution result

    Raises:
        HTTPException: If command fails or user lacks permission
    """
    # Check permission for command group
    full_command = f"step-{command_group}"
    if full_command not in COMMAND_METADATA:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"Unknown command group: {command_group}",
        )

    # Verify user has permission
    required_permission = Permission(COMMAND_METADATA[full_command]["permission"])
    if not await user.has_permission(required_permission):
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail=f"Insufficient permissions for command group: {command_group}",
        )

    # Build CLI command
    cli_args = [full_command, subcommand]

    # Add checkpoint ID if provided
    if command.checkpoint_id:
        cli_args.append(command.checkpoint_id)

    # Add command-specific arguments
    cli_args.extend(command.args)

    # Add position if provided
    if command.position is not None:
        cli_args.append(str(command.position))

    # Start time for metrics
    start_time = datetime.utcnow()

    try:
        # Track command metrics
        command_execution_count.labels(
            command=command_group, subcommand=subcommand, status="started"
        ).inc()

        # Build command string
        command_str = " ".join(cli_args)

        # Create context
        from ..services.cli_executor import CommandContext

        context = CommandContext(
            request_id=getattr(user, "request_id", "unknown"),
            session_id=command.checkpoint_id,
            timeout=settings.CLI_TIMEOUT,
        )

        # Execute command
        result = await cli_executor.execute_async(
            command=command_str, context=context, output_format=OutputFormat.JSON
        )

        # Calculate execution time
        execution_time = (datetime.utcnow() - start_time).total_seconds()

        # Check if command succeeded
        if not result.success:
            raise CommandExecutionError(
                f"Command failed with exit code {result.exit_code}: {result.stderr}"
            )

        # Parse result output
        parsed_output = result.parsed_output or {}

        # Parse result to extract step information
        step_result = StepResult(
            step_id=parsed_output.get("step_id", "unknown"),
            checkpoint_id=command.checkpoint_id
            or parsed_output.get("checkpoint_id", ""),
            position=command.position or parsed_output.get("position", 0),
            type=f"{command_group.upper()}_{subcommand.upper()}",
            description=command.description,
            created_at=datetime.utcnow(),
        )

        # Log command execution
        background_tasks.add_task(
            log_command_execution,
            user=user,
            command=f"{full_command} {subcommand}",
            checkpoint_id=command.checkpoint_id,
            args=command.args,
            success=True,
            duration_ms=int(execution_time * 1000),
        )

        # Update metrics
        command_execution_count.labels(
            command=command_group, subcommand=subcommand, status="success"
        ).inc()
        command_execution_duration.labels(
            command=command_group, subcommand=subcommand
        ).observe(execution_time)

        return CommandExecutionResponse(
            status=ResponseStatus.SUCCESS,
            data=step_result,
            message="Command executed successfully",
            execution_time_ms=int(execution_time * 1000),
        )

    except Exception as e:
        logger.error(f"Command execution failed: {e}")

        # Calculate execution time
        execution_time = (datetime.utcnow() - start_time).total_seconds()

        # Update error metrics
        command_execution_count.labels(
            command=command_group, subcommand=subcommand, status="error"
        ).inc()
        command_execution_duration.labels(
            command=command_group, subcommand=subcommand
        ).observe(execution_time)
        error_count.labels(
            error_type=type(e).__name__,
            endpoint=f"/commands/step/{command_group}/{subcommand}",
        ).inc()

        # Log failed execution
        background_tasks.add_task(
            log_command_execution,
            user=user,
            command=f"{full_command} {subcommand}",
            checkpoint_id=command.checkpoint_id,
            args=command.args,
            success=False,
            error=str(e),
            duration_ms=int(execution_time * 1000),
        )

        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Command execution failed: {str(e)}",
        )


@router.post(
    "/batch",
    response_model=BatchCommandResponse,
    dependencies=[Depends(rate_limit(10, 60, RateLimitStrategy.PER_USER))],
)
async def execute_batch_commands(
    request: BatchCommandRequest,
    user: AuthUser = Depends(get_authenticated_user),
    background_tasks: BackgroundTasks = BackgroundTasks(),
    use_async: bool = False,
) -> BatchCommandResponse:
    """
    Execute multiple commands in batch.

    This endpoint executes multiple Virtuoso CLI commands in sequence.
    If stop_on_error is true, execution stops at the first error.

    Args:
        request: Batch command request
        user: Authenticated user
        background_tasks: Background task queue

    Returns:
        Batch execution result
    """
    # Check permissions for all commands
    for cmd in request.commands:
        command_group = cmd.command.split("-")[1] if "-" in cmd.command else cmd.command
        full_command = (
            f"step-{command_group}"
            if not cmd.command.startswith("step-")
            else cmd.command
        )

        if full_command in COMMAND_METADATA:
            required_permission = Permission(
                COMMAND_METADATA[full_command]["permission"]
            )
            if not await user.has_permission(required_permission):
                raise HTTPException(
                    status_code=status.HTTP_403_FORBIDDEN,
                    detail=f"Insufficient permissions for command: {cmd.command}",
                )

    # Check if we should use Cloud Tasks for async execution
    if (
        use_async
        and settings.USE_CLOUD_TASKS
        and tasks_client
        and len(request.commands) > 5
    ):
        # Create async batch task
        batch_id = f"batch_{uuid4().hex[:8]}"

        try:
            # Create task for batch processing
            task = await tasks_client.create_batch_task(
                commands=[
                    {
                        "command": cmd.command,
                        "subcommand": cmd.subcommand,
                        "args": cmd.args,
                        "checkpoint_id": cmd.checkpoint_id,
                        "position": cmd.position,
                        "description": cmd.description,
                    }
                    for cmd in request.commands
                ],
                batch_id=batch_id,
                user_id=user.user_id,
                stop_on_error=request.stop_on_error,
                callback_url=request.callback_url,
            )

            # Log batch creation
            if settings.USE_FIRESTORE and firestore_client:
                await firestore_client.log_command(
                    user_id=user.user_id,
                    command="batch",
                    args=[f"async_{batch_id}"],
                    result={"task_name": task.name, "batch_id": batch_id},
                )

            # Publish event
            if settings.USE_PUBSUB and pubsub_client:
                await pubsub_client.publish_event(
                    topic="command-events",
                    event_type="batch.created",
                    data={
                        "batch_id": batch_id,
                        "user_id": user.user_id,
                        "command_count": len(request.commands),
                        "async": True,
                    },
                )

            return BatchCommandResponse(
                status=ResponseStatus.SUCCESS,
                data=BatchExecutionResult(
                    total_commands=len(request.commands),
                    successful=0,
                    failed=0,
                    skipped=0,
                    results=[],
                    execution_time_ms=0,
                    batch_id=batch_id,
                    async_execution=True,
                ),
                message=f"Batch queued for async execution. Batch ID: {batch_id}",
            )

        except Exception as e:
            logger.error(f"Failed to create async batch task: {e}")
            # Fall back to synchronous execution

    # Execute commands synchronously
    results = []
    successful = 0
    failed = 0
    skipped = 0
    start_time = datetime.utcnow()

    for idx, cmd in enumerate(request.commands):
        try:
            # Build CLI arguments
            cli_args = cmd.command.split()
            if cmd.subcommand:
                cli_args.append(cmd.subcommand)
            cli_args.extend(cmd.args)

            # Build command string
            command_str = " ".join(cli_args)

            # Create context
            from ..services.cli_executor import CommandContext

            context = CommandContext(
                request_id=getattr(user, "request_id", "unknown"),
                session_id=cmd.checkpoint_id,
                timeout=settings.CLI_TIMEOUT,
            )

            # Execute command
            result = await cli_executor.execute_async(
                command=command_str, context=context, output_format=OutputFormat.JSON
            )

            # Check if command succeeded
            if not result.success:
                raise CommandExecutionError(
                    f"Command failed with exit code {result.exit_code}: {result.stderr}"
                )

            # Parse result output
            parsed_output = result.parsed_output or {}

            # Create step result
            step_result = StepResult(
                step_id=parsed_output.get("step_id", f"batch_{idx}"),
                checkpoint_id=cmd.checkpoint_id or "",
                position=cmd.position or idx,
                type=cmd.command.upper(),
                description=cmd.description,
                created_at=datetime.utcnow(),
            )

            results.append(step_result)
            successful += 1

        except Exception as e:
            logger.error(f"Batch command {idx} failed: {e}")

            # Create error detail
            error_detail = ErrorDetail(
                field=f"commands[{idx}]", message=str(e), code="execution_failed"
            )

            results.append(error_detail)
            failed += 1

            # Stop on error if requested
            if request.stop_on_error:
                skipped = len(request.commands) - idx - 1
                break

    # Calculate execution time
    execution_time_ms = int((datetime.utcnow() - start_time).total_seconds() * 1000)

    # Log batch execution
    background_tasks.add_task(
        log_batch_execution,
        user=user,
        total_commands=len(request.commands),
        successful=successful,
        failed=failed,
        skipped=skipped,
    )

    # Prepare response
    batch_result = BatchExecutionResult(
        total_commands=len(request.commands),
        successful=successful,
        failed=failed,
        skipped=skipped,
        results=results,
        execution_time_ms=execution_time_ms,
    )

    return BatchCommandResponse(
        status=ResponseStatus.SUCCESS if failed == 0 else ResponseStatus.PARTIAL,
        data=batch_result,
        message=f"Executed {successful} commands successfully",
        partial_success=failed > 0 and successful > 0,
    )


@router.get(
    "/list",
    response_model=BaseResponse[List[Dict[str, Any]]],
    dependencies=[Depends(require_permissions(Permission.READ_TESTS))],
)
async def list_commands(
    user: AuthUser = Depends(get_authenticated_user),
) -> BaseResponse[List[Dict[str, Any]]]:
    """
    List all available CLI commands.

    Returns a list of all available commands with their descriptions
    and required permissions.

    Returns:
        List of available commands
    """
    commands = []

    for cmd, metadata in COMMAND_METADATA.items():
        # Check if user has permission for this command
        required_permission = Permission(metadata["permission"])
        has_permission = await user.has_permission(required_permission)

        commands.append(
            {
                "command": cmd,
                "description": metadata["description"],
                "subcommands": metadata["subcommands"],
                "permission_required": metadata["permission"],
                "user_has_permission": has_permission,
            }
        )

    return BaseResponse(
        status=ResponseStatus.SUCCESS,
        data=commands,
        message=f"Found {len(commands)} available commands",
    )


@router.get(
    "/{command}/help",
    response_model=BaseResponse[Dict[str, Any]],
    dependencies=[Depends(require_permissions(Permission.READ_TESTS))],
)
async def get_command_help(
    command: str, user: AuthUser = Depends(get_authenticated_user)
) -> BaseResponse[Dict[str, Any]]:
    """
    Get detailed help for a specific command.

    Args:
        command: Command name (e.g., "step-navigate", "step-interact")

    Returns:
        Command help information
    """
    if command not in COMMAND_METADATA:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Command not found: {command}",
        )

    metadata = COMMAND_METADATA[command]

    # Get detailed help from CLI
    try:
        help_result = await cli_executor.get_command_help(command)

        return BaseResponse(
            status=ResponseStatus.SUCCESS,
            data={
                "command": command,
                "description": metadata["description"],
                "subcommands": metadata["subcommands"],
                "permission_required": metadata["permission"],
                "usage": help_result.get(
                    "usage", f"api-cli {command} <subcommand> [options]"
                ),
                "examples": help_result.get("examples", []),
                "options": help_result.get("options", []),
            },
            message="Command help retrieved successfully",
        )
    except Exception as e:
        logger.error(f"Failed to get command help: {e}")

        # Return basic help if CLI help fails
        return BaseResponse(
            status=ResponseStatus.SUCCESS,
            data={
                "command": command,
                "description": metadata["description"],
                "subcommands": metadata["subcommands"],
                "permission_required": metadata["permission"],
                "usage": f"api-cli {command} <subcommand> [options]",
            },
            message="Basic command help retrieved",
        )


# Add new endpoint for async command execution
@router.post(
    "/execute-async",
    response_model=BaseResponse[Dict[str, str]],
    dependencies=[Depends(rate_limit(50, 60, RateLimitStrategy.PER_USER))],
)
async def execute_command_async(
    command: ExecuteCommandRequest,
    user: AuthUser = Depends(get_authenticated_user),
    priority: str = "normal",
    delay_seconds: Optional[int] = None,
) -> BaseResponse[Dict[str, str]]:
    """
    Execute a command asynchronously using Cloud Tasks.

    Args:
        command: Command to execute
        user: Authenticated user
        priority: Task priority
        delay_seconds: Delay before execution

    Returns:
        Task information
    """
    if not settings.USE_CLOUD_TASKS or not tasks_client:
        raise HTTPException(
            status_code=status.HTTP_501_NOT_IMPLEMENTED,
            detail="Async execution not available. Cloud Tasks not enabled.",
        )

    try:
        # Create task ID
        task_id = f"cmd_{uuid4().hex[:12]}"

        # Parse command to get full details
        cmd_parts = command.command.split("-")
        if len(cmd_parts) >= 2:
            command_group = cmd_parts[1]
            full_command = command.command
        else:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST, detail="Invalid command format"
            )

        # Create async task
        task = await tasks_client.create_command_task(
            command=full_command,
            args=command.args,
            checkpoint_id=command.checkpoint_id,
            user_id=user.user_id,
            session_id=command.checkpoint_id,  # Use checkpoint as session
            priority=priority,
            delay_seconds=delay_seconds,
            task_id=task_id,
        )

        # Log async command creation
        if settings.USE_FIRESTORE and firestore_client:
            await firestore_client.log_command(
                user_id=user.user_id,
                command=full_command,
                args=command.args,
                checkpoint_id=command.checkpoint_id,
                result={"task_id": task_id, "task_name": task.name},
            )

        # Publish event
        if settings.USE_PUBSUB and pubsub_client:
            await pubsub_client.publish_event(
                topic="command-events",
                event_type="command.async_created",
                data={
                    "task_id": task_id,
                    "command": full_command,
                    "user_id": user.user_id,
                    "checkpoint_id": command.checkpoint_id,
                },
            )

        return BaseResponse(
            status=ResponseStatus.SUCCESS,
            data={
                "task_id": task_id,
                "task_name": task.name,
                "status": "queued",
                "priority": priority.value,
                "scheduled_for": task.schedule_time.ToDatetime().isoformat()
                if task.schedule_time
                else "immediate",
            },
            message="Command queued for async execution",
        )

    except Exception as e:
        logger.error(f"Failed to create async command: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to queue command: {str(e)}",
        )


# Helper functions for background tasks
async def log_command_execution(
    user: AuthUser,
    command: str,
    checkpoint_id: Optional[str] = None,
    args: Optional[List[str]] = None,
    success: bool = True,
    error: Optional[str] = None,
    duration_ms: Optional[int] = None,
):
    """Log command execution for auditing."""
    from ..services.auth_service import auth_service

    # Log to auth service
    await auth_service.audit_log(
        user=user,
        action="execute_command",
        resource=command,
        details={"success": success, "error": error, "duration_ms": duration_ms},
    )

    # Log to Firestore if enabled
    if settings.USE_FIRESTORE and firestore_client:
        try:
            await firestore_client.log_command(
                user_id=user.user_id,
                command=command,
                args=args or [],
                checkpoint_id=checkpoint_id,
                result={"success": success} if success else None,
                error=error,
                duration_ms=duration_ms,
            )
        except Exception as e:
            logger.error(f"Failed to log command to Firestore: {e}")

    # Publish event to Pub/Sub if enabled
    if settings.USE_PUBSUB and pubsub_client:
        try:
            await pubsub_client.publish_event(
                topic="command-events",
                event_type="command.executed",
                data={
                    "user_id": user.user_id,
                    "command": command,
                    "checkpoint_id": checkpoint_id,
                    "success": success,
                    "error": error,
                    "duration_ms": duration_ms,
                    "timestamp": datetime.utcnow().isoformat(),
                },
            )
        except Exception as e:
            logger.error(f"Failed to publish event to Pub/Sub: {e}")


async def log_batch_execution(
    user: AuthUser,
    total_commands: int,
    successful: int,
    failed: int,
    skipped: int,
    batch_id: Optional[str] = None,
):
    """Log batch execution for auditing."""
    from ..services.auth_service import auth_service

    # Log to auth service
    await auth_service.audit_log(
        user=user,
        action="execute_batch",
        resource="batch_commands",
        details={
            "total": total_commands,
            "successful": successful,
            "failed": failed,
            "skipped": skipped,
            "batch_id": batch_id,
        },
    )

    # Publish batch completion event if enabled
    if settings.USE_PUBSUB and pubsub_client:
        try:
            await pubsub_client.publish_event(
                topic="command-events",
                event_type="batch.completed",
                data={
                    "user_id": user.user_id,
                    "batch_id": batch_id,
                    "total": total_commands,
                    "successful": successful,
                    "failed": failed,
                    "skipped": skipped,
                    "timestamp": datetime.utcnow().isoformat(),
                },
            )
        except Exception as e:
            logger.error(f"Failed to publish batch event: {e}")
