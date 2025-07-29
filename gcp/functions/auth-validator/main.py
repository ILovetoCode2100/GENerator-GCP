"""
Cloud Function for fast authentication validation and rate limiting.
"""

import json
import asyncio
import os
from typing import Dict, Any, Optional, Tuple
from datetime import datetime
import hashlib
import redis.asyncio as redis
from google.cloud import firestore, secretmanager
from flask import Request, Response
import logging

# Configure structured logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Environment variables
PROJECT_ID = os.environ.get("GCP_PROJECT", "virtuoso-api")
REDIS_HOST = os.environ.get("REDIS_HOST", "10.0.0.3")
REDIS_PORT = int(os.environ.get("REDIS_PORT", "6379"))
CACHE_TTL_SECONDS = int(os.environ.get("CACHE_TTL_SECONDS", "300"))  # 5 minutes
RATE_LIMIT_WINDOW = int(os.environ.get("RATE_LIMIT_WINDOW", "3600"))  # 1 hour
DEFAULT_RATE_LIMIT = int(
    os.environ.get("DEFAULT_RATE_LIMIT", "1000")
)  # requests per hour


class AuthValidator:
    """Validates API keys and manages rate limiting."""

    def __init__(self):
        self.db = firestore.AsyncClient()
        self.redis_client = None
        self.secret_client = secretmanager.SecretManagerServiceAsyncClient()

    async def _get_redis(self) -> redis.Redis:
        """Get or create Redis connection."""
        if not self.redis_client:
            self.redis_client = await redis.from_url(
                f"redis://{REDIS_HOST}:{REDIS_PORT}", decode_responses=True
            )
        return self.redis_client

    async def _close_redis(self):
        """Close Redis connection."""
        if self.redis_client:
            await self.redis_client.close()

    def _hash_api_key(self, api_key: str) -> str:
        """Hash API key for secure storage."""
        return hashlib.sha256(api_key.encode()).hexdigest()

    async def validate_api_key(
        self, api_key: str
    ) -> Tuple[bool, Optional[Dict[str, Any]]]:
        """
        Validate API key and return user information.

        Returns:
            Tuple of (is_valid, user_info)
        """
        if not api_key:
            return False, None

        try:
            # Hash the API key
            hashed_key = self._hash_api_key(api_key)

            # Check Redis cache first
            r = await self._get_redis()
            cache_key = f"auth:key:{hashed_key}"
            cached_data = await r.get(cache_key)

            if cached_data:
                logger.info("API key found in cache")
                return True, json.loads(cached_data)

            # Check Firestore
            api_keys_ref = self.db.collection("api_keys")
            query = (
                api_keys_ref.where("key_hash", "==", hashed_key)
                .where("active", "==", True)
                .limit(1)
            )

            docs = []
            async for doc in query.stream():
                docs.append(doc)

            if not docs:
                logger.warning(f"Invalid API key attempted: {hashed_key[:8]}...")
                return False, None

            # Get user information
            key_data = docs[0].to_dict()
            user_id = key_data.get("user_id")

            if not user_id:
                logger.error("API key has no associated user")
                return False, None

            # Get user details
            user_doc = await self.db.collection("users").document(user_id).get()
            if not user_doc.exists:
                logger.error(f"User {user_id} not found for API key")
                return False, None

            user_data = user_doc.to_dict()

            # Prepare user info
            user_info = {
                "user_id": user_id,
                "email": user_data.get("email"),
                "plan": user_data.get("plan", "free"),
                "rate_limit": user_data.get("rate_limit", DEFAULT_RATE_LIMIT),
                "permissions": user_data.get("permissions", []),
                "organization_id": user_data.get("organization_id"),
                "key_name": key_data.get("name", "default"),
                "key_created": key_data.get("created_at").isoformat()
                if key_data.get("created_at")
                else None,
            }

            # Update last used timestamp
            await docs[0].reference.update(
                {
                    "last_used": firestore.SERVER_TIMESTAMP,
                    "usage_count": firestore.Increment(1),
                }
            )

            # Cache the result
            await r.set(cache_key, json.dumps(user_info), ex=CACHE_TTL_SECONDS)

            logger.info(f"API key validated for user {user_id}")
            return True, user_info

        except Exception as e:
            logger.error(f"Error validating API key: {str(e)}")
            return False, None

    async def check_rate_limit(
        self, user_id: str, rate_limit: int
    ) -> Tuple[bool, Dict[str, Any]]:
        """
        Check if user has exceeded rate limit.

        Returns:
            Tuple of (is_allowed, rate_info)
        """
        try:
            r = await self._get_redis()

            # Rate limit key
            rate_key = f"rate:{user_id}"

            # Get current count
            current = await r.get(rate_key)
            current_count = int(current) if current else 0

            # Check limit
            if current_count >= rate_limit:
                ttl = await r.ttl(rate_key)
                return False, {
                    "allowed": False,
                    "limit": rate_limit,
                    "remaining": 0,
                    "used": current_count,
                    "reset_in_seconds": ttl if ttl > 0 else 0,
                }

            # Increment counter
            pipe = r.pipeline()
            pipe.incr(rate_key)
            pipe.expire(rate_key, RATE_LIMIT_WINDOW)
            results = await pipe.execute()

            new_count = results[0]

            return True, {
                "allowed": True,
                "limit": rate_limit,
                "remaining": max(0, rate_limit - new_count),
                "used": new_count,
                "reset_in_seconds": RATE_LIMIT_WINDOW,
            }

        except Exception as e:
            logger.error(f"Error checking rate limit: {str(e)}")
            # Allow request on error but log it
            return True, {
                "allowed": True,
                "error": str(e),
                "limit": rate_limit,
                "remaining": -1,
                "used": -1,
                "reset_in_seconds": -1,
            }

    async def get_user_permissions(
        self, user_id: str, organization_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """Get detailed user permissions."""
        try:
            # Check cache first
            r = await self._get_redis()
            cache_key = f"perms:{user_id}"
            if organization_id:
                cache_key += f":{organization_id}"

            cached_perms = await r.get(cache_key)
            if cached_perms:
                return json.loads(cached_perms)

            # Build permissions
            permissions = {
                "user_id": user_id,
                "organization_id": organization_id,
                "resources": {},
                "features": [],
            }

            # Get user document
            user_doc = await self.db.collection("users").document(user_id).get()
            if user_doc.exists:
                user_data = user_doc.to_dict()
                permissions["plan"] = user_data.get("plan", "free")
                permissions["features"] = user_data.get("features", [])

                # Get plan features
                if plan := user_data.get("plan"):
                    plan_doc = await self.db.collection("plans").document(plan).get()
                    if plan_doc.exists:
                        plan_data = plan_doc.to_dict()
                        permissions["features"].extend(plan_data.get("features", []))
                        permissions["resources"] = plan_data.get("resources", {})

            # Get organization permissions if applicable
            if organization_id:
                org_member_doc = (
                    await self.db.collection("organizations")
                    .document(organization_id)
                    .collection("members")
                    .document(user_id)
                    .get()
                )

                if org_member_doc.exists:
                    member_data = org_member_doc.to_dict()
                    permissions["role"] = member_data.get("role", "member")
                    permissions["org_permissions"] = member_data.get("permissions", [])

            # Cache permissions
            await r.set(cache_key, json.dumps(permissions), ex=CACHE_TTL_SECONDS)

            return permissions

        except Exception as e:
            logger.error(f"Error getting user permissions: {str(e)}")
            return {
                "user_id": user_id,
                "error": str(e),
                "resources": {},
                "features": [],
            }

    async def validate_request(self, api_key: str) -> Dict[str, Any]:
        """
        Perform complete request validation including auth and rate limiting.
        """
        # Validate API key
        is_valid, user_info = await self.validate_api_key(api_key)

        if not is_valid:
            return {"valid": False, "error": "Invalid API key", "status_code": 401}

        # Check rate limit
        user_id = user_info["user_id"]
        rate_limit = user_info.get("rate_limit", DEFAULT_RATE_LIMIT)

        is_allowed, rate_info = await self.check_rate_limit(user_id, rate_limit)

        if not is_allowed:
            return {
                "valid": False,
                "error": "Rate limit exceeded",
                "status_code": 429,
                "rate_limit": rate_info,
            }

        # Get permissions
        permissions = await self.get_user_permissions(
            user_id, user_info.get("organization_id")
        )

        return {
            "valid": True,
            "user": user_info,
            "rate_limit": rate_info,
            "permissions": permissions,
            "status_code": 200,
        }


def auth_validator(request: Request) -> Response:
    """
    Cloud Function entry point for authentication validation.

    Args:
        request: The Flask request object

    Returns:
        Response with validation results
    """
    # Handle CORS
    if request.method == "OPTIONS":
        headers = {
            "Access-Control-Allow-Origin": "*",
            "Access-Control-Allow-Methods": "POST",
            "Access-Control-Allow-Headers": "Content-Type, Authorization",
            "Access-Control-Max-Age": "3600",
        }
        return Response("", 204, headers)

    headers = {"Content-Type": "application/json"}

    try:
        # Get API key from header or body
        api_key = request.headers.get("Authorization", "").replace("Bearer ", "")

        if not api_key and request.is_json:
            api_key = request.get_json().get("api_key", "")

        if not api_key:
            return Response(
                json.dumps(
                    {
                        "valid": False,
                        "error": "No API key provided",
                        "hint": "Include API key in Authorization header or request body",
                    }
                ),
                401,
                headers,
            )

        # Validate request
        validator = AuthValidator()
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)

        try:
            result = loop.run_until_complete(validator.validate_request(api_key))

            # Add rate limit headers
            if "rate_limit" in result:
                headers.update(
                    {
                        "X-RateLimit-Limit": str(result["rate_limit"].get("limit", -1)),
                        "X-RateLimit-Remaining": str(
                            result["rate_limit"].get("remaining", -1)
                        ),
                        "X-RateLimit-Reset": str(
                            result["rate_limit"].get("reset_in_seconds", -1)
                        ),
                    }
                )

            status_code = result.pop("status_code", 200)

            return Response(json.dumps(result, indent=2), status_code, headers)

        finally:
            # Close connections
            loop.run_until_complete(validator._close_redis())
            loop.close()

    except Exception as e:
        logger.error(f"Auth validator failed: {str(e)}")
        error_response = {
            "valid": False,
            "error": "Internal server error",
            "type": type(e).__name__,
            "timestamp": datetime.utcnow().isoformat(),
        }
        return Response(json.dumps(error_response), 500, headers)


# For local testing
if __name__ == "__main__":
    from flask import Flask, request as flask_request

    app = Flask(__name__)

    @app.route("/", methods=["POST", "OPTIONS"])
    def main():
        return auth_validator(flask_request)

    app.run(debug=True, port=8080)
