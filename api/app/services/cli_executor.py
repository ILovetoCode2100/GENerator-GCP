"""
CLI Executor Service for Virtuoso API CLI.

This service handles the safe execution of CLI commands with proper
parsing, validation, timeout handling, and output processing.
"""

import asyncio
import json
import os
import re
import shlex
import subprocess
import tempfile
import threading
import time
import yaml
from concurrent.futures import (
    ThreadPoolExecutor,
    Future,
    TimeoutError as FutureTimeoutError,
)
from contextlib import contextmanager
from dataclasses import dataclass, field
from enum import Enum
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple, Union, AsyncIterator
import uuid
from datetime import datetime

from ..config import settings
from ..utils.logger import get_logger

# GCP imports
if settings.is_gcp_enabled:
    from ..gcp.cloud_tasks_client import CloudTasksClient
    from ..gcp.cloud_storage_client import CloudStorageClient
    from ..gcp.pubsub_client import PubSubClient


logger = get_logger(__name__)

# Initialize GCP clients if enabled
tasks_client = None
storage_client = None
pubsub_client = None

if settings.is_gcp_enabled:
    if settings.USE_CLOUD_TASKS:
        tasks_client = CloudTasksClient()
    if settings.USE_CLOUD_STORAGE:
        storage_client = CloudStorageClient()
    if settings.USE_PUBSUB:
        pubsub_client = PubSubClient()


class OutputFormat(Enum):
    """Supported output formats for CLI commands."""

    JSON = "json"
    YAML = "yaml"
    HUMAN = "human"
    AI = "ai"
    RAW = "raw"  # For commands that don't support formatting


class CommandError(Exception):
    """Base exception for command execution errors."""

    pass


class CommandValidationError(CommandError):
    """Raised when command validation fails."""

    pass


class CommandTimeoutError(CommandError):
    """Raised when command execution times out."""

    pass


class CommandExecutionError(CommandError):
    """Raised when command execution fails."""

    def __init__(self, message: str, exit_code: int, stderr: str = ""):
        super().__init__(message)
        self.exit_code = exit_code
        self.stderr = stderr


@dataclass
class CommandResult:
    """Result of a CLI command execution."""

    success: bool
    exit_code: int
    stdout: str
    stderr: str
    duration: float
    command: List[str]
    output_format: OutputFormat
    parsed_output: Optional[Union[Dict, List, str]] = None
    error: Optional[str] = None


@dataclass
class CommandContext:
    """Context for command execution."""

    request_id: str
    session_id: Optional[str] = None
    timeout: Optional[float] = None
    environment: Dict[str, str] = field(default_factory=dict)
    working_dir: Optional[str] = None
    stream_output: bool = False


