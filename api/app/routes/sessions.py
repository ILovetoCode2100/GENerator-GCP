"""
Session management endpoints.
"""

from datetime import datetime, timezone
from typing import Dict, Any, List, Optional
from uuid import uuid4

from fastapi import APIRouter, HTTPException, status, Depends, BackgroundTasks
from pydantic import BaseModel, Field

from ..config import settings
from ..utils.logger import get_logger
from ..services.auth_service import AuthUser, Permission
from ..middleware.auth import get_authenticated_user, require_permissions
from ..middleware.rate_limit import rate_limit, RateLimitStrategy
from ..models.responses import BaseResponse, ResponseStatus

# GCP imports
if settings.is_gcp_enabled:
    from ..gcp.firestore_client import FirestoreClient
    from ..gcp.pubsub_client import PubSubClient

router = APIRouter()
logger = get_logger(__name__)

# Initialize GCP clients if enabled
firestore_client = None
pubsub_client = None

if settings.is_gcp_enabled:
    if settings.USE_FIRESTORE:
        firestore_client = FirestoreClient()
    if settings.USE_PUBSUB:
        pubsub_client = PubSubClient()


class SessionCreate(BaseModel):
    """Request model for creating a session."""

    name: str = Field(..., description="Session name")
    description: Optional[str] = Field(None, description="Session description")
    checkpoint_id: Optional[str] = Field(None, description="Associated checkpoint ID")


class Session(BaseModel):
    """Session model."""

    session_id: str = Field(..., description="Session ID")
    name: str = Field(..., description="Session name")
    description: Optional[str] = Field(None, description="Session description")
    checkpoint_id: Optional[str] = Field(None, description="Associated checkpoint ID")
    created_at: datetime = Field(..., description="Creation timestamp")
    updated_at: datetime = Field(..., description="Last update timestamp")
    status: str = Field(..., description="Session status")


class SessionUpdate(BaseModel):
    """Request model for updating a session."""

    name: Optional[str] = Field(None, description="Session name")
    description: Optional[str] = Field(None, description="Session description")
    checkpoint_id: Optional[str] = Field(None, description="Associated checkpoint ID")
    status: Optional[str] = Field(None, description="Session status")


@router.post(
    "/",
    response_model=Session,
    dependencies=[Depends(rate_limit(20, 60, RateLimitStrategy.PER_USER))],
)
async def create_session(
    request: SessionCreate,
    user: AuthUser = Depends(get_authenticated_user),
    background_tasks: BackgroundTasks = BackgroundTasks(),
) -> Session:
    """
    Create a new session.

    Args:
        request: Session creation request
        user: Authenticated user
        background_tasks: Background tasks

    Returns:
        Created session
    """
    session_id = f"sess_{uuid4().hex[:12]}"
    now = datetime.now(timezone.utc)

    # Use Firestore if available
    if settings.USE_FIRESTORE and firestore_client:
        try:
            session_data = await firestore_client.create_session(
                session_id=session_id,
                user_id=user.user_id,
                checkpoint_id=request.checkpoint_id,
                metadata={
                    "name": request.name,
                    "description": request.description,
                    "created_by": user.username,
                },
            )

            # Publish event
            if settings.USE_PUBSUB and pubsub_client:
                background_tasks.add_task(
                    publish_session_event,
                    event_type="session.created",
                    session_id=session_id,
                    user_id=user.user_id,
                    data={"checkpoint_id": request.checkpoint_id},
                )

            return Session(
                session_id=session_data["session_id"],
                name=request.name,
                description=request.description,
                checkpoint_id=session_data["checkpoint_id"],
                created_at=session_data["created_at"],
                updated_at=session_data["updated_at"],
                status="active",
            )

        except Exception as e:
            logger.error(f"Failed to create session in Firestore: {e}")
            # Fall back to in-memory

    # In-memory fallback (not persistent)
    return Session(
        session_id=session_id,
        name=request.name,
        description=request.description,
        checkpoint_id=request.checkpoint_id,
        created_at=now,
        updated_at=now,
        status="active",
    )


