"""
Authentication service for API key and user management.

This service handles API key validation, user/tenant identification,
permission checking, and token management.
"""

import hashlib
import hmac
import secrets
import time
from typing import Optional, Dict, List, Tuple, Any
from datetime import datetime, timedelta
from enum import Enum
import jwt
from pydantic import BaseModel, Field

from ..config import settings
from ..utils.logger import get_logger


logger = get_logger(__name__)


class AuthMethod(str, Enum):
    """Authentication methods"""

    API_KEY = "api_key"
    JWT_TOKEN = "jwt_token"
    SESSION_TOKEN = "session_token"


class Permission(str, Enum):
    """Permission levels for different operations"""

    # Read permissions
    READ_PROJECTS = "read:projects"
    READ_TESTS = "read:tests"
    READ_EXECUTIONS = "read:executions"
    READ_LIBRARY = "read:library"
    READ_TEMPLATES = "read:templates"
    READ_ANALYTICS = "read:analytics"
    READ_WEBHOOKS = "read:webhooks"

    # Write permissions
    WRITE_PROJECTS = "write:projects"
    WRITE_TESTS = "write:tests"
    WRITE_EXECUTIONS = "write:executions"
    WRITE_LIBRARY = "write:library"

    # Special permissions
    GENERATE_REPORTS = "generate:reports"

    # Admin permissions
    ADMIN_USERS = "admin:users"
    ADMIN_SETTINGS = "admin:settings"
    ADMIN_ALL = "admin:*"


class UserRole(str, Enum):
    """User roles"""

    VIEWER = "viewer"
    DEVELOPER = "developer"
    ADMIN = "admin"
    SERVICE_ACCOUNT = "service_account"


# Role-based permission mappings
ROLE_PERMISSIONS: Dict[UserRole, List[Permission]] = {
    UserRole.VIEWER: [
        Permission.READ_PROJECTS,
        Permission.READ_TESTS,
        Permission.READ_EXECUTIONS,
        Permission.READ_LIBRARY,
        Permission.READ_TEMPLATES,
    ],
    UserRole.DEVELOPER: [
        # All viewer permissions
        Permission.READ_PROJECTS,
        Permission.READ_TESTS,
        Permission.READ_EXECUTIONS,
        Permission.READ_LIBRARY,
        Permission.READ_TEMPLATES,
        # Plus write permissions
        Permission.WRITE_PROJECTS,
        Permission.WRITE_TESTS,
        Permission.WRITE_EXECUTIONS,
        Permission.WRITE_LIBRARY,
    ],
    UserRole.ADMIN: [
        # All permissions
        Permission.ADMIN_ALL,
    ],
    UserRole.SERVICE_ACCOUNT: [
        # Customizable permissions, typically all read/write
        Permission.READ_PROJECTS,
        Permission.READ_TESTS,
        Permission.READ_EXECUTIONS,
        Permission.READ_LIBRARY,
        Permission.READ_TEMPLATES,
        Permission.WRITE_PROJECTS,
        Permission.WRITE_TESTS,
        Permission.WRITE_EXECUTIONS,
        Permission.WRITE_LIBRARY,
    ],
}


class AuthUser(BaseModel):
    """Authenticated user information"""

    user_id: str = Field(..., description="Unique user identifier")
    tenant_id: str = Field(..., description="Tenant/organization ID")
    username: str = Field(..., description="Username")
    email: Optional[str] = Field(None, description="User email")
    role: UserRole = Field(..., description="User role")
    permissions: List[Permission] = Field(..., description="User permissions")
    auth_method: AuthMethod = Field(..., description="Authentication method used")
    api_key_id: Optional[str] = Field(None, description="API key ID if applicable")
    metadata: Dict[str, Any] = Field(
        default_factory=dict, description="Additional user metadata"
    )

    async def has_permission(self, permission: Permission) -> bool:
        """Check if user has a specific permission."""
        # Admin users have all permissions
        if Permission.ADMIN_ALL in self.permissions:
            return True
        # Check specific permission
        return permission in self.permissions

    @property
    def is_admin(self) -> bool:
        """Check if user is an admin."""
        return self.role == UserRole.ADMIN or Permission.ADMIN_ALL in self.permissions