class CLIExecutor:
    """
    Service for executing Virtuoso API CLI commands safely.

    Handles command parsing, validation, execution with timeout,
    output parsing, and concurrent execution support.
    """

    # Command pattern: api-cli <command> <subcommand> [checkpoint-id] <args...> [position]
    COMMAND_PATTERN = re.compile(r"^([\w-]+)(?:\s+([\w-]+))?(?:\s+(\d+))?(?:\s+(.*))?$")

    # Commands that require checkpoint ID
    CHECKPOINT_COMMANDS = {
        "step-assert",
        "step-interact",
        "step-navigate",
        "step-window",
        "step-data",
        "step-dialog",
        "step-wait",
        "step-file",
        "step-misc",
        "library",
    }

    # Exit code meanings
    EXIT_CODES = {
        0: "Success",
        1: "General error",
        2: "Command line parsing error",
        3: "Authentication error",
        4: "Configuration error",
        5: "Resource not found",
        6: "Validation error",
        7: "API error",
        8: "Timeout error",
        9: "Permission denied",
        127: "Command not found",
    }

    def __init__(
        self,
        cli_path: Optional[str] = None,
        default_timeout: float = 300.0,
        max_workers: int = 4,
    ):
        """
        Initialize CLI executor.

        Args:
            cli_path: Path to CLI binary (defaults to settings.CLI_PATH)
            default_timeout: Default command timeout in seconds
            max_workers: Maximum concurrent command executions
        """
        self.cli_path = cli_path or settings.CLI_PATH
        self.default_timeout = default_timeout
        self.executor = ThreadPoolExecutor(max_workers=max_workers)

        # Validate CLI binary exists and is executable
        self._validate_cli_binary()

        # Cache for CLI capabilities
        self._commands_cache: Optional[Dict[str, Any]] = None
        self._cache_lock = threading.Lock()

    def _validate_cli_binary(self) -> None:
        """Validate that CLI binary exists and is executable."""
        cli_path = Path(self.cli_path)

        # Log detailed information about the CLI path
        logger.info(f"Checking CLI binary at: {self.cli_path}")
        logger.info(f"Path exists: {cli_path.exists()}")
        logger.info(
            f"Path is file: {cli_path.is_file() if cli_path.exists() else 'N/A'}"
        )
        logger.info(
            f"Path is executable: {os.access(str(cli_path), os.X_OK) if cli_path.exists() else 'N/A'}"
        )

        # Also check alternative paths
        alternative_paths = [
            "/bin/api-cli",
            "/usr/local/bin/api-cli",
            "./bin/api-cli",
            "api-cli",
        ]
        for alt_path in alternative_paths:
            alt = Path(alt_path)
            if alt.exists():
                logger.info(f"Found CLI at alternative path: {alt_path}")

        if not cli_path.exists():
            raise CommandValidationError(f"CLI binary not found at: {self.cli_path}")

        if not cli_path.is_file():
            raise CommandValidationError(f"CLI path is not a file: {self.cli_path}")

        if not os.access(str(cli_path), os.X_OK):
            raise CommandValidationError(
                f"CLI binary is not executable: {self.cli_path}"
            )

        logger.info(f"CLI binary validated at: {self.cli_path}")

    def parse_command(
        self, command: str
    ) -> Tuple[str, Optional[str], Optional[str], List[str]]:
        """
        Parse a CLI command string into components.

        Args:
            command: Command string to parse

        Returns:
            Tuple of (command, subcommand, checkpoint_id, args)

        Raises:
            CommandValidationError: If command format is invalid
        """
        command = command.strip()
        if not command:
            raise CommandValidationError("Empty command")

        # Remove 'api-cli' prefix if present
        if command.startswith("api-cli "):
            command = command[8:]

        # Split command into parts
        parts = shlex.split(command)
        if not parts:
            raise CommandValidationError("No command specified")

        cmd = parts[0]
        subcommand = None
        checkpoint_id = None
        args = []

        # Parse based on command structure
        if len(parts) > 1:
            # Check if this is a command that uses subcommands
            if cmd in self.CHECKPOINT_COMMANDS:
                subcommand = parts[1] if len(parts) > 1 else None

                # Check for checkpoint ID (numeric third argument)
                if len(parts) > 2 and parts[2].isdigit():
                    checkpoint_id = parts[2]
                    args = parts[3:]
                else:
                    args = parts[2:]
            else:
                # Commands without subcommands
                args = parts[1:]

        return cmd, subcommand, checkpoint_id, args

    def validate_command(
        self,
        command: str,
        subcommand: Optional[str] = None,
        checkpoint_id: Optional[str] = None,
        context: Optional[CommandContext] = None,
    ) -> None:
        """
        Validate a command before execution.

        Args:
            command: Main command
            subcommand: Subcommand if applicable
            checkpoint_id: Checkpoint ID if required
            context: Command execution context

        Raises:
            CommandValidationError: If validation fails
        """
        # Check if command requires checkpoint ID
        if command in self.CHECKPOINT_COMMANDS:
            # Check session context or explicit checkpoint ID
            if not checkpoint_id and context:
                checkpoint_id = context.session_id

            if not checkpoint_id and command != "library":
                # Most step commands require checkpoint ID
                raise CommandValidationError(
                    f"Command '{command}' requires checkpoint ID. "
                    "Provide it explicitly or set VIRTUOSO_SESSION_ID"
                )

        # Validate specific command requirements
        if command == "step-window" and subcommand == "resize":
            # Resize requires specific format
            pass  # Validation done during execution

        if command == "run-test":
            # Requires test file
            pass  # Validation done during execution

    def build_command_args(
        self,
        command: str,
        subcommand: Optional[str] = None,
        checkpoint_id: Optional[str] = None,
        args: List[str] = None,
        output_format: OutputFormat = OutputFormat.JSON,
    ) -> List[str]:
        """
        Build complete command arguments list.

        Args:
            command: Main command
            subcommand: Subcommand if applicable
            checkpoint_id: Checkpoint ID if required
            args: Additional arguments
            output_format: Desired output format

        Returns:
            Complete command arguments list
        """
        cmd_args = [self.cli_path, command]

        if subcommand:
            cmd_args.append(subcommand)

        if checkpoint_id:
            cmd_args.append(checkpoint_id)

        if args:
            cmd_args.extend(args)

        # Add output format if not raw
        if output_format != OutputFormat.RAW:
            cmd_args.extend(["--output", output_format.value])

        return cmd_args

    def execute(
        self,
        command: str,
        context: Optional[CommandContext] = None,
        output_format: OutputFormat = OutputFormat.JSON,
    ) -> CommandResult:
        """
        Execute a CLI command synchronously.

        Args:
            command: Complete command string
            context: Execution context
            output_format: Desired output format

        Returns:
            CommandResult with execution details
        """
        start_time = time.time()
        context = context or CommandContext(request_id="sync")

        try:
            # Parse command
            cmd, subcmd, checkpoint_id, args = self.parse_command(command)

            # Validate command
            self.validate_command(cmd, subcmd, checkpoint_id, context)

            # Build command arguments
            cmd_args = self.build_command_args(
                cmd, subcmd, checkpoint_id, args, output_format
            )

            # Prepare environment
            env = os.environ.copy()
            env.update(settings.get_cli_env())
            env.update(context.environment)

            if context.session_id:
                env["VIRTUOSO_SESSION_ID"] = context.session_id

            # Execute command
            timeout = context.timeout or self.default_timeout

            logger.info(f"Executing command: {' '.join(cmd_args)}")

            process = subprocess.Popen(
                cmd_args,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                env=env,
                cwd=context.working_dir,
                text=True,
            )

            try:
                stdout, stderr = process.communicate(timeout=timeout)
                exit_code = process.returncode
            except subprocess.TimeoutExpired:
                process.kill()
                stdout, stderr = process.communicate()
                raise CommandTimeoutError(f"Command timed out after {timeout} seconds")

            duration = time.time() - start_time

            # Parse output if needed
            parsed_output = None
            if stdout and output_format != OutputFormat.RAW:
                parsed_output = self._parse_output(stdout, output_format)

            # Determine success
            success = exit_code == 0
            error_msg = None

            if not success:
                error_msg = self._get_error_message(exit_code, stderr)
                logger.error(f"Command failed with exit code {exit_code}: {error_msg}")

            return CommandResult(
                success=success,
                exit_code=exit_code,
                stdout=stdout,
                stderr=stderr,
                duration=duration,
                command=cmd_args,
                output_format=output_format,
                parsed_output=parsed_output,
                error=error_msg,
            )

        except CommandTimeoutError:
            raise
        except Exception as e:
            duration = time.time() - start_time
            logger.exception(f"Command execution failed: {e}")

            return CommandResult(
                success=False,
                exit_code=-1,
                stdout="",
                stderr=str(e),
                duration=duration,
                command=[command],
                output_format=output_format,
                error=str(e),
            )

    async def execute_async(
        self,
        command: str,
        context: Optional[CommandContext] = None,
        output_format: OutputFormat = OutputFormat.JSON,
    ) -> CommandResult:
        """
        Execute a CLI command asynchronously.

        Args:
            command: Complete command string
            context: Execution context
            output_format: Desired output format

        Returns:
            CommandResult with execution details
        """
        loop = asyncio.get_event_loop()

        # Run in thread pool to avoid blocking
        future = loop.run_in_executor(
            self.executor, self.execute, command, context, output_format
        )

        return await future

    async def execute_stream(
        self,
        command: str,
        context: Optional[CommandContext] = None,
        output_format: OutputFormat = OutputFormat.JSON,
    ) -> AsyncIterator[str]:
        """
        Execute a CLI command and stream output.

        Args:
            command: Complete command string
            context: Execution context
            output_format: Desired output format

        Yields:
            Output lines as they are produced
        """
        context = context or CommandContext(request_id="stream")

        # Parse and validate command
        cmd, subcmd, checkpoint_id, args = self.parse_command(command)
        self.validate_command(cmd, subcmd, checkpoint_id, context)

        # Build command arguments
        cmd_args = self.build_command_args(
            cmd, subcmd, checkpoint_id, args, output_format
        )

        # Prepare environment
        env = os.environ.copy()
        env.update(settings.get_cli_env())
        env.update(context.environment)

        if context.session_id:
            env["VIRTUOSO_SESSION_ID"] = context.session_id

        # Create subprocess
        process = await asyncio.create_subprocess_exec(
            *cmd_args,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE,
            env=env,
            cwd=context.working_dir,
        )

        # Stream output
        try:
            while True:
                line = await process.stdout.readline()
                if not line:
                    break
                yield line.decode().rstrip()

            # Wait for process to complete
            await process.wait()

            # Check for errors
            if process.returncode != 0:
                stderr = await process.stderr.read()
                error_msg = self._get_error_message(process.returncode, stderr.decode())
                yield f"ERROR: {error_msg}"

        except asyncio.CancelledError:
            # Handle cancellation
            process.kill()
            await process.wait()
            raise

    def execute_batch(
        self,
        commands: List[str],
        context: Optional[CommandContext] = None,
        output_format: OutputFormat = OutputFormat.JSON,
        max_concurrent: Optional[int] = None,
    ) -> List[CommandResult]:
        """
        Execute multiple commands concurrently.

        Args:
            commands: List of command strings
            context: Execution context (shared by all commands)
            output_format: Desired output format
            max_concurrent: Maximum concurrent executions

        Returns:
            List of CommandResult objects
        """
        max_concurrent = max_concurrent or self.executor._max_workers
        results = []

        # Submit all commands
        futures: List[Tuple[int, Future]] = []
        for i, cmd in enumerate(commands):
            future = self.executor.submit(self.execute, cmd, context, output_format)
            futures.append((i, future))

        # Collect results in order
        for i, future in futures:
            try:
                result = future.result(timeout=self.default_timeout)
                results.append(result)
            except FutureTimeoutError:
                # Create timeout result
                results.append(
                    CommandResult(
                        success=False,
                        exit_code=8,
                        stdout="",
                        stderr="Command timed out",
                        duration=self.default_timeout,
                        command=[commands[i]],
                        output_format=output_format,
                        error="Timeout",
                    )
                )
            except Exception as e:
                # Create error result
                results.append(
                    CommandResult(
                        success=False,
                        exit_code=-1,
                        stdout="",
                        stderr=str(e),
                        duration=0.0,
                        command=[commands[i]],
                        output_format=output_format,
                        error=str(e),
                    )
                )

        return results

    def _parse_output(
        self, output: str, format_type: OutputFormat
    ) -> Optional[Union[Dict, List, str]]:
        """
        Parse command output based on format type.

        Args:
            output: Raw command output
            format_type: Expected output format

        Returns:
            Parsed output or None if parsing fails
        """
        if not output.strip():
            return None

        try:
            if format_type == OutputFormat.JSON:
                return json.loads(output)
            elif format_type == OutputFormat.YAML:
                return yaml.safe_load(output)
            else:
                return output
        except Exception as e:
            logger.warning(f"Failed to parse output as {format_type.value}: {e}")
            return output

    def _get_error_message(self, exit_code: int, stderr: str) -> str:
        """
        Get user-friendly error message based on exit code.

        Args:
            exit_code: Process exit code
            stderr: Standard error output

        Returns:
            Error message
        """
        base_msg = self.EXIT_CODES.get(exit_code, f"Unknown error (code: {exit_code})")

        if stderr.strip():
            return f"{base_msg}: {stderr.strip()}"

        return base_msg

    @contextmanager
    def temporary_config(self, config_data: Dict[str, Any]):
        """
        Context manager for temporary CLI configuration.

        Args:
            config_data: Configuration data to write

        Yields:
            Path to temporary config file
        """
        with tempfile.NamedTemporaryFile(mode="w", suffix=".yaml", delete=False) as f:
            yaml.dump(config_data, f)
            temp_path = f.name

        try:
            yield temp_path
        finally:
            os.unlink(temp_path)

    def get_available_commands(self, force_refresh: bool = False) -> Dict[str, Any]:
        """
        Get list of available CLI commands.

        Args:
            force_refresh: Force refresh of command cache

        Returns:
            Dictionary of available commands and their metadata
        """
        with self._cache_lock:
            if self._commands_cache and not force_refresh:
                return self._commands_cache

            # Execute list-commands
            result = self.execute("list-commands", output_format=OutputFormat.JSON)

            if result.success and result.parsed_output:
                self._commands_cache = result.parsed_output
                return self._commands_cache

            return {}

    async def execute_with_cloud_tasks(
        self,
        command: str,
        context: Optional[CommandContext] = None,
        output_format: OutputFormat = OutputFormat.JSON,
        priority: str = "normal",
        callback_url: Optional[str] = None,
    ) -> Dict[str, str]:
        """
        Execute a CLI command asynchronously using Cloud Tasks.

        Args:
            command: Complete command string
            context: Execution context
            output_format: Desired output format
            priority: Task priority
            callback_url: URL to call when complete

        Returns:
            Task information including task ID
        """
        if not settings.USE_CLOUD_TASKS or not tasks_client:
            raise CommandError("Cloud Tasks not enabled")

        # Parse command
        cmd, subcmd, checkpoint_id, args = self.parse_command(command)

        # Create task ID
        task_id = f"cli_{uuid.uuid4().hex[:12]}"

        # Create Cloud Task
        task = await tasks_client.create_command_task(
            command=cmd,
            args=[subcmd] + args if subcmd else args,
            checkpoint_id=checkpoint_id or context.session_id if context else None,
            user_id=context.request_id if context else "system",
            session_id=context.session_id if context else None,
            priority=priority,
            task_id=task_id,
            callback_url=callback_url,
        )

        # Publish event if enabled
        if settings.USE_PUBSUB and pubsub_client:
            try:
                await pubsub_client.publish_event(
                    topic="command-events",
                    event_type="command.queued",
                    data={
                        "task_id": task_id,
                        "command": cmd,
                        "subcommand": subcmd,
                        "task_name": task.name,
                    },
                )
            except Exception as e:
                logger.error(f"Failed to publish task event: {e}")

        return {
            "task_id": task_id,
            "task_name": task.name,
            "status": "queued",
            "priority": priority.value,
        }

    async def store_command_logs(
        self, command_id: str, result: CommandResult
    ) -> Optional[str]:
        """
        Store command execution logs in Cloud Storage.

        Args:
            command_id: Unique command ID
            result: Command execution result

        Returns:
            Storage path if successful
        """
        if not settings.USE_CLOUD_STORAGE or not storage_client:
            return None

        try:
            # Prepare log data
            log_data = {
                "command_id": command_id,
                "command": " ".join(result.command),
                "success": result.success,
                "exit_code": result.exit_code,
                "duration": result.duration,
                "output_format": result.output_format.value,
                "timestamp": datetime.utcnow().isoformat(),
                "stdout": result.stdout,
                "stderr": result.stderr,
                "parsed_output": result.parsed_output,
                "error": result.error,
            }

            # Upload to Cloud Storage
            blob_name = f"command-logs/{datetime.utcnow().strftime('%Y/%m/%d')}/{command_id}.json"
            storage_path = await storage_client.upload_json(
                bucket_name="virtuoso-logs", blob_name=blob_name, data=log_data
            )

            logger.info(f"Stored command logs: {storage_path}")
            return storage_path

        except Exception as e:
            logger.error(f"Failed to store command logs: {e}")
            return None

    async def publish_execution_event(
        self,
        event_type: str,
        command: str,
        result: Optional[CommandResult] = None,
        context: Optional[CommandContext] = None,
        metadata: Optional[Dict[str, Any]] = None,
    ) -> bool:
        """
        Publish command execution event to Pub/Sub.

        Args:
            event_type: Type of event (e.g., "command.started", "command.completed")
            command: Command string
            result: Command result if available
            context: Command context
            metadata: Additional metadata

        Returns:
            True if published successfully
        """
        if not settings.USE_PUBSUB or not pubsub_client:
            return False

        try:
            event_data = {
                "command": command,
                "timestamp": datetime.utcnow().isoformat(),
                "request_id": context.request_id if context else None,
                "session_id": context.session_id if context else None,
            }

            if result:
                event_data.update(
                    {
                        "success": result.success,
                        "exit_code": result.exit_code,
                        "duration": result.duration,
                        "error": result.error,
                    }
                )

            if metadata:
                event_data.update(metadata)

            await pubsub_client.publish_event(
                topic="command-events", event_type=event_type, data=event_data
            )

            return True

        except Exception as e:
            logger.error(f"Failed to publish execution event: {e}")
            return False

    def shutdown(self) -> None:
        """Shutdown the executor service."""
        self.executor.shutdown(wait=True)
        logger.info("CLI executor shutdown complete")


# Singleton instance
_executor: Optional[CLIExecutor] = None
_lock = threading.Lock()


def get_executor() -> CLIExecutor:
    """
    Get or create the singleton CLI executor instance.

    Returns:
        CLIExecutor instance
    """
    global _executor

    if _executor is None:
        with _lock:
            if _executor is None:
                _executor = CLIExecutor()

    return _executor
