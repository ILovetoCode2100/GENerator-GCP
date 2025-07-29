# Virtuoso API CLI - FastAPI Service

A RESTful API wrapper for the Virtuoso CLI, providing HTTP endpoints for test automation commands.

## Features

- **Command Execution**: Execute any Virtuoso CLI command via HTTP endpoints
- **Test Management**: Create and run test suites from YAML/JSON definitions
- **Session Management**: Manage test sessions and checkpoints
- **Health Monitoring**: Check API and CLI health status
- **Request Tracking**: Unique request IDs for tracing
- **Rate Limiting**: Configurable rate limits for API protection
- **CORS Support**: Configurable CORS for browser-based clients
- **Structured Logging**: JSON-formatted logs for production

## Quick Start

### Installation

```bash
cd api
pip install -r requirements.txt
```

### Configuration

1. Copy the example environment file:

```bash
cp .env.example .env
```

2. Update `.env` with your configuration:

- Set `VIRTUOSO_API_KEY` to your Virtuoso API key
- Set `VIRTUOSO_ORG_ID` to your organization ID
- Configure `API_KEYS` for API authentication
- Adjust other settings as needed

### Running the API

#### Development Mode

```bash
# With auto-reload
uvicorn app.main:app --reload --port 8000

# Or use the main.py directly
python app/main.py
```

#### Production Mode

```bash
# With multiple workers
uvicorn app.main:app --host 0.0.0.0 --port 8000 --workers 4

# Or with gunicorn
gunicorn app.main:app -w 4 -k uvicorn.workers.UvicornWorker
```

### API Documentation

Once running, visit:

- Swagger UI: http://localhost:8000/docs
- ReDoc: http://localhost:8000/redoc
- OpenAPI Schema: http://localhost:8000/openapi.json

## API Endpoints

### Health Check

```bash
GET /health              # Basic health check
GET /health/ready        # Readiness check
GET /health/live         # Liveness check
```

### Commands

```bash
POST   /api/v1/commands/execute      # Execute a CLI command
GET    /api/v1/commands/list         # List available commands
GET    /api/v1/commands/{cmd}/help   # Get command help
```

### Tests

```bash
POST   /api/v1/tests/run             # Run a test from definition
POST   /api/v1/tests/upload          # Upload and run a test file
GET    /api/v1/tests/templates       # List test templates
GET    /api/v1/tests/templates/{id}  # Get a test template
```

### Sessions

```bash
POST   /api/v1/sessions              # Create a session
GET    /api/v1/sessions              # List sessions
GET    /api/v1/sessions/{id}         # Get session details
PATCH  /api/v1/sessions/{id}         # Update a session
DELETE /api/v1/sessions/{id}         # Delete a session
POST   /api/v1/sessions/{id}/activate # Activate a session
```

## Authentication

Include your API key in the `X-API-Key` header:

```bash
curl -H "X-API-Key: your-api-key" http://localhost:8000/api/v1/commands/list
```

## Request/Response Format

### Execute Command Example

Request:

```json
POST /api/v1/commands/execute
{
  "command": "step-navigate",
  "args": ["to", "https://example.com"],
  "checkpoint_id": "12345"
}
```

Response:

```json
{
  "success": true,
  "output": "Navigation step created successfully",
  "error": null,
  "exit_code": 0,
  "command": "step-navigate"
}
```

### Run Test Example

Request:

```json
POST /api/v1/tests/run
{
  "definition": {
    "name": "Login Test",
    "steps": [
      {"navigate": "https://example.com"},
      {"click": "#login"},
      {"write": {"selector": "#email", "text": "test@example.com"}},
      {"click": "button[type='submit']"},
      {"assert": "Welcome"}
    ]
  },
  "dry_run": false,
  "execute": true
}
```

Response:

```json
{
  "test_id": "test_123",
  "status": "created",
  "project_id": "proj_456",
  "checkpoint_id": "cp_789",
  "execution_id": "exec_012",
  "steps_created": 5
}
```

## Error Handling

The API returns consistent error responses:

```json
{
  "error": {
    "message": "Validation error",
    "type": "validation_error",
    "details": [
      {
        "field": "command",
        "message": "field required",
        "type": "value_error.missing"
      }
    ]
  },
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

## Rate Limiting

API requests are rate-limited. Check response headers:

- `X-RateLimit-Limit`: Request limit per period
- `X-RateLimit-Remaining`: Remaining requests
- `X-RateLimit-Reset`: Reset timestamp

## Development

### Running Tests

```bash
pytest
pytest --cov=app tests/
```

### Code Quality

```bash
# Format code
black app/

# Lint
flake8 app/

# Type check
mypy app/
```

### Docker Support

Build and run with Docker:

```bash
# Build
docker build -f Dockerfile.api -t virtuoso-api .

# Run
docker run -p 8000:8000 --env-file .env virtuoso-api
```

## Environment Variables

See `.env.example` for all available configuration options.

Key variables:

- `ENVIRONMENT`: Set to "production" for production deployments
- `DEBUG`: Enable debug mode (auto-reload, detailed errors)
- `API_KEYS`: Comma-separated list of valid API keys
- `VIRTUOSO_API_KEY`: Your Virtuoso API key
- `CLI_PATH`: Path to the api-cli binary

## Monitoring

The API includes:

- Structured JSON logging
- Request ID tracking
- Performance metrics in logs
- Health check endpoints for monitoring

## Security Considerations

- Always use HTTPS in production
- Rotate API keys regularly
- Set appropriate CORS origins
- Enable rate limiting
- Use environment variables for secrets
- Never commit `.env` files

## License

[License information here]
