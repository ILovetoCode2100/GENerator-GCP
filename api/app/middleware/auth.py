"""
Authentication middleware for FastAPI.

This middleware handles API key and JWT authentication, user identification,
permission checking, and audit logging.
"""

from typing import Optional, List, Dict, Any
from fastapi import Request, HTTPException, status
from fastapi.security import APIKeyHeader, HTTPBearer, HTTPAuthorizationCredentials
from fastapi.responses import JSONResponse
import time
import aiohttp
import asyncio
from functools import lru_cache

from ..config import settings
from ..services.auth_service import auth_service, AuthUser, AuthMethod, Permission
from ..models.responses import ErrorResponse, ErrorType
from ..utils.logger import get_logger

# GCP imports
if settings.is_gcp_enabled:
    from ..gcp.firestore_client import FirestoreClient


logger = get_logger(__name__)

# Initialize GCP clients if enabled
firestore_client = None
functions_client = None

if settings.is_gcp_enabled:
    if settings.USE_FIRESTORE:
        firestore_client = FirestoreClient()
    # Cloud Functions client would be initialized here if we had one

# Security schemes
api_key_header = APIKeyHeader(
    name=settings.API_KEY_HEADER,
    auto_error=False,
    description="API key for authentication",
)

bearer_scheme = HTTPBearer(
    auto_error=False, description="JWT bearer token for authentication"
)

# Cache for validated API keys
_api_key_cache: Dict[str, tuple[Optional[AuthUser], float]] = {}
API_KEY_CACHE_TTL = 300  # 5 minutes


class AuthenticationError(HTTPException):
    """Custom authentication error with detailed response."""

    def __init__(
        self,
        detail: str,
        headers: Optional[dict] = None,
        request_id: Optional[str] = None,
    ):
        super().__init__(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail=detail,
            headers=headers or {"WWW-Authenticate": "Bearer"},
        )
        self.request_id = request_id


class AuthorizationError(HTTPException):
    """Custom authorization error with detailed response."""

    def __init__(
        self,
        detail: str,
        headers: Optional[dict] = None,
        request_id: Optional[str] = None,
    ):
        super().__init__(
            status_code=status.HTTP_403_FORBIDDEN, detail=detail, headers=headers
        )
        self.request_id = request_id


async def get_current_user(
    request: Request,
    api_key: Optional[str] = None,
    bearer_token: Optional[HTTPAuthorizationCredentials] = None,
) -> Optional[AuthUser]:
    """
    Get the current authenticated user from request.

    Tries multiple authentication methods in order:
    1. API key from header
    2. JWT token from Authorization header
    3. API key from query parameter (if enabled)

    Args:
        request: The FastAPI request
        api_key: API key from header
        bearer_token: Bearer token credentials

    Returns:
        Authenticated user or None
    """
    user = None

    # Try API key authentication first
    if api_key:
        # Check cache first
        cached_result = _api_key_cache.get(api_key)
        if cached_result:
            cached_user, cached_time = cached_result
            if time.time() - cached_time < API_KEY_CACHE_TTL:
                logger.debug(
                    f"Authenticated via cached API key: {cached_user.username if cached_user else 'None'}"
                )
                return cached_user

        # Try Firestore validation if enabled
        if settings.USE_FIRESTORE and firestore_client:
            try:
                # Check Firestore for API key
                key_data = await firestore_client.validate_api_key(api_key)
                if key_data:
                    # Create AuthUser from Firestore data
                    user = AuthUser(
                        user_id=key_data["user_id"],
                        username=key_data.get("username", key_data["user_id"]),
                        tenant_id=key_data.get("tenant_id", "default"),
                        permissions=[
                            Permission(p) for p in key_data.get("permissions", [])
                        ],
                        auth_method=AuthMethod.API_KEY,
                        api_key_id=api_key[:8] + "...",
                        metadata=key_data.get("metadata", {}),
                    )

                    # Cache the result
                    _api_key_cache[api_key] = (user, time.time())
                    logger.debug(
                        f"Authenticated via Firestore API key: {user.username}"
                    )
                    return user
                else:
                    # Cache negative result
                    _api_key_cache[api_key] = (None, time.time())
            except Exception as e:
                logger.error(f"Firestore API key validation failed: {e}")
                # Fall back to local validation

        # Fall back to local auth service
        user = await auth_service.validate_api_key(api_key)
        if user:
            # Cache the result
            _api_key_cache[api_key] = (user, time.time())
            logger.debug(f"Authenticated via local API key: {user.username}")
            return user
        else:
            # Cache negative result
            _api_key_cache[api_key] = (None, time.time())

    # Try JWT bearer token
    if bearer_token and bearer_token.credentials:
        # Could add Cloud Function validation here for JWT tokens
        # For now, use local validation
        user = await auth_service.validate_jwt_token(bearer_token.credentials)
        if user:
            logger.debug(f"Authenticated via JWT: {user.username}")
            return user

    # Try API key from query parameter (only if explicitly enabled)
    if getattr(settings, "ALLOW_API_KEY_IN_QUERY", False):
        query_api_key = request.query_params.get("api_key")
        if query_api_key:
            user = await auth_service.validate_api_key(query_api_key)
            if user:
                logger.debug(f"Authenticated via query parameter: {user.username}")
                return user

    return None


