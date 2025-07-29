---
id: overview
title: API Overview
sidebar_position: 1
---

# API Reference Overview

The Virtuoso Test Converter API is a RESTful service that provides programmatic access to test conversion, execution, and management capabilities. This reference guide covers all available endpoints, request/response formats, and integration patterns.

## Base URL

```
https://api.virtuoso.qa/v1
```

## API Versioning

The API uses URL versioning. The current version is `v1`. When breaking changes are introduced, a new version will be released while maintaining backward compatibility.

## Content Types

The API accepts and returns JSON by default:

```
Content-Type: application/json
Accept: application/json
```

## HTTP Methods

The API follows RESTful conventions:

| Method   | Usage                               |
| -------- | ----------------------------------- |
| `GET`    | Retrieve resources                  |
| `POST`   | Create resources or trigger actions |
| `PUT`    | Update entire resources             |
| `PATCH`  | Partial resource updates            |
| `DELETE` | Remove resources                    |

## Core Endpoints

### Conversion Endpoints

Convert tests from various formats to Virtuoso format:

- `POST /convert` - Convert a single test
- `POST /convert/batch` - Convert multiple tests
- `GET /convert/formats` - List supported formats
- `GET /convert/patterns` - Get conversion patterns

### Status Endpoints

Monitor conversion and execution jobs:

- `GET /status/{job_id}` - Get job status
- `GET /status` - List all jobs
- `DELETE /status/{job_id}` - Cancel a job

### Pattern Endpoints

Access and manage conversion patterns:

- `GET /patterns` - List all patterns
- `GET /patterns/{pattern_id}` - Get pattern details
- `POST /patterns/match` - Test pattern matching
- `GET /patterns/confidence` - Get confidence scores

### Feedback Endpoints

Submit feedback to improve conversions:

- `POST /feedback` - Submit conversion feedback
- `POST /feedback/pattern` - Suggest new patterns
- `GET /feedback/{feedback_id}` - Get feedback status

## Request Structure

### Standard Request Headers

```http
X-API-Key: YOUR_API_KEY
Content-Type: application/json
X-Request-ID: unique-request-id
X-Client-Version: 1.0.0
```

### Request Body Example

```json
{
  "source_format": "selenium-python",
  "test_content": "...",
  "options": {
    "preserve_comments": true,
    "generate_descriptions": true,
    "confidence_threshold": 0.8
  }
}
```

## Response Structure

### Success Response

```json
{
  "success": true,
  "data": {
    "job_id": "job_123456",
    "status": "completed",
    "result": {
      // Response data
    }
  },
  "meta": {
    "request_id": "req_abc123",
    "timestamp": "2024-01-15T10:30:00Z",
    "duration_ms": 245
  }
}
```

### Error Response

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid source format",
    "details": {
      "field": "source_format",
      "accepted_values": [
        "selenium-java",
        "selenium-python",
        "cypress",
        "playwright"
      ]
    }
  },
  "meta": {
    "request_id": "req_abc123",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

## Status Codes

| Code  | Meaning                     |
| ----- | --------------------------- |
| `200` | Success                     |
| `201` | Created                     |
| `202` | Accepted (async processing) |
| `400` | Bad Request                 |
| `401` | Unauthorized                |
| `403` | Forbidden                   |
| `404` | Not Found                   |
| `429` | Rate Limited                |
| `500` | Internal Server Error       |
| `503` | Service Unavailable         |

## Pagination

List endpoints support pagination:

```http
GET /patterns?page=2&limit=50
```

Response includes pagination metadata:

```json
{
  "data": [...],
  "pagination": {
    "page": 2,
    "limit": 50,
    "total": 245,
    "pages": 5,
    "has_next": true,
    "has_prev": true
  }
}
```

## Filtering and Sorting

Many endpoints support filtering and sorting:

```http
GET /status?status=completed&sort=-created_at&format=selenium
```

## Rate Limiting

API requests are rate-limited to ensure fair usage:

- **Default**: 1000 requests per hour
- **Pro**: 10,000 requests per hour
- **Enterprise**: Custom limits

Rate limit headers:

```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 950
X-RateLimit-Reset: 1642251600
```

## Webhooks

Configure webhooks for async notifications:

```json
{
  "webhook_url": "https://your-domain.com/webhook",
  "events": ["conversion.completed", "test.finished"],
  "secret": "your-webhook-secret"
}
```

## SDK Support

Official SDKs are available for:

- [JavaScript/Node.js](./sdks/javascript)
- [Python](./sdks/python)
- [Java](./sdks/java)
- [C#/.NET](./sdks/csharp)
- [Go](./sdks/go)

## OpenAPI Specification

Download the complete OpenAPI 3.0 specification:

- [JSON Format](https://api.virtuoso.qa/v1/openapi.json)
- [YAML Format](https://api.virtuoso.qa/v1/openapi.yaml)

## Quick Examples

### Convert a Selenium Test

```bash
curl -X POST https://api.virtuoso.qa/v1/convert \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "source_format": "selenium-java",
    "test_content": "driver.get(\"https://example.com\");"
  }'
```

### Check Conversion Status

```bash
curl -X GET https://api.virtuoso.qa/v1/status/job_123456 \
  -H "X-API-Key: YOUR_API_KEY"
```

### List Supported Patterns

```bash
curl -X GET https://api.virtuoso.qa/v1/patterns \
  -H "X-API-Key: YOUR_API_KEY"
```

## Next Steps

- **[Authentication](./authentication)** - Set up API authentication
- **[Convert Endpoint](./endpoints/convert/overview)** - Convert your first test
- **[Error Handling](./error-handling)** - Handle API errors gracefully
- **[SDKs](./sdks/overview)** - Use our official SDKs