@router.get(
    "/",
    response_model=List[Session],
    dependencies=[Depends(require_permissions(Permission.READ_TESTS))],
)
async def list_sessions(
    status: Optional[str] = None,
    limit: int = 100,
    offset: int = 0,
    user: AuthUser = Depends(get_authenticated_user),
) -> List[Session]:
    """
    List all sessions for the current user.

    Args:
        status: Filter by status
        limit: Maximum number of results
        offset: Offset for pagination
        user: Authenticated user

    Returns:
        List of sessions
    """
    sessions = []

    # Use Firestore if available
    if settings.USE_FIRESTORE and firestore_client:
        try:
            session_data_list = await firestore_client.list_user_sessions(
                user_id=user.user_id, active_only=(status == "active")
            )

            # Apply pagination
            start_idx = offset
            end_idx = offset + limit
            paginated_sessions = session_data_list[start_idx:end_idx]

            for session_data in paginated_sessions:
                metadata = session_data.get("metadata", {})
                sessions.append(
                    Session(
                        session_id=session_data["session_id"],
                        name=metadata.get("name", "Unnamed Session"),
                        description=metadata.get("description"),
                        checkpoint_id=session_data.get("checkpoint_id"),
                        created_at=session_data["created_at"],
                        updated_at=session_data["updated_at"],
                        status="active" if session_data.get("active") else "inactive",
                    )
                )

        except Exception as e:
            logger.error(f"Failed to list sessions from Firestore: {e}")

    return sessions


@router.get(
    "/{session_id}",
    response_model=Session,
    dependencies=[Depends(require_permissions(Permission.READ_TESTS))],
)
async def get_session(
    session_id: str, user: AuthUser = Depends(get_authenticated_user)
) -> Session:
    """
    Get a specific session.

    Args:
        session_id: Session ID
        user: Authenticated user

    Returns:
        Session details
    """
    # Use Firestore if available
    if settings.USE_FIRESTORE and firestore_client:
        try:
            session_data = await firestore_client.get_session(session_id)

            if session_data:
                # Verify user owns this session
                if session_data.get("user_id") != user.user_id:
                    raise HTTPException(
                        status_code=status.HTTP_403_FORBIDDEN,
                        detail="Access denied to this session",
                    )

                metadata = session_data.get("metadata", {})
                return Session(
                    session_id=session_data["session_id"],
                    name=metadata.get("name", "Unnamed Session"),
                    description=metadata.get("description"),
                    checkpoint_id=session_data.get("checkpoint_id"),
                    created_at=session_data["created_at"],
                    updated_at=session_data["updated_at"],
                    status="active" if session_data.get("active") else "inactive",
                )

        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Failed to get session from Firestore: {e}")

    raise HTTPException(
        status_code=status.HTTP_404_NOT_FOUND, detail=f"Session not found: {session_id}"
    )


@router.patch(
    "/{session_id}",
    response_model=Session,
    dependencies=[Depends(rate_limit(50, 60, RateLimitStrategy.PER_USER))],
)
async def update_session(
    session_id: str,
    request: SessionUpdate,
    user: AuthUser = Depends(get_authenticated_user),
    background_tasks: BackgroundTasks = BackgroundTasks(),
) -> Session:
    """
    Update a session.

    Args:
        session_id: Session ID
        request: Update request
        user: Authenticated user
        background_tasks: Background tasks

    Returns:
        Updated session
    """
    # Use Firestore if available
    if settings.USE_FIRESTORE and firestore_client:
        try:
            # Verify ownership first
            existing_session = await firestore_client.get_session(session_id)
            if not existing_session or existing_session.get("user_id") != user.user_id:
                raise HTTPException(
                    status_code=status.HTTP_403_FORBIDDEN,
                    detail="Access denied to this session",
                )

            # Prepare metadata update
            metadata_update = {}
            if request.name is not None:
                metadata_update["name"] = request.name
            if request.description is not None:
                metadata_update["description"] = request.description

            # Update session
            updated_session = await firestore_client.update_session(
                session_id=session_id,
                checkpoint_id=request.checkpoint_id,
                metadata={**existing_session.get("metadata", {}), **metadata_update},
                extend_expiry=True,
            )

            if updated_session:
                # Publish event
                if settings.USE_PUBSUB and pubsub_client:
                    background_tasks.add_task(
                        publish_session_event,
                        event_type="session.updated",
                        session_id=session_id,
                        user_id=user.user_id,
                        data={"updates": request.dict(exclude_unset=True)},
                    )

                metadata = updated_session.get("metadata", {})
                return Session(
                    session_id=updated_session["session_id"],
                    name=metadata.get("name", "Unnamed Session"),
                    description=metadata.get("description"),
                    checkpoint_id=updated_session.get("checkpoint_id"),
                    created_at=updated_session["created_at"],
                    updated_at=updated_session["updated_at"],
                    status="active" if updated_session.get("active") else "inactive",
                )

        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Failed to update session in Firestore: {e}")

    raise HTTPException(
        status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
        detail="Failed to update session",
    )


