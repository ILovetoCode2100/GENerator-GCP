"""
Pydantic models for API responses.

This module contains standardized response models for all API operations,
including success responses, error responses, and streaming responses.
"""
from typing import Optional, List, Dict, Any, Union, Generic, TypeVar
from enum import Enum
from pydantic import BaseModel, Field
from datetime import datetime


T = TypeVar("T")


# ========================================
# Response Status and Error Types
# ========================================


class ResponseStatus(str, Enum):
    """Response status values"""

    SUCCESS = "success"
    ERROR = "error"
    WARNING = "warning"
    PARTIAL = "partial"


class ErrorType(str, Enum):
    """Error types for categorization"""

    VALIDATION = "validation_error"
    AUTHENTICATION = "authentication_error"
    AUTHORIZATION = "authorization_error"
    NOT_FOUND = "not_found"
    CONFLICT = "conflict"
    RATE_LIMIT = "rate_limit_exceeded"
    SERVER_ERROR = "server_error"
    CLIENT_ERROR = "client_error"
    TIMEOUT = "timeout"
    API_ERROR = "api_error"


class ErrorDetail(BaseModel):
    """Detailed error information"""

    field: Optional[str] = Field(None, description="Field that caused the error")
    message: str = Field(..., description="Error message")
    code: Optional[str] = Field(None, description="Error code")
    context: Optional[Dict[str, Any]] = Field(
        None, description="Additional error context"
    )


# ========================================
# Base Response Models
# ========================================


class BaseResponse(BaseModel, Generic[T]):
    """Base response model with metadata"""

    status: ResponseStatus = Field(..., description="Response status")
    data: Optional[T] = Field(None, description="Response data")
    message: Optional[str] = Field(None, description="Human-readable message")
    timestamp: datetime = Field(
        default_factory=datetime.utcnow, description="Response timestamp"
    )
    request_id: Optional[str] = Field(None, description="Request tracking ID")


class ErrorResponse(BaseModel):
    """Standard error response"""

    status: ResponseStatus = Field(ResponseStatus.ERROR, description="Response status")
    error_type: ErrorType = Field(..., description="Error type")
    message: str = Field(..., description="Error message")
    details: Optional[List[ErrorDetail]] = Field(
        None, description="Detailed error information"
    )
    timestamp: datetime = Field(
        default_factory=datetime.utcnow, description="Response timestamp"
    )
    request_id: Optional[str] = Field(None, description="Request tracking ID")

    class Config:
        json_schema_extra = {
            "example": {
                "status": "error",
                "error_type": "validation_error",
                "message": "Invalid command parameters",
                "details": [
                    {
                        "field": "selector",
                        "message": "Selector cannot be empty",
                        "code": "required",
                    }
                ],
                "timestamp": "2024-01-20T10:30:00Z",
            }
        }


class PaginatedResponse(BaseModel, Generic[T]):
    """Paginated list response"""

    status: ResponseStatus = Field(
        ResponseStatus.SUCCESS, description="Response status"
    )
    data: List[T] = Field(..., description="Page data")
    total: int = Field(..., description="Total number of items", ge=0)
    limit: int = Field(..., description="Items per page", ge=1)
    offset: int = Field(..., description="Current offset", ge=0)
    has_more: bool = Field(..., description="Whether more pages exist")
    timestamp: datetime = Field(
        default_factory=datetime.utcnow, description="Response timestamp"
    )


# ========================================
# Command Execution Responses
# ========================================


class StepResult(BaseModel):
    """Result of creating a test step"""

    step_id: str = Field(..., description="Created step ID")
    checkpoint_id: str = Field(..., description="Checkpoint ID")
    position: int = Field(..., description="Step position")
    type: str = Field(..., description="Step type")
    description: Optional[str] = Field(None, description="Step description")
    created_at: datetime = Field(..., description="Creation timestamp")

    class Config:
        json_schema_extra = {
            "example": {
                "step_id": "step_123",
                "checkpoint_id": "12345",
                "position": 1,
                "type": "CLICK",
                "description": "Click submit button",
                "created_at": "2024-01-20T10:30:00Z",
            }
        }


class CommandExecutionResponse(BaseResponse[StepResult]):
    """Response for single command execution"""

    execution_time_ms: Optional[int] = Field(
        None, description="Execution time in milliseconds"
    )
    warnings: Optional[List[str]] = Field(None, description="Execution warnings")


