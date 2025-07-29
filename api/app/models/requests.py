"""
Pydantic models for API requests.

This module contains request models for command execution, test execution,
session management, and batch operations.
"""
from typing import Optional, List, Dict, Any, Union, Literal
from pydantic import BaseModel, Field, model_validator
from datetime import datetime

from .commands import VirtuosoCommand, SimplifiedStep, OutputFormat


# ========================================
# Session Management Requests
# ========================================


class CreateSessionRequest(BaseModel):
    """Request to create a new session"""

    checkpoint_id: str = Field(
        ..., description="Checkpoint ID to associate with session"
    )
    description: Optional[str] = Field(None, description="Session description")
    timeout: Optional[int] = Field(3600, description="Session timeout in seconds", gt=0)

    class Config:
        json_schema_extra = {
            "example": {
                "checkpoint_id": "12345",
                "description": "Test session for login flow",
                "timeout": 7200,
            }
        }


class UpdateSessionRequest(BaseModel):
    """Request to update session"""

    description: Optional[str] = Field(None, description="Updated description")
    timeout: Optional[int] = Field(None, description="Updated timeout in seconds", gt=0)
    extend: Optional[bool] = Field(
        False, description="Extend session timeout from current time"
    )


class SessionContext(BaseModel):
    """Session context information"""

    session_id: str = Field(..., description="Active session ID")
    checkpoint_id: str = Field(..., description="Associated checkpoint ID")
    created_at: datetime = Field(..., description="Session creation time")
    expires_at: datetime = Field(..., description="Session expiration time")
    last_position: Optional[int] = Field(
        None, description="Last step position in session"
    )


# ========================================
# Command Execution Requests
# ========================================


class ExecuteCommandRequest(BaseModel):
    """Request to execute a single command"""

    command: VirtuosoCommand = Field(..., description="Command to execute")
    session_id: Optional[str] = Field(
        None, description="Session ID (overrides command checkpoint_id)"
    )
    auto_increment_position: Optional[bool] = Field(
        True, description="Auto-increment position in session"
    )
    dry_run: Optional[bool] = Field(False, description="Validate without executing")

    class Config:
        json_schema_extra = {
            "example": {
                "command": {
                    "command": "click",
                    "selector": "button#submit",
                    "checkpoint_id": "12345",
                    "position": 1,
                },
                "session_id": "session_123",
                "auto_increment_position": True,
            }
        }


class BatchCommandRequest(BaseModel):
    """Request to execute multiple commands in sequence"""

    commands: List[VirtuosoCommand] = Field(
        ..., description="Commands to execute in order"
    )
    session_id: Optional[str] = Field(None, description="Session ID for all commands")
    stop_on_error: Optional[bool] = Field(
        True, description="Stop execution on first error"
    )
    transaction_mode: Optional[bool] = Field(
        False, description="Roll back all on any error"
    )
    callback_url: Optional[str] = Field(
        None, description="URL to call when batch completes (async)"
    )

    class Config:
        json_schema_extra = {
            "example": {
                "commands": [
                    {"command": "navigate", "url": "https://example.com"},
                    {"command": "click", "selector": "#login"},
                    {
                        "command": "write",
                        "selector": "#email",
                        "text": "test@example.com",
                    },
                ],
                "session_id": "session_123",
                "stop_on_error": False,
            }
        }


# ========================================
# Test Execution Requests
# ========================================


class TestInfrastructure(BaseModel):
    """Test infrastructure configuration"""

    organization_id: Optional[str] = Field(None, description="Organization ID")
    project_id: Optional[str] = Field(None, description="Existing project ID")
    project_name: Optional[str] = Field(
        None, description="Project name (creates if not exists)"
    )
    starting_url: Optional[str] = Field(None, description="Default starting URL")

    @model_validator(mode="after")
    def validate_project(self):
        if not self.project_id and not self.project_name:
            raise ValueError("Either project_id or project_name must be provided")
        return self


class TestVariable(BaseModel):
    """Test variable definition"""

    name: str = Field(..., description="Variable name", pattern=r"^[a-zA-Z_]\w*$")
    value: str = Field(..., description="Variable value")
    encrypted: Optional[bool] = Field(
        False, description="Whether value should be encrypted"
    )


class TestConfiguration(BaseModel):
    """Test execution configuration"""

    continue_on_error: Optional[bool] = Field(
        False, description="Continue test on step failure"
    )
    timeout: Optional[int] = Field(
        300000, description="Overall test timeout in milliseconds", gt=0
    )
    default_wait_timeout: Optional[int] = Field(
        30000, description="Default element wait timeout", gt=0
    )
    screenshot_on_failure: Optional[bool] = Field(
        True, description="Capture screenshots on failure"
    )
    video_recording: Optional[bool] = Field(False, description="Enable video recording")
    headless: Optional[bool] = Field(False, description="Run in headless mode")
    browser: Optional[str] = Field(
        "chrome",
        description="Browser to use",
        pattern=r"^(chrome|firefox|edge|safari)$",
    )


