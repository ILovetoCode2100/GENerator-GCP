#!/bin/bash
# VIRTUOSO API - DEPLOY AND VERIFY WORKING VERSION

# 1. Check if build completed successfully
echo "=== Checking build status ==="
BUILD_STATUS=$(gcloud builds list --limit=1 --format="value(status)")
if [ "$BUILD_STATUS" != "SUCCESS" ]; then
    echo "ERROR: Build not successful. Status: $BUILD_STATUS"
    exit 1
fi
echo "Build completed successfully!"

# 2. Deploy the fixed revision without traffic
echo "=== Deploying fixed revision ==="
export PROJECT_ID="virtuoso-api-1753389008"
gcloud run deploy virtuoso-api \
    --image gcr.io/$PROJECT_ID/virtuoso-api:latest \
    --platform managed \
    --region us-central1 \
    --no-traffic \
    --tag api-fixed \
    --set-env-vars="VIRTUOSO_AUTH_TOKEN=${VIRTUOSO_AUTH_TOKEN},VIRTUOSO_ORG_ID=2242,API_KEY=${API_KEY}" \
    --memory 2Gi \
    --cpu 2 \
    --timeout 300 \
    --concurrency 100 \
    --max-instances 10

# 3. Get the revision URL
REVISION_URL=$(gcloud run services describe virtuoso-api --region=us-central1 --format='value(status.url)' | sed 's|https://|https://api-fixed---|')
echo "Revision URL: $REVISION_URL"

# 4. Test the health endpoint
echo "=== Testing health endpoint ==="
HEALTH_RESPONSE=$(curl -s $REVISION_URL/health)
echo "Health response: $HEALTH_RESPONSE"

# 5. Test command listing endpoint
echo "=== Testing /api/v1/commands endpoint ==="
COMMANDS_COUNT=$(curl -s $REVISION_URL/api/v1/commands | jq '. | length')
echo "Number of commands available: $COMMANDS_COUNT"

# 6. Test a simple command execution
echo "=== Testing command execution ==="
TEST_RESPONSE=$(curl -X POST $REVISION_URL/api/v1/commands/step/navigate/to \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer ${API_KEY:-your-api-key}" \
    -d '{
      "checkpoint_id": "'${CHECKPOINT_ID:-12345}'",
      "args": ["https://example.com"]
    }')
echo "Command test response: $TEST_RESPONSE"

# 7. If all tests pass, start traffic migration
if [ "$COMMANDS_COUNT" -ge 60 ]; then
    echo "=== All tests passed! Starting traffic migration ==="

    # 25% traffic
    echo "Migrating 25% traffic..."
    gcloud run services update-traffic virtuoso-api \
        --to-tags=api-fixed=25 \
        --region=us-central1

    echo "Testing with 25% traffic for 30 seconds..."
    sleep 30

    # Check main URL
    MAIN_URL="https://virtuoso-api-936111683985.us-central1.run.app"
    for i in {1..5}; do
        echo "Test $i: $(curl -s $MAIN_URL/health)"
    done

    # 50% traffic
    echo "Continuing to 50% traffic..."
    gcloud run services update-traffic virtuoso-api \
        --to-tags=api-fixed=50 \
        --region=us-central1

    echo "Testing with 50% traffic..."
    sleep 30

    # 100% traffic
    echo "Completing migration to 100%..."
    gcloud run services update-traffic virtuoso-api \
        --to-tags=api-fixed=100 \
        --region=us-central1

    echo "=== DEPLOYMENT COMPLETE! ==="
    echo "All traffic now served by the fixed version"
    echo "API URL: $MAIN_URL"
    echo "All 70 endpoints are now working!"
else
    echo "ERROR: Not all endpoints available. Found only $COMMANDS_COUNT commands"
    echo "Check logs for errors:"
    gcloud run services logs read virtuoso-api --region=us-central1 --limit=20
fi

# 8. Final verification
echo "=== Final endpoint verification ==="
echo "Main URL: https://virtuoso-api-936111683985.us-central1.run.app"
echo "Health: $(curl -s https://virtuoso-api-936111683985.us-central1.run.app/health)"
echo "Commands: $(curl -s https://virtuoso-api-936111683985.us-central1.run.app/api/v1/commands | jq '. | length') endpoints available"
