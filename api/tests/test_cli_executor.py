"""
Tests for CLI Executor Service.
"""

import asyncio
import os
import pytest
from unittest.mock import Mock, patch, MagicMock

from api.app.services.cli_executor import (
    CLIExecutor,
    CommandContext,
    CommandResult,
    OutputFormat,
    CommandValidationError,
    CommandTimeoutError,
    get_executor,
)


@pytest.fixture
def mock_settings():
    """Mock settings for testing."""
    with patch("api.app.services.cli_executor.settings") as mock:
        mock.CLI_PATH = "/usr/bin/true"  # Command that always succeeds
        mock.CLI_TIMEOUT = 5
        mock.get_cli_env.return_value = {"TEST_ENV": "value"}
        yield mock


@pytest.fixture
def cli_executor(mock_settings):
    """Create CLI executor instance for testing."""
    with patch.object(CLIExecutor, "_validate_cli_binary"):
        executor = CLIExecutor(cli_path="/usr/bin/echo")
        yield executor
        executor.shutdown()


@pytest.fixture
def command_context():
    """Create a test command context."""
    return CommandContext(
        request_id="test-123",
        session_id="12345",
        timeout=10.0,
        environment={"CUSTOM_VAR": "test"},
    )


class TestCLIExecutorInit:
    """Test CLI executor initialization."""

    def test_init_default_values(self, mock_settings):
        """Test initialization with default values."""
        with patch.object(CLIExecutor, "_validate_cli_binary"):
            executor = CLIExecutor()
            assert executor.cli_path == mock_settings.CLI_PATH
            assert executor.default_timeout == 300.0
            assert executor.executor._max_workers == 4

    def test_init_custom_values(self, mock_settings):
        """Test initialization with custom values."""
        with patch.object(CLIExecutor, "_validate_cli_binary"):
            executor = CLIExecutor(
                cli_path="/custom/path", default_timeout=60.0, max_workers=8
            )
            assert executor.cli_path == "/custom/path"
            assert executor.default_timeout == 60.0
            assert executor.executor._max_workers == 8

    def test_validate_cli_binary_not_found(self):
        """Test validation when CLI binary not found."""
        with pytest.raises(CommandValidationError) as exc:
            CLIExecutor(cli_path="/nonexistent/path")
        assert "not found" in str(exc.value)

    def test_validate_cli_binary_not_executable(self, tmp_path):
        """Test validation when CLI binary not executable."""
        # Create non-executable file
        cli_file = tmp_path / "cli"
        cli_file.write_text("#!/bin/bash\necho test")

        with pytest.raises(CommandValidationError) as exc:
            CLIExecutor(cli_path=str(cli_file))
        assert "not executable" in str(exc.value)


class TestCommandParsing:
    """Test command parsing functionality."""

    def test_parse_simple_command(self, cli_executor):
        """Test parsing simple command."""
        cmd, subcmd, checkpoint, args = cli_executor.parse_command("list-projects")
        assert cmd == "list-projects"
        assert subcmd is None
        assert checkpoint is None
        assert args == []

    def test_parse_command_with_args(self, cli_executor):
        """Test parsing command with arguments."""
        cmd, subcmd, checkpoint, args = cli_executor.parse_command(
            'create-project "My Project" --description "Test"'
        )
        assert cmd == "create-project"
        assert subcmd is None
        assert checkpoint is None
        assert args == ["My Project", "--description", "Test"]

    def test_parse_step_command(self, cli_executor):
        """Test parsing step command with checkpoint."""
        cmd, subcmd, checkpoint, args = cli_executor.parse_command(
            'step-navigate to 12345 "https://example.com" 1'
        )
        assert cmd == "step-navigate"
        assert subcmd == "to"
        assert checkpoint == "12345"
        assert args == ["https://example.com", "1"]

    def test_parse_step_command_no_checkpoint(self, cli_executor):
        """Test parsing step command without checkpoint."""
        cmd, subcmd, checkpoint, args = cli_executor.parse_command(
            'step-interact click "button.submit"'
        )
        assert cmd == "step-interact"
        assert subcmd == "click"
        assert checkpoint is None
        assert args == ["button.submit"]

    def test_parse_command_with_api_cli_prefix(self, cli_executor):
        """Test parsing command with api-cli prefix."""
        cmd, subcmd, checkpoint, args = cli_executor.parse_command(
            "api-cli list-projects --format json"
        )
        assert cmd == "list-projects"
        assert args == ["--format", "json"]

    def test_parse_empty_command(self, cli_executor):
        """Test parsing empty command."""
        with pytest.raises(CommandValidationError) as exc:
            cli_executor.parse_command("")
        assert "Empty command" in str(exc.value)

    def test_parse_command_with_quotes(self, cli_executor):
        """Test parsing command with quoted arguments."""
        cmd, subcmd, checkpoint, args = cli_executor.parse_command(
            'step-assert equals "h1" "Welcome to "My Site""'
        )
        assert cmd == "step-assert"
        assert subcmd == "equals"
        assert args == ["h1", 'Welcome to "My Site"']


