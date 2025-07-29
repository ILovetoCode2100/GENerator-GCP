# 🚀 Prompt: Deploy Full Virtuoso API to GCP

## Context

I have successfully deployed a basic version of the Virtuoso API to Google Cloud Platform. The infrastructure is working with:

- Project: virtuoso-api-1753389008
- URL: https://virtuoso-api-936111683985.us-central1.run.app
- Secrets configured: VIRTUOSO_API_TOKEN and API_KEYS
- Basic endpoints working (/health, /, /api/v1/commands)

## Current State

- Using simplified `simple_api.py` instead of full FastAPI application
- No CLI integration yet
- No database or async processing
- Located in: /Users/marklovelady/\_dev/\_projects/virtuoso-GENerator

## Objective

Deploy the complete Virtuoso API with all 70 CLI commands and full functionality to the existing GCP project.

## Tasks

### 1. Enable Required GCP Services

Enable the following APIs for the project virtuoso-api-1753389008:

- Firestore API (for session/data storage)
- Cloud Tasks API (for async command execution)
- Cloud Pub/Sub API (for event-driven architecture)
- Cloud Storage API (for logs and artifacts)
- BigQuery API (for analytics)

### 2. Update the Deployment

- Switch from `simple_api.py` to the full FastAPI application in `/api`
- Ensure the CLI binary (`/bin/api-cli`) is included and executable
- Configure the Virtuoso config file template
- Set up proper Python path and module imports

### 3. Configure GCP Services

- Create Firestore database in Native mode
- Set up Cloud Tasks queues for async operations
- Create Pub/Sub topics for events
- Configure service account permissions

### 4. Deploy Full Application

- Build and deploy the complete API with all routes:
  - All 70 CLI command endpoints
  - Session management
  - Test execution (run-test)
  - Analytics endpoints
  - Webhook handling
- Ensure proper error handling and logging
- Configure health checks for all services

### 5. Test the Deployment

- Verify all command groups work (step-assert, step-interact, etc.)
- Test CLI command execution
- Verify Firestore connectivity
- Test async command execution via Cloud Tasks
- Ensure authentication is working with API keys

### 6. Set Up Monitoring

- Configure Cloud Monitoring dashboards
- Set up alerts for errors and performance
- Enable distributed tracing
- Configure log aggregation

## Expected Outcome

A fully functional Virtuoso API on GCP that can:

- Execute all 70 CLI commands via REST endpoints
- Store sessions and data in Firestore
- Process async commands via Cloud Tasks
- Handle webhooks and events via Pub/Sub
- Scale automatically based on load
- Provide comprehensive monitoring and logging

## Important Details

- Keep using the existing project: virtuoso-api-1753389008
- Maintain the current URL: https://virtuoso-api-936111683985.us-central1.run.app
- Preserve the existing secrets (VIRTUOSO_API_TOKEN and API_KEYS)
- Ensure backward compatibility with the test endpoints
- Use the GCP free tier where possible to minimize costs

## Constraints

- Minimize downtime during deployment
- Ensure all changes are reversible
- Document any manual steps required
- Keep costs within free tier limits where possible

Please help me complete this full deployment step by step, starting with enabling the required GCP services.