class BatchExecutionResult(BaseModel):
    """Result of batch command execution"""

    total_commands: int = Field(..., description="Total commands in batch")
    successful: int = Field(..., description="Successfully executed commands")
    failed: int = Field(..., description="Failed commands")
    skipped: int = Field(..., description="Skipped commands")
    results: List[Union[StepResult, ErrorDetail]] = Field(
        ..., description="Individual command results"
    )
    execution_time_ms: int = Field(..., description="Total execution time")
    batch_id: Optional[str] = Field(None, description="Batch ID for async execution")
    async_execution: bool = Field(
        False, description="Whether batch is executing asynchronously"
    )


class BatchCommandResponse(BaseResponse[BatchExecutionResult]):
    """Response for batch command execution"""

    partial_success: Optional[bool] = Field(
        None, description="Whether some commands succeeded"
    )


# ========================================
# Test Execution Responses
# ========================================


class TestCreationResult(BaseModel):
    """Result of test creation"""

    project_id: str = Field(..., description="Project ID")
    goal_id: str = Field(..., description="Goal ID")
    journey_id: str = Field(..., description="Journey ID")
    checkpoint_id: str = Field(..., description="Checkpoint ID")
    steps_created: int = Field(..., description="Number of steps created")
    test_url: Optional[str] = Field(None, description="URL to view test in UI")


class TestExecutionResult(BaseModel):
    """Result of test execution"""

    execution_id: str = Field(..., description="Execution ID")
    status: str = Field(..., description="Execution status")
    started_at: datetime = Field(..., description="Start time")
    completed_at: Optional[datetime] = Field(None, description="Completion time")
    duration_ms: Optional[int] = Field(None, description="Duration in milliseconds")
    passed_steps: int = Field(..., description="Number of passed steps")
    failed_steps: int = Field(..., description="Number of failed steps")
    total_steps: int = Field(..., description="Total number of steps")
    error_message: Optional[str] = Field(None, description="Error message if failed")
    report_url: Optional[str] = Field(None, description="URL to execution report")


class RunTestResponse(BaseResponse[TestCreationResult]):
    """Response for run-test command"""

    execution_result: Optional[TestExecutionResult] = Field(
        None, description="Execution result if run immediately"
    )
    dry_run_validation: Optional[Dict[str, Any]] = Field(
        None, description="Validation results for dry run"
    )


# ========================================
# Project Management Responses
# ========================================


class ProjectInfo(BaseModel):
    """Project information"""

    project_id: str = Field(..., description="Project ID")
    name: str = Field(..., description="Project name")
    description: Optional[str] = Field(None, description="Project description")
    created_at: datetime = Field(..., description="Creation time")
    updated_at: datetime = Field(..., description="Last update time")
    tags: List[str] = Field(default_factory=list, description="Project tags")
    goal_count: Optional[int] = Field(None, description="Number of goals")
    last_execution: Optional[datetime] = Field(None, description="Last execution time")


class GoalInfo(BaseModel):
    """Goal information"""

    goal_id: str = Field(..., description="Goal ID")
    project_id: str = Field(..., description="Parent project ID")
    snapshot_id: str = Field(..., description="Associated snapshot ID")
    name: str = Field(..., description="Goal name")
    description: Optional[str] = Field(None, description="Goal description")
    starting_url: Optional[str] = Field(None, description="Starting URL")
    created_at: datetime = Field(..., description="Creation time")
    journey_count: Optional[int] = Field(None, description="Number of journeys")


class JourneyInfo(BaseModel):
    """Journey information"""

    journey_id: str = Field(..., description="Journey ID")
    goal_id: str = Field(..., description="Parent goal ID")
    name: str = Field(..., description="Journey name")
    description: Optional[str] = Field(None, description="Journey description")
    created_at: datetime = Field(..., description="Creation time")
    checkpoint_count: Optional[int] = Field(None, description="Number of checkpoints")


class CheckpointInfo(BaseModel):
    """Checkpoint information"""

    checkpoint_id: str = Field(..., description="Checkpoint ID")
    journey_id: str = Field(..., description="Parent journey ID")
    name: str = Field(..., description="Checkpoint name")
    description: Optional[str] = Field(None, description="Checkpoint description")
    created_at: datetime = Field(..., description="Creation time")
    step_count: int = Field(..., description="Number of steps")
    last_modified: Optional[datetime] = Field(
        None, description="Last modification time"
    )