class TestCommandValidation:
    """Test command validation functionality."""

    def test_validate_step_command_requires_checkpoint(self, cli_executor):
        """Test validation of step command without checkpoint."""
        with pytest.raises(CommandValidationError) as exc:
            cli_executor.validate_command("step-navigate", "to")
        assert "requires checkpoint ID" in str(exc.value)

    def test_validate_step_command_with_checkpoint(self, cli_executor):
        """Test validation of step command with checkpoint."""
        # Should not raise
        cli_executor.validate_command("step-navigate", "to", "12345")

    def test_validate_step_command_with_session(self, cli_executor, command_context):
        """Test validation of step command with session context."""
        # Should not raise - uses session ID
        cli_executor.validate_command("step-navigate", "to", None, command_context)

    def test_validate_non_step_command(self, cli_executor):
        """Test validation of non-step command."""
        # Should not raise
        cli_executor.validate_command("list-projects")

    def test_validate_library_command_optional_checkpoint(self, cli_executor):
        """Test validation of library command (checkpoint optional)."""
        # Should not raise even without checkpoint
        cli_executor.validate_command("library", "get")


class TestCommandBuilding:
    """Test command building functionality."""

    def test_build_simple_command(self, cli_executor):
        """Test building simple command."""
        args = cli_executor.build_command_args(
            "list-projects", output_format=OutputFormat.JSON
        )
        assert args == ["/usr/bin/echo", "list-projects", "--output", "json"]

    def test_build_step_command(self, cli_executor):
        """Test building step command."""
        args = cli_executor.build_command_args(
            "step-navigate",
            subcommand="to",
            checkpoint_id="12345",
            args=["https://example.com", "1"],
            output_format=OutputFormat.YAML,
        )
        expected = [
            "/usr/bin/echo",
            "step-navigate",
            "to",
            "12345",
            "https://example.com",
            "1",
            "--output",
            "yaml",
        ]
        assert args == expected

    def test_build_command_raw_format(self, cli_executor):
        """Test building command with raw output format."""
        args = cli_executor.build_command_args(
            "list-projects", output_format=OutputFormat.RAW
        )
        # No output format added for raw
        assert args == ["/usr/bin/echo", "list-projects"]


