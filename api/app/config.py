"""
Application configuration using Pydantic settings.
"""

import os
from typing import List, Optional
from pydantic_settings import BaseSettings, SettingsConfigDict
from pydantic import Field, validator


class Settings(BaseSettings):
    """
    Application settings with environment variable support.
    """

    # Application settings
    APP_NAME: str = "Virtuoso API CLI"
    VERSION: str = "1.0.0"
    ENVIRONMENT: str = Field(default="production", env="ENVIRONMENT")
    DEBUG: bool = Field(default=False, env="DEBUG")

    # Server settings
    HOST: str = Field(default="0.0.0.0", env="HOST")
    PORT: int = Field(default=8000, env="PORT")
    WORKERS: int = Field(default=4, env="WORKERS")

    # API settings
    API_PREFIX: str = "/api/v1"
    API_KEY_HEADER: str = "X-API-Key"

    # Security
    AUTH_ENABLED: str = Field(default="true", env="AUTH_ENABLED")
    SKIP_AUTH: str = Field(default="false", env="SKIP_AUTH")
    API_KEYS: List[str] = Field(default_factory=list, env="API_KEYS")
    SECRET_KEY: str = Field(default="your-secret-key-here", env="SECRET_KEY")

    # CORS settings
    CORS_ORIGINS: List[str] = Field(default=["*"], env="CORS_ORIGINS")

    # CLI settings
    CLI_PATH: str = Field(default="./bin/api-cli", env="CLI_PATH")
    CLI_CONFIG_PATH: Optional[str] = Field(default=None, env="CLI_CONFIG_PATH")
    CLI_TIMEOUT: int = Field(default=300, env="CLI_TIMEOUT")  # 5 minutes

    # Virtuoso API settings
    VIRTUOSO_API_KEY: Optional[str] = Field(default=None, env="VIRTUOSO_API_KEY")
    VIRTUOSO_BASE_URL: str = Field(
        default="https://api-app2.virtuoso.qa/api", env="VIRTUOSO_BASE_URL"
    )
    VIRTUOSO_ORG_ID: Optional[str] = Field(default=None, env="VIRTUOSO_ORG_ID")

    # Logging
    LOG_LEVEL: str = Field(default="INFO", env="LOG_LEVEL")
    LOG_FORMAT: str = Field(
        default="%(asctime)s - %(name)s - %(levelname)s - %(message)s", env="LOG_FORMAT"
    )

    # Rate limiting
    RATE_LIMIT_ENABLED: bool = Field(default=True, env="RATE_LIMIT_ENABLED")
    RATE_LIMIT_REQUESTS: int = Field(default=100, env="RATE_LIMIT_REQUESTS")
    RATE_LIMIT_PERIOD: int = Field(default=60, env="RATE_LIMIT_PERIOD")  # seconds

    # Redis settings (for rate limiting and caching)
    REDIS_URL: str = Field(default="redis://localhost:6379", env="REDIS_URL")
    REDIS_PASSWORD: Optional[str] = Field(default=None, env="REDIS_PASSWORD")
    REDIS_DB: int = Field(default=0, env="REDIS_DB")
    REDIS_TIMEOUT: int = Field(default=5, env="REDIS_TIMEOUT")  # seconds

    # Session management
    SESSION_TIMEOUT: int = Field(default=3600, env="SESSION_TIMEOUT")  # 1 hour
    SESSION_MAX_AGE: int = Field(default=86400, env="SESSION_MAX_AGE")  # 24 hours

    # File upload
    MAX_UPLOAD_SIZE: int = Field(
        default=10 * 1024 * 1024, env="MAX_UPLOAD_SIZE"
    )  # 10MB
    ALLOWED_UPLOAD_EXTENSIONS: List[str] = Field(
        default=[".yaml", ".yml", ".json"], env="ALLOWED_UPLOAD_EXTENSIONS"
    )

    # Cache settings
    CACHE_ENABLED: bool = Field(default=True, env="CACHE_ENABLED")
    CACHE_TTL: int = Field(default=300, env="CACHE_TTL")  # 5 minutes

    # GCP Settings
    GCP_PROJECT_ID: Optional[str] = Field(default=None, env="GCP_PROJECT_ID")
    GCP_LOCATION: str = Field(default="us-central1", env="GCP_LOCATION")
    GCP_SERVICE_ACCOUNT_EMAIL: Optional[str] = Field(
        default=None, env="GCP_SERVICE_ACCOUNT_EMAIL"
    )
    PUBSUB_PUSH_TOKEN: Optional[str] = Field(default=None, env="PUBSUB_PUSH_TOKEN")

    # GCP Feature Flags
    USE_FIRESTORE: bool = Field(default=False, env="USE_FIRESTORE")
    USE_CLOUD_TASKS: bool = Field(default=False, env="USE_CLOUD_TASKS")
    USE_PUBSUB: bool = Field(default=False, env="USE_PUBSUB")
    USE_SECRET_MANAGER: bool = Field(default=False, env="USE_SECRET_MANAGER")
    USE_CLOUD_STORAGE: bool = Field(default=False, env="USE_CLOUD_STORAGE")
    USE_CLOUD_MONITORING: bool = Field(default=False, env="USE_CLOUD_MONITORING")
    USE_BIGQUERY: bool = Field(default=False, env="USE_BIGQUERY")
    USE_CLOUD_FUNCTIONS: bool = Field(default=False, env="USE_CLOUD_FUNCTIONS")

    # GCP Emulator Settings (for local development)
    FIRESTORE_EMULATOR_HOST: Optional[str] = Field(
        default=None, env="FIRESTORE_EMULATOR_HOST"
    )
    PUBSUB_EMULATOR_HOST: Optional[str] = Field(
        default=None, env="PUBSUB_EMULATOR_HOST"
    )

    # Secret Manager Settings
    SECRET_CACHE_TTL_MINUTES: int = Field(default=30, env="SECRET_CACHE_TTL_MINUTES")

    # API Base URL (for Cloud Tasks callbacks)
    API_BASE_URL: str = Field(default="http://localhost:8000", env="API_BASE_URL")

    @validator("API_KEYS", pre=True)
    def parse_api_keys(cls, v):
        """Parse comma-separated API keys."""
        if isinstance(v, str):
            return [key.strip() for key in v.split(",") if key.strip()]
        return v

    @validator("CORS_ORIGINS", pre=True)
    def parse_cors_origins(cls, v):
        """Parse comma-separated CORS origins."""
        if isinstance(v, str):
            return [origin.strip() for origin in v.split(",") if origin.strip()]
        return v

    @validator("ALLOWED_UPLOAD_EXTENSIONS", pre=True)
    def parse_upload_extensions(cls, v):
        """Parse comma-separated upload extensions."""
        if isinstance(v, str):
            return [ext.strip() for ext in v.split(",") if ext.strip()]
        return v

    @validator("CLI_PATH")
    def validate_cli_path(cls, v):
        """Ensure CLI path exists."""
        if not os.path.exists(v):
            # Try relative to project root
            project_root = os.path.dirname(os.path.dirname(os.path.dirname(__file__)))
            alt_path = os.path.join(project_root, v)
            if os.path.exists(alt_path):
                return alt_path
            raise ValueError(f"CLI binary not found at: {v}")
        return v

    model_config = SettingsConfigDict(
        env_file=".env", env_file_encoding="utf-8", case_sensitive=True, extra="ignore"
    )

    def get_cli_env(self) -> dict:
        """
        Get environment variables for CLI execution.
        """
        env = os.environ.copy()

        if self.VIRTUOSO_API_KEY:
            env["VIRTUOSO_API_KEY"] = self.VIRTUOSO_API_KEY

        if self.VIRTUOSO_ORG_ID:
            env["VIRTUOSO_ORG_ID"] = self.VIRTUOSO_ORG_ID

        if self.CLI_CONFIG_PATH:
            env["VIRTUOSO_CONFIG_PATH"] = self.CLI_CONFIG_PATH

        return env

    @property
    def is_production(self) -> bool:
        """Check if running in production."""
        return self.ENVIRONMENT.lower() == "production"

    @property
    def is_development(self) -> bool:
        """Check if running in development."""
        return self.ENVIRONMENT.lower() in ("development", "dev")

    @property
    def is_gcp_enabled(self) -> bool:
        """Check if any GCP services are enabled."""
        return any(
            [
                self.USE_FIRESTORE,
                self.USE_CLOUD_TASKS,
                self.USE_PUBSUB,
                self.USE_SECRET_MANAGER,
                self.USE_CLOUD_STORAGE,
                self.USE_CLOUD_MONITORING,
            ]
        )

    @property
    def is_local_gcp(self) -> bool:
        """Check if using GCP emulators for local development."""
        return bool(self.FIRESTORE_EMULATOR_HOST or self.PUBSUB_EMULATOR_HOST)

    def validate_gcp_settings(self):
        """Validate GCP settings."""
        if self.is_gcp_enabled and not self.GCP_PROJECT_ID:
            raise ValueError("GCP_PROJECT_ID is required when GCP services are enabled")


# Create settings instance
settings = Settings()