class StepInfo(BaseModel):
    """Step information"""

    step_id: str = Field(..., description="Step ID")
    checkpoint_id: str = Field(..., description="Parent checkpoint ID")
    position: int = Field(..., description="Step position")
    type: str = Field(..., description="Step type")
    selector: Optional[str] = Field(None, description="Element selector")
    value: Optional[str] = Field(None, description="Step value")
    description: Optional[str] = Field(None, description="Step description")
    enabled: bool = Field(True, description="Whether step is enabled")
    created_at: datetime = Field(..., description="Creation time")
    meta: Optional[Dict[str, Any]] = Field(None, description="Step metadata")


# Response types for list operations
CreateProjectResponse = BaseResponse[ProjectInfo]
CreateGoalResponse = BaseResponse[GoalInfo]
CreateJourneyResponse = BaseResponse[JourneyInfo]
CreateCheckpointResponse = BaseResponse[CheckpointInfo]

ListProjectsResponse = PaginatedResponse[ProjectInfo]
ListGoalsResponse = PaginatedResponse[GoalInfo]
ListJourneysResponse = PaginatedResponse[JourneyInfo]
ListCheckpointsResponse = PaginatedResponse[CheckpointInfo]
ListStepsResponse = BaseResponse[List[StepInfo]]


# ========================================
# Execution Management Responses
# ========================================


class EnvironmentInfo(BaseModel):
    """Test environment information"""

    environment_id: str = Field(..., description="Environment ID")
    name: str = Field(..., description="Environment name")
    base_url: str = Field(..., description="Base URL")
    variables: Dict[str, str] = Field(
        default_factory=dict, description="Environment variables"
    )
    created_at: datetime = Field(..., description="Creation time")
    last_used: Optional[datetime] = Field(None, description="Last used time")


class ExecutionStatus(BaseModel):
    """Execution status information"""

    execution_id: str = Field(..., description="Execution ID")
    status: str = Field(..., description="Current status")
    progress: float = Field(..., description="Progress percentage", ge=0, le=100)
    current_step: Optional[str] = Field(None, description="Currently executing step")
    elapsed_time_ms: int = Field(..., description="Elapsed time in milliseconds")
    estimated_remaining_ms: Optional[int] = Field(
        None, description="Estimated remaining time"
    )
    errors: List[ErrorDetail] = Field(
        default_factory=list, description="Execution errors"
    )


class ExecutionAnalysis(BaseModel):
    """Execution analysis results"""

    execution_id: str = Field(..., description="Execution ID")
    summary: Dict[str, Any] = Field(..., description="Execution summary")
    performance_metrics: Dict[str, float] = Field(
        ..., description="Performance metrics"
    )
    error_analysis: List[Dict[str, Any]] = Field(
        default_factory=list, description="Error analysis"
    )
    recommendations: List[str] = Field(
        default_factory=list, description="Improvement recommendations"
    )
    screenshots: List[str] = Field(default_factory=list, description="Screenshot URLs")
    video_url: Optional[str] = Field(None, description="Video recording URL")


CreateEnvironmentResponse = BaseResponse[EnvironmentInfo]
ExecuteGoalResponse = BaseResponse[TestExecutionResult]
MonitorExecutionResponse = BaseResponse[ExecutionStatus]
GetExecutionAnalysisResponse = BaseResponse[ExecutionAnalysis]


# ========================================
# Session Management Responses
# ========================================


class SessionInfo(BaseModel):
    """Session information"""

    session_id: str = Field(..., description="Session ID")
    checkpoint_id: str = Field(..., description="Associated checkpoint ID")
    created_at: datetime = Field(..., description="Creation time")
    expires_at: datetime = Field(..., description="Expiration time")
    last_activity: datetime = Field(..., description="Last activity time")
    steps_added: int = Field(..., description="Number of steps added in session")
    active: bool = Field(..., description="Whether session is active")


CreateSessionResponse = BaseResponse[SessionInfo]
GetSessionResponse = BaseResponse[SessionInfo]


# ========================================
# Library Management Responses
# ========================================


class LibraryStepInfo(BaseModel):
    """Library step information"""

    library_step_id: str = Field(..., description="Library step ID")
    name: str = Field(..., description="Step name")
    category: Optional[str] = Field(None, description="Step category")
    description: Optional[str] = Field(None, description="Step description")
    type: str = Field(..., description="Step type")
    configuration: Dict[str, Any] = Field(..., description="Step configuration")
    usage_count: int = Field(0, description="Number of times used")
    created_at: datetime = Field(..., description="Creation time")
    last_used: Optional[datetime] = Field(None, description="Last used time")