class TestCommandExecution:
    """Test command execution functionality."""

    @patch("subprocess.Popen")
    def test_execute_success(self, mock_popen, cli_executor, command_context):
        """Test successful command execution."""
        # Mock successful execution
        mock_process = Mock()
        mock_process.communicate.return_value = ('{"success": true}', "")
        mock_process.returncode = 0
        mock_popen.return_value = mock_process

        result = cli_executor.execute(
            "list-projects", context=command_context, output_format=OutputFormat.JSON
        )

        assert result.success is True
        assert result.exit_code == 0
        assert result.parsed_output == {"success": True}
        assert result.error is None
        assert result.duration > 0

    @patch("subprocess.Popen")
    def test_execute_failure(self, mock_popen, cli_executor):
        """Test failed command execution."""
        # Mock failed execution
        mock_process = Mock()
        mock_process.communicate.return_value = ("", "Error: Not found")
        mock_process.returncode = 5
        mock_popen.return_value = mock_process

        result = cli_executor.execute("get-project invalid")

        assert result.success is False
        assert result.exit_code == 5
        assert "Resource not found" in result.error
        assert "Error: Not found" in result.error

    @patch("subprocess.Popen")
    def test_execute_timeout(self, mock_popen, cli_executor, command_context):
        """Test command execution timeout."""
        # Mock timeout
        mock_process = Mock()
        mock_process.communicate.side_effect = subprocess.TimeoutExpired("cmd", 5)
        mock_process.kill = Mock()
        mock_popen.return_value = mock_process

        command_context.timeout = 0.1

        with pytest.raises(CommandTimeoutError) as exc:
            cli_executor.execute("long-running-command", context=command_context)

        assert "timed out after 0.1 seconds" in str(exc.value)
        mock_process.kill.assert_called_once()

    @patch("subprocess.Popen")
    def test_execute_with_environment(self, mock_popen, cli_executor, command_context):
        """Test command execution with custom environment."""
        mock_process = Mock()
        mock_process.communicate.return_value = ("", "")
        mock_process.returncode = 0
        mock_popen.return_value = mock_process

        cli_executor.execute("test-command", context=command_context)

        # Check environment variables were passed
        call_args = mock_popen.call_args
        env = call_args[1]["env"]
        assert env["CUSTOM_VAR"] == "test"
        assert env["VIRTUOSO_SESSION_ID"] == "12345"
        assert env["TEST_ENV"] == "value"

    def test_execute_parse_json_output(self, cli_executor):
        """Test parsing JSON output."""
        with patch("subprocess.Popen") as mock_popen:
            mock_process = Mock()
            mock_process.communicate.return_value = ('{"id": 123, "name": "Test"}', "")
            mock_process.returncode = 0
            mock_popen.return_value = mock_process

            result = cli_executor.execute(
                "get-project 123", output_format=OutputFormat.JSON
            )

            assert result.parsed_output == {"id": 123, "name": "Test"}

    def test_execute_parse_yaml_output(self, cli_executor):
        """Test parsing YAML output."""
        with patch("subprocess.Popen") as mock_popen:
            mock_process = Mock()
            mock_process.communicate.return_value = ("id: 123\nname: Test\n", "")
            mock_process.returncode = 0
            mock_popen.return_value = mock_process

            result = cli_executor.execute(
                "get-project 123", output_format=OutputFormat.YAML
            )

            assert result.parsed_output == {"id": 123, "name": "Test"}


class TestAsyncExecution:
    """Test asynchronous command execution."""

    @pytest.mark.asyncio
    async def test_execute_async(self, cli_executor):
        """Test async command execution."""
        with patch.object(cli_executor, "execute") as mock_execute:
            mock_result = CommandResult(
                success=True,
                exit_code=0,
                stdout='{"success": true}',
                stderr="",
                duration=0.1,
                command=["test"],
                output_format=OutputFormat.JSON,
                parsed_output={"success": True},
            )
            mock_execute.return_value = mock_result

            result = await cli_executor.execute_async("test-command")

            assert result.success is True
            assert result.parsed_output == {"success": True}

    @pytest.mark.asyncio
    async def test_execute_stream(self, cli_executor):
        """Test streaming command execution."""
        with patch("asyncio.create_subprocess_exec") as mock_create:
            # Mock subprocess with streaming output
            mock_process = MagicMock()
            mock_process.returncode = 0

            # Mock readline to return lines then empty
            lines = [b"Line 1\n", b"Line 2\n", b""]
            mock_process.stdout.readline = asyncio.coroutine(
                lambda: lines.pop(0) if lines else b""
            )
            mock_process.wait = asyncio.coroutine(lambda: None)
            mock_process.stderr.read = asyncio.coroutine(lambda: b"")

            mock_create.return_value = mock_process

            # Collect streamed output
            output = []
            async for line in cli_executor.execute_stream("test-command"):
                output.append(line)

            assert output == ["Line 1", "Line 2"]