@router.delete(
    "/{session_id}",
    status_code=status.HTTP_204_NO_CONTENT,
    dependencies=[Depends(rate_limit(20, 60, RateLimitStrategy.PER_USER))],
)
async def delete_session(
    session_id: str,
    user: AuthUser = Depends(get_authenticated_user),
    background_tasks: BackgroundTasks = BackgroundTasks(),
) -> None:
    """
    Delete a session.

    Args:
        session_id: Session ID
        user: Authenticated user
        background_tasks: Background tasks
    """
    # Use Firestore if available
    if settings.USE_FIRESTORE and firestore_client:
        try:
            # Verify ownership first
            existing_session = await firestore_client.get_session(session_id)
            if not existing_session or existing_session.get("user_id") != user.user_id:
                raise HTTPException(
                    status_code=status.HTTP_403_FORBIDDEN,
                    detail="Access denied to this session",
                )

            # Delete session
            success = await firestore_client.delete_session(session_id)

            if success:
                # Publish event
                if settings.USE_PUBSUB and pubsub_client:
                    background_tasks.add_task(
                        publish_session_event,
                        event_type="session.deleted",
                        session_id=session_id,
                        user_id=user.user_id,
                        data={},
                    )
                return

        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Failed to delete session from Firestore: {e}")

    raise HTTPException(
        status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
        detail="Failed to delete session",
    )


@router.post(
    "/{session_id}/activate",
    response_model=Dict[str, Any],
    dependencies=[Depends(rate_limit(30, 60, RateLimitStrategy.PER_USER))],
)
async def activate_session(
    session_id: str,
    user: AuthUser = Depends(get_authenticated_user),
    background_tasks: BackgroundTasks = BackgroundTasks(),
) -> Dict[str, Any]:
    """
    Activate a session (set as current).

    Args:
        session_id: Session ID
        user: Authenticated user
        background_tasks: Background tasks

    Returns:
        Activation result
    """
    # Use Firestore if available
    if settings.USE_FIRESTORE and firestore_client:
        try:
            # Get and verify session
            session_data = await firestore_client.get_session(session_id)
            if not session_data or session_data.get("user_id") != user.user_id:
                raise HTTPException(
                    status_code=status.HTTP_403_FORBIDDEN,
                    detail="Access denied to this session",
                )

            # Update session to extend expiry
            await firestore_client.update_session(
                session_id=session_id, extend_expiry=True
            )

            # Store current session for user (in cache)
            cache_key = f"user_current_session:{user.user_id}"
            await firestore_client.cache_set(
                key=cache_key,
                value={
                    "session_id": session_id,
                    "checkpoint_id": session_data.get("checkpoint_id"),
                    "activated_at": datetime.now(timezone.utc).isoformat(),
                },
                ttl_seconds=settings.SESSION_TIMEOUT,
            )

            # Publish event
            if settings.USE_PUBSUB and pubsub_client:
                background_tasks.add_task(
                    publish_session_event,
                    event_type="session.activated",
                    session_id=session_id,
                    user_id=user.user_id,
                    data={"checkpoint_id": session_data.get("checkpoint_id")},
                )

            return {
                "session_id": session_id,
                "checkpoint_id": session_data.get("checkpoint_id"),
                "activated": True,
                "message": "Session activated successfully",
            }

        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Failed to activate session: {e}")

    raise HTTPException(
        status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
        detail="Failed to activate session",
    )