AddLibraryStepResponse = BaseResponse[LibraryStepInfo]
GetLibraryStepResponse = BaseResponse[LibraryStepInfo]
AttachLibraryStepResponse = BaseResponse[StepResult]
UpdateLibraryStepResponse = BaseResponse[LibraryStepInfo]


# ========================================
# Template Management Responses
# ========================================


class TemplateInfo(BaseModel):
    """Test template information"""

    template_id: str = Field(..., description="Template ID")
    name: str = Field(..., description="Template name")
    description: Optional[str] = Field(None, description="Template description")
    category: str = Field(..., description="Template category")
    variables: List[Dict[str, Any]] = Field(..., description="Template variables")
    step_count: int = Field(..., description="Number of steps")
    tags: List[str] = Field(default_factory=list, description="Template tags")


class GeneratedCommands(BaseModel):
    """Generated commands from template"""

    commands: List[Dict[str, Any]] = Field(..., description="Generated commands")
    infrastructure_commands: Optional[List[Dict[str, Any]]] = Field(
        None, description="Infrastructure setup commands"
    )
    total_steps: int = Field(..., description="Total number of steps")
    estimated_duration_ms: Optional[int] = Field(
        None, description="Estimated execution duration"
    )


LoadTemplateResponse = BaseResponse[TemplateInfo]
GenerateCommandsResponse = BaseResponse[GeneratedCommands]
ListTemplatesResponse = PaginatedResponse[TemplateInfo]


# ========================================
# Validation Responses
# ========================================


class ValidationResult(BaseModel):
    """Validation result"""

    valid: bool = Field(..., description="Whether validation passed")
    errors: List[ErrorDetail] = Field(
        default_factory=list, description="Validation errors"
    )
    warnings: List[str] = Field(default_factory=list, description="Validation warnings")
    suggestions: List[str] = Field(
        default_factory=list, description="Improvement suggestions"
    )


ValidateCommandResponse = BaseResponse[ValidationResult]
ValidateTestResponse = BaseResponse[ValidationResult]


# ========================================
# Export/Import Responses
# ========================================


class ExportResult(BaseModel):
    """Test export result"""

    format: str = Field(..., description="Export format")
    content: str = Field(..., description="Exported content")
    metadata: Dict[str, Any] = Field(..., description="Export metadata")


class ImportResult(BaseModel):
    """Test import result"""

    imported_items: Dict[str, int] = Field(
        ..., description="Count of imported items by type"
    )
    project_id: Optional[str] = Field(None, description="Created/used project ID")
    checkpoint_ids: List[str] = Field(..., description="Created checkpoint IDs")
    warnings: List[str] = Field(default_factory=list, description="Import warnings")


ExportTestResponse = BaseResponse[ExportResult]
ImportTestResponse = BaseResponse[ImportResult]


# ========================================
# Streaming Response Models
# ========================================


class StreamEventType(str, Enum):
    """Types of streaming events"""

    STEP_START = "step_start"
    STEP_COMPLETE = "step_complete"
    STEP_FAILED = "step_failed"
    LOG = "log"
    SCREENSHOT = "screenshot"
    ERROR = "error"
    WARNING = "warning"
    PROGRESS = "progress"
    COMPLETE = "complete"


class StreamEvent(BaseModel):
    """Streaming event model"""

    event_type: StreamEventType = Field(..., description="Event type")
    timestamp: datetime = Field(
        default_factory=datetime.utcnow, description="Event timestamp"
    )
    data: Dict[str, Any] = Field(..., description="Event data")

    class Config:
        json_schema_extra = {
            "example": {
                "event_type": "step_complete",
                "timestamp": "2024-01-20T10:30:00Z",
                "data": {
                    "step_id": "step_123",
                    "position": 1,
                    "type": "CLICK",
                    "duration_ms": 150,
                },
            }
        }


# ========================================
# Health Check Response
# ========================================


class HealthStatus(BaseModel):
    """System health status"""

    healthy: bool = Field(..., description="Overall health status")
    api_version: str = Field(..., description="API version")
    cli_compatible: bool = Field(..., description="CLI compatibility status")
    services: Dict[str, bool] = Field(..., description="Individual service statuses")
    timestamp: datetime = Field(
        default_factory=datetime.utcnow, description="Check timestamp"
    )


HealthCheckResponse = BaseResponse[HealthStatus]