class TestBatchExecution:
    """Test batch command execution."""

    def test_execute_batch(self, cli_executor):
        """Test batch command execution."""
        with patch.object(cli_executor, "execute") as mock_execute:
            # Mock different results
            mock_execute.side_effect = [
                CommandResult(
                    success=True,
                    exit_code=0,
                    stdout="Result 1",
                    stderr="",
                    duration=0.1,
                    command=["cmd1"],
                    output_format=OutputFormat.JSON,
                ),
                CommandResult(
                    success=False,
                    exit_code=1,
                    stdout="",
                    stderr="Error",
                    duration=0.1,
                    command=["cmd2"],
                    output_format=OutputFormat.JSON,
                    error="Error",
                ),
                CommandResult(
                    success=True,
                    exit_code=0,
                    stdout="Result 3",
                    stderr="",
                    duration=0.1,
                    command=["cmd3"],
                    output_format=OutputFormat.JSON,
                ),
            ]

            commands = ["command1", "command2", "command3"]
            results = cli_executor.execute_batch(commands)

            assert len(results) == 3
            assert results[0].success is True
            assert results[1].success is False
            assert results[2].success is True

    def test_execute_batch_with_timeout(self, cli_executor):
        """Test batch execution with timeout."""
        with patch.object(cli_executor.executor, "submit") as mock_submit:
            # Create futures that timeout
            future1 = Mock()
            future1.result.side_effect = TimeoutError()

            future2 = Mock()
            future2.result.return_value = CommandResult(
                success=True,
                exit_code=0,
                stdout="OK",
                stderr="",
                duration=0.1,
                command=["cmd2"],
                output_format=OutputFormat.JSON,
            )

            mock_submit.side_effect = [future1, future2]

            results = cli_executor.execute_batch(["cmd1", "cmd2"])

            assert len(results) == 2
            assert results[0].success is False
            assert results[0].exit_code == 8  # Timeout code
            assert results[1].success is True


class TestUtilityMethods:
    """Test utility methods."""

    def test_temporary_config(self, cli_executor):
        """Test temporary config context manager."""
        config_data = {
            "api": {"auth_token": "test-token"},
            "organization": {"id": "123"},
        }

        with cli_executor.temporary_config(config_data) as config_path:
            assert os.path.exists(config_path)

            # Read and verify content
            with open(config_path, "r") as f:
                loaded = yaml.safe_load(f)
                assert loaded == config_data

        # File should be deleted after context
        assert not os.path.exists(config_path)

    def test_get_available_commands(self, cli_executor):
        """Test getting available commands."""
        with patch.object(cli_executor, "execute") as mock_execute:
            mock_execute.return_value = CommandResult(
                success=True,
                exit_code=0,
                stdout='{"commands": ["list", "create"]}',
                stderr="",
                duration=0.1,
                command=["list-commands"],
                output_format=OutputFormat.JSON,
                parsed_output={"commands": ["list", "create"]},
            )

            commands = cli_executor.get_available_commands()
            assert commands == {"commands": ["list", "create"]}

            # Test caching
            commands2 = cli_executor.get_available_commands()
            assert commands2 == commands
            mock_execute.assert_called_once()  # Not called again

            # Test force refresh
            commands3 = cli_executor.get_available_commands(force_refresh=True)
            assert mock_execute.call_count == 2

    def test_get_error_message(self, cli_executor):
        """Test error message generation."""
        # Known exit code
        msg = cli_executor._get_error_message(3, "Invalid token")
        assert msg == "Authentication error: Invalid token"

        # Unknown exit code
        msg = cli_executor._get_error_message(99, "")
        assert msg == "Unknown error (code: 99)"

        # Known code without stderr
        msg = cli_executor._get_error_message(5, "")
        assert msg == "Resource not found"


class TestSingleton:
    """Test singleton pattern."""

    def test_get_executor_singleton(self):
        """Test get_executor returns singleton."""
        with patch.object(CLIExecutor, "_validate_cli_binary"):
            executor1 = get_executor()
            executor2 = get_executor()

            assert executor1 is executor2

    def test_executor_shutdown(self, cli_executor):
        """Test executor shutdown."""
        with patch.object(cli_executor.executor, "shutdown") as mock_shutdown:
            cli_executor.shutdown()
            mock_shutdown.assert_called_once_with(wait=True)