class RunTestRequest(BaseModel):
    """Request to run a test from YAML/JSON definition"""

    name: str = Field(..., description="Test name", min_length=1)
    description: Optional[str] = Field(None, description="Test description")
    infrastructure: Optional[TestInfrastructure] = Field(
        None, description="Infrastructure configuration"
    )
    variables: Optional[List[TestVariable]] = Field(None, description="Test variables")
    steps: List[Union[SimplifiedStep, Dict[str, Any]]] = Field(
        ..., description="Test steps in simplified format"
    )
    configuration: Optional[TestConfiguration] = Field(
        None, description="Test configuration"
    )
    execute_immediately: Optional[bool] = Field(
        False, description="Execute test after creation"
    )
    dry_run: Optional[bool] = Field(False, description="Validate without creating")

    class Config:
        json_schema_extra = {
            "example": {
                "name": "Login Test",
                "description": "Test user login flow",
                "steps": [
                    {"navigate": "https://example.com"},
                    {"click": "#login"},
                    {"write": {"selector": "#email", "text": "test@example.com"}},
                    {"write": {"selector": "#password", "text": "password123"}},
                    {"click": "button[type='submit']"},
                    {"assert": "Welcome"},
                ],
                "execute_immediately": True,
            }
        }


class RunTestFromFileRequest(BaseModel):
    """Request to run test from file"""

    file_path: str = Field(..., description="Path to YAML/JSON test file")
    override_name: Optional[str] = Field(None, description="Override test name")
    infrastructure: Optional[TestInfrastructure] = Field(
        None, description="Override infrastructure"
    )
    variables: Optional[List[TestVariable]] = Field(
        None, description="Additional variables"
    )
    execute_immediately: Optional[bool] = Field(
        False, description="Execute test after creation"
    )
    dry_run: Optional[bool] = Field(False, description="Validate without creating")


# ========================================
# Project Management Requests
# ========================================


class CreateProjectRequest(BaseModel):
    """Request to create a project"""

    name: str = Field(..., description="Project name", min_length=1)
    description: Optional[str] = Field(None, description="Project description")
    tags: Optional[List[str]] = Field(None, description="Project tags")

    class Config:
        json_schema_extra = {
            "example": {
                "name": "E-commerce Tests",
                "description": "Automated tests for e-commerce platform",
                "tags": ["regression", "e2e"],
            }
        }


class CreateGoalRequest(BaseModel):
    """Request to create a goal"""

    project_id: str = Field(..., description="Project ID")
    name: str = Field(..., description="Goal name", min_length=1)
    description: Optional[str] = Field(None, description="Goal description")
    starting_url: Optional[str] = Field(None, description="Starting URL for tests")


class CreateJourneyRequest(BaseModel):
    """Request to create a journey"""

    goal_id: str = Field(..., description="Goal ID")
    snapshot_id: str = Field(..., description="Snapshot ID")
    name: str = Field(..., description="Journey name", min_length=1)
    description: Optional[str] = Field(None, description="Journey description")


class CreateCheckpointRequest(BaseModel):
    """Request to create a checkpoint"""

    journey_id: str = Field(..., description="Journey ID")
    goal_id: str = Field(..., description="Goal ID")
    snapshot_id: str = Field(..., description="Snapshot ID")
    name: str = Field(..., description="Checkpoint name", min_length=1)
    description: Optional[str] = Field(None, description="Checkpoint description")


# ========================================
# Execution Management Requests
# ========================================


class CreateEnvironmentRequest(BaseModel):
    """Request to create test environment"""

    name: str = Field(..., description="Environment name", min_length=1)
    base_url: str = Field(..., description="Base URL for environment")
    variables: Optional[Dict[str, str]] = Field(
        None, description="Environment variables"
    )
    description: Optional[str] = Field(None, description="Environment description")


class ExecuteGoalRequest(BaseModel):
    """Request to execute a goal"""

    goal_id: str = Field(..., description="Goal ID to execute")
    environment_id: Optional[str] = Field(None, description="Environment ID")
    variables: Optional[Dict[str, str]] = Field(None, description="Execution variables")
    browsers: Optional[List[str]] = Field(["chrome"], description="Browsers to test on")
    parallel: Optional[bool] = Field(False, description="Run browsers in parallel")
    notification_emails: Optional[List[str]] = Field(
        None, description="Email addresses for notifications"
    )