class APIKey(BaseModel):
    """API key model"""

    key_id: str = Field(..., description="API key ID")
    key_hash: str = Field(..., description="Hashed API key")
    user_id: str = Field(..., description="Associated user ID")
    tenant_id: str = Field(..., description="Associated tenant ID")
    name: str = Field(..., description="Key name/description")
    role: UserRole = Field(..., description="Role for this key")
    permissions: List[Permission] = Field(..., description="Custom permissions")
    created_at: datetime = Field(..., description="Creation time")
    expires_at: Optional[datetime] = Field(None, description="Expiration time")
    last_used: Optional[datetime] = Field(None, description="Last usage time")
    usage_count: int = Field(0, description="Usage count")
    is_active: bool = Field(True, description="Whether key is active")
    metadata: Dict[str, Any] = Field(
        default_factory=dict, description="Additional metadata"
    )


class AuthService:
    """
    Authentication service for managing API keys and user authentication.

    In a production environment, this would interface with a database.
    For now, it uses in-memory storage and configuration-based keys.
    """

    def __init__(self):
        self._api_keys: Dict[str, APIKey] = {}
        self._users: Dict[str, AuthUser] = {}
        self._initialize_default_keys()

    def _initialize_default_keys(self):
        """Initialize API keys from configuration."""
        # Create default admin user
        admin_user = AuthUser(
            user_id="admin",
            tenant_id=settings.VIRTUOSO_ORG_ID or "default",
            username="admin",
            email="admin@virtuoso.local",
            role=UserRole.ADMIN,
            permissions=[Permission.ADMIN_ALL],
            auth_method=AuthMethod.API_KEY,
        )
        self._users[admin_user.user_id] = admin_user

        # Initialize configured API keys
        for idx, api_key in enumerate(settings.API_KEYS):
            if not api_key:
                continue

            # Generate a key ID
            key_id = f"key_{idx + 1}"

            # Hash the API key
            key_hash = self._hash_api_key(api_key)

            # Determine role based on key pattern (example logic)
            role = UserRole.ADMIN if "admin" in api_key.lower() else UserRole.DEVELOPER

            # Create API key record
            api_key_record = APIKey(
                key_id=key_id,
                key_hash=key_hash,
                user_id="admin" if role == UserRole.ADMIN else f"user_{idx + 1}",
                tenant_id=settings.VIRTUOSO_ORG_ID or "default",
                name=f"Default API Key {idx + 1}",
                role=role,
                permissions=ROLE_PERMISSIONS[role],
                created_at=datetime.utcnow(),
                is_active=True,
            )

            self._api_keys[api_key] = api_key_record

            # Create associated user if not admin
            if api_key_record.user_id != "admin":
                user = AuthUser(
                    user_id=api_key_record.user_id,
                    tenant_id=api_key_record.tenant_id,
                    username=f"api_user_{idx + 1}",
                    role=role,
                    permissions=api_key_record.permissions,
                    auth_method=AuthMethod.API_KEY,
                    api_key_id=key_id,
                )
                self._users[user.user_id] = user

    def _hash_api_key(self, api_key: str) -> str:
        """Hash an API key for secure storage."""
        return hashlib.sha256(api_key.encode()).hexdigest()

    def _verify_api_key_hash(self, api_key: str, key_hash: str) -> bool:
        """Verify an API key against its hash."""
        return hmac.compare_digest(self._hash_api_key(api_key), key_hash)

    async def validate_api_key(self, api_key: str) -> Optional[AuthUser]:
        """
        Validate an API key and return authenticated user information.

        Args:
            api_key: The API key to validate

        Returns:
            AuthUser if valid, None otherwise
        """
        try:
            # Check if key exists
            api_key_record = self._api_keys.get(api_key)
            if not api_key_record:
                logger.warning(f"Invalid API key attempted: {api_key[:8]}...")
                return None

            # Check if key is active
            if not api_key_record.is_active:
                logger.warning(f"Inactive API key used: {api_key_record.key_id}")
                return None

            # Check expiration
            if (
                api_key_record.expires_at
                and api_key_record.expires_at < datetime.utcnow()
            ):
                logger.warning(f"Expired API key used: {api_key_record.key_id}")
                return None

            # Update usage stats
            api_key_record.last_used = datetime.utcnow()
            api_key_record.usage_count += 1

            # Get associated user
            user = self._users.get(api_key_record.user_id)
            if not user:
                logger.error(f"User not found for API key: {api_key_record.key_id}")
                return None

            # Return authenticated user with key info
            return user.model_copy(update={"api_key_id": api_key_record.key_id})

        except Exception as e:
            logger.error(f"Error validating API key: {e}")
            return None

    async def validate_jwt_token(self, token: str) -> Optional[AuthUser]:
        """
        Validate a JWT token and return authenticated user information.

        Args:
            token: The JWT token to validate

        Returns:
            AuthUser if valid, None otherwise
        """
        try:
            # Decode and verify JWT
            payload = jwt.decode(token, settings.SECRET_KEY, algorithms=["HS256"])

            # Extract user information
            user_id = payload.get("sub")
            tenant_id = payload.get("tenant_id")
            role = UserRole(payload.get("role", UserRole.VIEWER))

            if not user_id or not tenant_id:
                logger.warning("Invalid JWT payload: missing user_id or tenant_id")
                return None

            # Create authenticated user
            return AuthUser(
                user_id=user_id,
                tenant_id=tenant_id,
                username=payload.get("username", user_id),
                email=payload.get("email"),
                role=role,
                permissions=ROLE_PERMISSIONS[role],
                auth_method=AuthMethod.JWT_TOKEN,
                metadata=payload.get("metadata", {}),
            )

        except jwt.ExpiredSignatureError:
            logger.warning("Expired JWT token")
            return None
        except jwt.InvalidTokenError as e:
            logger.warning(f"Invalid JWT token: {e}")
            return None
        except Exception as e:
            logger.error(f"Error validating JWT token: {e}")
            return None

    async def check_permission(
        self, user: AuthUser, required_permission: Permission
    ) -> bool:
        """
        Check if a user has a specific permission.

        Args:
            user: The authenticated user
            required_permission: The required permission

        Returns:
            True if user has permission, False otherwise
        """
        # Admin users have all permissions
        if Permission.ADMIN_ALL in user.permissions:
            return True

        # Check specific permission
        return required_permission in user.permissions

    async def check_command_group_permission(
        self, user: AuthUser, command_group: str
    ) -> bool:
        """
        Check if a user has permission for a command group.

        Args:
            user: The authenticated user
            command_group: The command group (e.g., "step-interact", "create-project")

        Returns:
            True if user has permission, False otherwise
        """
        # Define command group to permission mappings
        command_permissions = {
            # Read operations
            "list-projects": Permission.READ_PROJECTS,
            "list-goals": Permission.READ_PROJECTS,
            "list-journeys": Permission.READ_PROJECTS,
            "list-checkpoints": Permission.READ_PROJECTS,
            "list-steps": Permission.READ_TESTS,
            "list-executions": Permission.READ_EXECUTIONS,
            "list-templates": Permission.READ_TEMPLATES,
            "get-library": Permission.READ_LIBRARY,
            # Write operations - project management
            "create-project": Permission.WRITE_PROJECTS,
            "create-goal": Permission.WRITE_PROJECTS,
            "create-journey": Permission.WRITE_PROJECTS,
            "create-checkpoint": Permission.WRITE_PROJECTS,
            # Write operations - test steps
            "step-assert": Permission.WRITE_TESTS,
            "step-interact": Permission.WRITE_TESTS,
            "step-navigate": Permission.WRITE_TESTS,
            "step-window": Permission.WRITE_TESTS,
            "step-data": Permission.WRITE_TESTS,
            "step-dialog": Permission.WRITE_TESTS,
            "step-wait": Permission.WRITE_TESTS,
            "step-file": Permission.WRITE_TESTS,
            "step-misc": Permission.WRITE_TESTS,
            "run-test": Permission.WRITE_TESTS,
            # Library operations
            "library": Permission.WRITE_LIBRARY,
            # Execution operations
            "execute-goal": Permission.WRITE_EXECUTIONS,
            "monitor-execution": Permission.READ_EXECUTIONS,
            "get-execution-analysis": Permission.READ_EXECUTIONS,
            "create-environment": Permission.WRITE_EXECUTIONS,
        }

        # Get required permission for command group
        required_permission = command_permissions.get(command_group)
        if not required_permission:
            # Unknown command group - deny by default
            logger.warning(f"Unknown command group: {command_group}")
            return False

        return await self.check_permission(user, required_permission)

    async def generate_jwt_token(self, user: AuthUser, expires_in: int = 3600) -> str:
        """
        Generate a JWT token for a user.

        Args:
            user: The user to generate token for
            expires_in: Token expiration time in seconds

        Returns:
            JWT token string
        """
        payload = {
            "sub": user.user_id,
            "tenant_id": user.tenant_id,
            "username": user.username,
            "email": user.email,
            "role": user.role.value,
            "exp": datetime.utcnow() + timedelta(seconds=expires_in),
            "iat": datetime.utcnow(),
            "metadata": user.metadata,
        }

        return jwt.encode(payload, settings.SECRET_KEY, algorithm="HS256")

    async def create_api_key(
        self,
        user_id: str,
        tenant_id: str,
        name: str,
        role: UserRole = UserRole.DEVELOPER,
        permissions: Optional[List[Permission]] = None,
        expires_in_days: Optional[int] = None,
    ) -> Tuple[str, APIKey]:
        """
        Create a new API key.

        Args:
            user_id: User ID to associate with key
            tenant_id: Tenant ID
            name: Key name/description
            role: User role
            permissions: Custom permissions (uses role defaults if None)
            expires_in_days: Expiration time in days

        Returns:
            Tuple of (raw_api_key, api_key_record)
        """
        # Generate secure random key
        raw_api_key = f"vrt_{secrets.token_urlsafe(32)}"

        # Create API key record
        api_key_record = APIKey(
            key_id=f"key_{int(time.time())}",
            key_hash=self._hash_api_key(raw_api_key),
            user_id=user_id,
            tenant_id=tenant_id,
            name=name,
            role=role,
            permissions=permissions or ROLE_PERMISSIONS[role],
            created_at=datetime.utcnow(),
            expires_at=(
                datetime.utcnow() + timedelta(days=expires_in_days)
                if expires_in_days
                else None
            ),
            is_active=True,
        )

        # Store the key
        self._api_keys[raw_api_key] = api_key_record

        return raw_api_key, api_key_record

    async def revoke_api_key(self, key_id: str) -> bool:
        """
        Revoke an API key.

        Args:
            key_id: The key ID to revoke

        Returns:
            True if revoked, False if not found
        """
        for api_key, record in self._api_keys.items():
            if record.key_id == key_id:
                record.is_active = False
                logger.info(f"Revoked API key: {key_id}")
                return True

        return False

    async def audit_log(
        self,
        user: AuthUser,
        action: str,
        resource: str,
        details: Optional[Dict[str, Any]] = None,
    ):
        """
        Log an authenticated action for auditing.

        Args:
            user: The authenticated user
            action: The action performed
            resource: The resource accessed
            details: Additional details
        """
        log_entry = {
            "timestamp": datetime.utcnow().isoformat(),
            "user_id": user.user_id,
            "tenant_id": user.tenant_id,
            "username": user.username,
            "auth_method": user.auth_method.value,
            "api_key_id": user.api_key_id,
            "action": action,
            "resource": resource,
            "details": details or {},
        }

        # In production, this would write to an audit log database
        logger.info(f"Audit log: {log_entry}")


# Create singleton instance
auth_service = AuthService()