@router.get(
    "/current",
    response_model=BaseResponse[Optional[Session]],
    dependencies=[Depends(require_permissions(Permission.READ_TESTS))],
)
async def get_current_session(
    user: AuthUser = Depends(get_authenticated_user),
) -> BaseResponse[Optional[Session]]:
    """
    Get the current active session for the user.

    Args:
        user: Authenticated user

    Returns:
        Current session or None
    """
    # Use Firestore if available
    if settings.USE_FIRESTORE and firestore_client:
        try:
            # Check cache for current session
            cache_key = f"user_current_session:{user.user_id}"
            current_session_info = await firestore_client.cache_get(cache_key)

            if current_session_info:
                session_id = current_session_info.get("session_id")
                session_data = await firestore_client.get_session(session_id)

                if session_data:
                    metadata = session_data.get("metadata", {})
                    return BaseResponse(
                        status=ResponseStatus.SUCCESS,
                        data=Session(
                            session_id=session_data["session_id"],
                            name=metadata.get("name", "Unnamed Session"),
                            description=metadata.get("description"),
                            checkpoint_id=session_data.get("checkpoint_id"),
                            created_at=session_data["created_at"],
                            updated_at=session_data["updated_at"],
                            status="active",
                        ),
                        message="Current session retrieved",
                    )

        except Exception as e:
            logger.error(f"Failed to get current session: {e}")

    return BaseResponse(
        status=ResponseStatus.SUCCESS, data=None, message="No active session"
    )


@router.get(
    "/{session_id}/analytics",
    response_model=BaseResponse[Dict[str, Any]],
    dependencies=[Depends(require_permissions(Permission.READ_TESTS))],
)
async def get_session_analytics(
    session_id: str, user: AuthUser = Depends(get_authenticated_user)
) -> BaseResponse[Dict[str, Any]]:
    """
    Get analytics for a session.

    Args:
        session_id: Session ID
        user: Authenticated user

    Returns:
        Session analytics
    """
    analytics = {
        "session_id": session_id,
        "commands_executed": 0,
        "success_rate": 0.0,
        "average_duration_ms": 0.0,
        "most_used_commands": [],
        "error_summary": {},
    }

    # Use Firestore if available
    if settings.USE_FIRESTORE and firestore_client:
        try:
            # Verify session ownership
            session_data = await firestore_client.get_session(session_id)
            if not session_data or session_data.get("user_id") != user.user_id:
                raise HTTPException(
                    status_code=status.HTTP_403_FORBIDDEN,
                    detail="Access denied to this session",
                )

            # Get command history for session
            command_history = await firestore_client.get_command_history(
                session_id=session_id, limit=1000
            )

            if command_history:
                # Calculate analytics
                total_commands = len(command_history)
                successful_commands = sum(
                    1 for cmd in command_history if cmd.get("success", False)
                )
                total_duration = sum(
                    cmd.get("duration_ms", 0) for cmd in command_history
                )

                analytics["commands_executed"] = total_commands
                analytics["success_rate"] = (
                    (successful_commands / total_commands * 100)
                    if total_commands > 0
                    else 0
                )
                analytics["average_duration_ms"] = (
                    (total_duration / total_commands) if total_commands > 0 else 0
                )

                # Command frequency
                command_counts = {}
                error_counts = {}

                for cmd in command_history:
                    command_name = cmd.get("command", "unknown")
                    command_counts[command_name] = (
                        command_counts.get(command_name, 0) + 1
                    )

                    if not cmd.get("success", False) and cmd.get("error"):
                        error_type = cmd.get("error", "Unknown error")
                        error_counts[error_type] = error_counts.get(error_type, 0) + 1

                # Sort by frequency
                analytics["most_used_commands"] = [
                    {"command": cmd, "count": count}
                    for cmd, count in sorted(
                        command_counts.items(), key=lambda x: x[1], reverse=True
                    )[:10]
                ]

                analytics["error_summary"] = error_counts

        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Failed to get session analytics: {e}")

    return BaseResponse(
        status=ResponseStatus.SUCCESS,
        data=analytics,
        message="Session analytics retrieved",
    )


# Helper functions
async def publish_session_event(
    event_type: str, session_id: str, user_id: str, data: Dict[str, Any]
):
    """Publish session event to Pub/Sub."""
    if settings.USE_PUBSUB and pubsub_client:
        try:
            await pubsub_client.publish_event(
                topic="session-events",
                event_type=event_type,
                data={
                    "session_id": session_id,
                    "user_id": user_id,
                    "timestamp": datetime.now(timezone.utc).isoformat(),
                    **data,
                },
            )
        except Exception as e:
            logger.error(f"Failed to publish session event: {e}")