class MonitorExecutionRequest(BaseModel):
    """Request to monitor execution"""

    execution_id: str = Field(..., description="Execution ID to monitor")
    poll_interval: Optional[int] = Field(
        5, description="Poll interval in seconds", gt=0, le=60
    )
    timeout: Optional[int] = Field(3600, description="Monitor timeout in seconds", gt=0)


# ========================================
# Query/Filter Requests
# ========================================


class ListFilter(BaseModel):
    """Common list filter options"""

    search: Optional[str] = Field(None, description="Search term")
    tags: Optional[List[str]] = Field(None, description="Filter by tags")
    created_after: Optional[datetime] = Field(None, description="Created after date")
    created_before: Optional[datetime] = Field(None, description="Created before date")
    limit: Optional[int] = Field(50, description="Result limit", ge=1, le=1000)
    offset: Optional[int] = Field(0, description="Result offset", ge=0)
    sort_by: Optional[str] = Field("created_at", description="Sort field")
    sort_order: Optional[str] = Field(
        "desc", description="Sort order", pattern=r"^(asc|desc)$"
    )


class ListProjectsRequest(BaseModel):
    """Request to list projects"""

    filter: Optional[ListFilter] = Field(None, description="Filter options")
    include_archived: Optional[bool] = Field(
        False, description="Include archived projects"
    )


class ListGoalsRequest(BaseModel):
    """Request to list goals"""

    project_id: str = Field(..., description="Project ID")
    filter: Optional[ListFilter] = Field(None, description="Filter options")


class ListJourneysRequest(BaseModel):
    """Request to list journeys"""

    goal_id: str = Field(..., description="Goal ID")
    filter: Optional[ListFilter] = Field(None, description="Filter options")


class ListCheckpointsRequest(BaseModel):
    """Request to list checkpoints"""

    journey_id: str = Field(..., description="Journey ID")
    filter: Optional[ListFilter] = Field(None, description="Filter options")


class ListStepsRequest(BaseModel):
    """Request to list steps in checkpoint"""

    checkpoint_id: str = Field(..., description="Checkpoint ID")
    include_disabled: Optional[bool] = Field(
        False, description="Include disabled steps"
    )


# ========================================
# Template Management Requests
# ========================================


class LoadTemplateRequest(BaseModel):
    """Request to load test template"""

    template_name: str = Field(..., description="Template name or path")
    variables: Optional[Dict[str, str]] = Field(None, description="Template variables")
    target_checkpoint_id: Optional[str] = Field(
        None, description="Target checkpoint for steps"
    )


class GenerateFromTemplateRequest(BaseModel):
    """Request to generate commands from template"""

    template_name: str = Field(..., description="Template name")
    context: Dict[str, Any] = Field(..., description="Template context/variables")
    output_format: Optional[OutputFormat] = Field(
        OutputFormat.JSON, description="Output format"
    )
    include_infrastructure: Optional[bool] = Field(
        True, description="Include infrastructure setup"
    )


# ========================================
# Validation Requests
# ========================================


class ValidateCommandRequest(BaseModel):
    """Request to validate command syntax"""

    command: Union[VirtuosoCommand, Dict[str, Any]] = Field(
        ..., description="Command to validate"
    )
    strict_mode: Optional[bool] = Field(True, description="Enable strict validation")


class ValidateTestRequest(BaseModel):
    """Request to validate test definition"""

    test_definition: Union[RunTestRequest, Dict[str, Any]] = Field(
        ..., description="Test to validate"
    )
    check_selectors: Optional[bool] = Field(
        False, description="Validate selectors against page"
    )
    target_url: Optional[str] = Field(None, description="URL for selector validation")


# ========================================
# Export/Import Requests
# ========================================


class ExportTestRequest(BaseModel):
    """Request to export test"""

    checkpoint_id: str = Field(..., description="Checkpoint ID to export")
    format: Literal["yaml", "json", "gherkin"] = Field(
        "yaml", description="Export format"
    )
    include_infrastructure: Optional[bool] = Field(
        True, description="Include infrastructure details"
    )
    simplify: Optional[bool] = Field(True, description="Use simplified step format")


class ImportTestRequest(BaseModel):
    """Request to import test"""

    content: str = Field(..., description="Test content to import")
    format: Literal["yaml", "json", "gherkin"] = Field(
        ..., description="Content format"
    )
    target_project_id: Optional[str] = Field(None, description="Target project ID")
    auto_create_infrastructure: Optional[bool] = Field(
        True, description="Auto-create missing infrastructure"
    )