async def authenticate_request(
    request: Request, required: bool = True
) -> Optional[AuthUser]:
    """
    Authenticate a request and return the user.

    Args:
        request: The FastAPI request
        required: Whether authentication is required

    Returns:
        Authenticated user or None (if not required)

    Raises:
        AuthenticationError: If authentication fails and is required
    """
    # Get request ID for error tracking
    request_id = (
        request.state.request_id if hasattr(request.state, "request_id") else None
    )

    # Try to get API key and bearer token
    try:
        api_key = await api_key_header(request)
        bearer_token = await bearer_scheme(request)
    except Exception as e:
        logger.error(f"Error extracting auth credentials: {e}")
        api_key = None
        bearer_token = None

    # Get authenticated user
    user = await get_current_user(request, api_key, bearer_token)

    # Check if authentication is required
    if required and not user:
        logger.warning(f"Unauthenticated request to {request.url.path}")
        raise AuthenticationError(
            detail="Authentication required. Please provide a valid API key or JWT token.",
            request_id=request_id,
        )

    # Store user in request state for later use
    if user:
        request.state.user = user

        # Log authenticated request
        # Use background task to not block request
        asyncio.create_task(log_authenticated_request(user=user, request=request))

    return user


async def require_permission(request: Request, permission: Permission) -> AuthUser:
    """
    Require a specific permission for the request.

    Args:
        request: The FastAPI request
        permission: Required permission

    Returns:
        Authenticated user with permission

    Raises:
        AuthenticationError: If not authenticated
        AuthorizationError: If lacking permission
    """
    # Get authenticated user
    user = await authenticate_request(request, required=True)

    # Check permission
    if not await auth_service.check_permission(user, permission):
        logger.warning(
            f"Permission denied for user {user.username}: "
            f"required {permission}, has {user.permissions}"
        )
        raise AuthorizationError(
            detail=f"Insufficient permissions. Required: {permission.value}",
            request_id=request.state.request_id
            if hasattr(request.state, "request_id")
            else None,
        )

    return user


async def require_command_permission(request: Request, command_group: str) -> AuthUser:
    """
    Require permission for a command group.

    Args:
        request: The FastAPI request
        command_group: The command group

    Returns:
        Authenticated user with permission

    Raises:
        AuthenticationError: If not authenticated
        AuthorizationError: If lacking permission
    """
    # Get authenticated user
    user = await authenticate_request(request, required=True)

    # Check command group permission
    if not await auth_service.check_command_group_permission(user, command_group):
        logger.warning(
            f"Command permission denied for user {user.username}: "
            f"command group {command_group}"
        )
        raise AuthorizationError(
            detail=f"Insufficient permissions for command group: {command_group}",
            request_id=request.state.request_id
            if hasattr(request.state, "request_id")
            else None,
        )

    return user


class AuthMiddleware:
    """
    Authentication middleware for automatic auth on specific routes.
    """

    def __init__(
        self,
        app,
        exclude_paths: Optional[List[str]] = None,
        require_auth_paths: Optional[List[str]] = None,
    ):
        """
        Initialize authentication middleware.

        Args:
            app: The FastAPI application
            exclude_paths: Paths to exclude from authentication
            require_auth_paths: Paths that require authentication (if None, all paths)
        """
        self.app = app
        self.exclude_paths = exclude_paths or ["/health", "/docs", "/openapi.json", "/"]
        self.require_auth_paths = require_auth_paths

    async def __call__(self, scope, receive, send):
        """Process the request through authentication."""
        if scope["type"] != "http":
            await self.app(scope, receive, send)
            return

        # Create request object for easier handling
        request = Request(scope, receive)
        start_time = time.time()

        # Generate request ID if not present
        if not hasattr(request.state, "request_id"):
            request.state.request_id = f"req_{int(time.time() * 1000)}"

        # Check if path is excluded
        path = request.url.path
        if any(path.startswith(exclude) for exclude in self.exclude_paths):
            return await call_next(request)

        # Check if authentication is required for this path
        require_auth = True
        if self.require_auth_paths is not None:
            require_auth = any(
                path.startswith(auth_path) for auth_path in self.require_auth_paths
            )

        try:
            # Authenticate request
            user = await authenticate_request(request, required=require_auth)

            # Add auth headers to response
            response = await call_next(request)

            if user:
                response.headers["X-Authenticated-User"] = user.username
                response.headers["X-Tenant-ID"] = user.tenant_id

            # Add timing header
            process_time = time.time() - start_time
            response.headers["X-Process-Time"] = str(process_time)

            return response

        except (AuthenticationError, AuthorizationError) as e:
            # Return structured error response
            error_response = ErrorResponse(
                error_type=ErrorType.AUTHENTICATION
                if isinstance(e, AuthenticationError)
                else ErrorType.AUTHORIZATION,
                message=e.detail,
                request_id=e.request_id,
            )

            return JSONResponse(
                status_code=e.status_code,
                content=error_response.model_dump(),
                headers=e.headers,
            )

        except Exception as e:
            logger.error(f"Unexpected error in auth middleware: {e}")
            error_response = ErrorResponse(
                error_type=ErrorType.SERVER_ERROR,
                message="Internal authentication error",
                request_id=request.state.request_id,
            )

            return JSONResponse(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                content=error_response.model_dump(),
            )


