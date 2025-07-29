# Using the GCP API

The GCP API is deployed but requires authentication. Since modifying the authentication in the deployed code requires redeployment, here are your options:

## Option 1: Use a Test API Key

The GCP API expects an API key in the `X-API-Key` header. For testing, use:

```bash
curl -X POST "https://virtuoso-api-5e22h3hywa-uc.a.run.app/api/v1/tests/run" \
  -H "X-API-Key: test-api-key-123" \
  -H "Content-Type: application/json" \
  -d '{
    "definition": {
      "name": "Your Test",
      "steps": [...]
    }
  }'
```

## Option 2: Use Local CLI (Recommended)

Since the local CLI is already working perfectly:

```bash
./bin/api-cli run-test rocketshop-test.yaml --execute
```

## Option 3: Redeploy with Auth Disabled

To properly disable auth, you'd need to:

1. Modify `api/app/middleware/auth.py` to skip authentication
2. Redeploy the API: `gcloud run deploy virtuoso-api --source=api --region=us-central1`

## Summary

The GCP API is successfully deployed and running at:

- https://virtuoso-api-5e22h3hywa-uc.a.run.app

However, it requires proper authentication setup. For now, use the local CLI which works perfectly with your Virtuoso API key.