# Dependency functions for route protection
async def get_authenticated_user(request: Request) -> AuthUser:
    """Dependency to get authenticated user (required)."""
    # Check if authentication is disabled
    if settings.AUTH_ENABLED == "false" or settings.SKIP_AUTH == "true":
        # Return a dummy user for testing
        return AuthUser(
            id="test-user",
            name="Test User",
            email="test@example.com",
            api_key="test-key",
            auth_method=AuthMethod.API_KEY,
            permissions=[Permission.READ, Permission.WRITE, Permission.ADMIN],
            created_at=time.time(),
            last_seen=time.time(),
        )
    return await authenticate_request(request, required=True)


async def get_optional_user(request: Request) -> Optional[AuthUser]:
    """Dependency to get authenticated user (optional)."""
    return await authenticate_request(request, required=False)


def require_permissions(*permissions: Permission):
    """
    Dependency factory for requiring specific permissions.

    Args:
        *permissions: Required permissions (user must have at least one)

    Returns:
        Dependency function
    """

    async def permission_checker(request: Request) -> AuthUser:
        user = await authenticate_request(request, required=True)

        # Check if user has any of the required permissions
        has_permission = any(
            await auth_service.check_permission(user, perm) for perm in permissions
        )

        if not has_permission:
            raise AuthorizationError(
                detail=f"Insufficient permissions. Required one of: {[p.value for p in permissions]}",
                request_id=request.state.request_id
                if hasattr(request.state, "request_id")
                else None,
            )

        return user

    return permission_checker


async def log_authenticated_request(user: AuthUser, request: Request):
    """Log authenticated request in background."""
    try:
        # Log to local auth service
        await auth_service.audit_log(
            user=user,
            action="api_request",
            resource=str(request.url.path),
            details={
                "method": request.method,
                "client_host": request.client.host if request.client else None,
                "user_agent": request.headers.get("user-agent"),
            },
        )

        # Also log to Firestore if enabled for analytics
        if settings.USE_FIRESTORE and firestore_client:
            try:
                await firestore_client.log_api_request(
                    user_id=user.user_id,
                    path=str(request.url.path),
                    method=request.method,
                    client_host=request.client.host if request.client else None,
                    user_agent=request.headers.get("user-agent"),
                )
            except Exception as e:
                logger.error(f"Failed to log request to Firestore: {e}")

    except Exception as e:
        logger.error(f"Failed to log authenticated request: {e}")


async def validate_with_cloud_function(api_key: str) -> Optional[Dict[str, Any]]:
    """
    Validate API key using Cloud Function for fast validation.

    This would call a lightweight Cloud Function that checks
    API keys with minimal latency.
    """
    if not settings.USE_CLOUD_FUNCTIONS:
        return None

    try:
        # Example Cloud Function call
        function_url = f"https://{settings.GCP_LOCATION}-{settings.GCP_PROJECT_ID}.cloudfunctions.net/auth-validator"

        async with aiohttp.ClientSession() as session:
            async with session.post(
                function_url,
                json={"api_key": api_key},
                headers={
                    "Authorization": f"Bearer {await get_function_token()}",
                    "Content-Type": "application/json",
                },
                timeout=aiohttp.ClientTimeout(total=2.0),  # Fast timeout
            ) as response:
                if response.status == 200:
                    return await response.json()
                elif response.status == 401:
                    return None
                else:
                    logger.warning(
                        f"Cloud Function validation failed with status {response.status}"
                    )
                    return None

    except asyncio.TimeoutError:
        logger.warning("Cloud Function validation timed out")
        return None
    except Exception as e:
        logger.error(f"Cloud Function validation error: {e}")
        return None


@lru_cache(maxsize=1)
async def get_function_token() -> str:
    """
    Get authentication token for Cloud Functions.

    In production, this would use proper service account authentication.
    """
    # This is a placeholder - in production, use google-auth library
    # to get proper authentication tokens
    return "dummy-token"


def clear_api_key_cache():
    """Clear the API key cache."""
    global _api_key_cache
    _api_key_cache.clear()
    logger.info("API key cache cleared")


def require_command_group(command_group: str):
    """
    Dependency factory for requiring command group permission.

    Args:
        command_group: The command group

    Returns:
        Dependency function
    """

    async def command_permission_checker(request: Request) -> AuthUser:
        return await require_command_permission(request, command_group)

    return command_permission_checker
