# Virtuoso API Comprehensive Endpoint Catalog

## Executive Summary

This document provides a comprehensive catalog of all available API endpoints from the Virtuoso API based on analysis of two Postman collections. The analysis identified **187 total endpoints** across 13 functional categories, with detailed priority assessments for CLI implementation.

## Key Statistics

- **Total Endpoints**: 187
- **High Priority**: 113 endpoints (60.4%)
- **Medium Priority**: 57 endpoints (30.5%)
- **Low Priority**: 17 endpoints (9.1%)

### Method Distribution
- **GET**: 86 endpoints (46.0%)
- **POST**: 77 endpoints (41.2%)
- **PUT**: 11 endpoints (5.9%)
- **DELETE**: 13 endpoints (7.0%)

### Category Distribution
- **Test Steps**: 55 endpoints (29.4%)
- **Test Suites (Journeys)**: 26 endpoints (13.9%)
- **Goals**: 17 endpoints (9.1%)
- **Projects**: 16 endpoints (8.6%)
- **Other**: 16 endpoints (8.6%)
- **Scripts**: 13 endpoints (7.0%)
- **Test Data**: 13 endpoints (7.0%)
- **API Tests**: 9 endpoints (4.8%)
- **Environments**: 8 endpoints (4.3%)
- **Executions**: 5 endpoints (2.7%)
- **Test Cases (Checkpoints)**: 4 endpoints (2.1%)
- **Integrations**: 3 endpoints (1.6%)
- **Healing**: 2 endpoints (1.1%)

## High Priority Endpoints for CLI Implementation

### 1. Projects Management (15 endpoints)
Core project operations essential for CLI functionality.

#### Key Endpoints:
- `POST /projects` - Create new project
- `GET /projects` - List projects by organization
- `GET /projects/{projectId}` - Get project details
- `GET /projects/{projectId}/goals` - List project goals
- `GET /projects/{projectId}/scripts` - List project scripts
- `GET /projects/metrics/*` - Project metrics and analytics

#### CLI Commands to Implement:
- `create-project`
- `list-projects`
- `get-project`
- `get-project-goals`
- `get-project-metrics`

### 2. Goals Management (16 endpoints)
Goal creation and management - central to test automation.

#### Key Endpoints:
- `POST /goals` - Create new goal
- `GET /goals/{goalId}` - Get goal details
- `GET /goals/{goalId}/versions` - Get goal versions/snapshots
- `GET /goals/{goalId}/checkpoints` - List goal checkpoints
- `POST /goals/{goalId}/snapshots/{snapshotId}/execute` - Execute goal

#### CLI Commands to Implement:
- `create-goal`
- `list-goals`
- `get-goal`
- `get-goal-versions`
- `execute-goal`

### 3. Test Suites (Journeys) Management (23 endpoints)
Journey management for organizing test execution.

#### Key Endpoints:
- `POST /testsuites` - Create new journey
- `GET /testsuites/{journeyId}` - Get journey details
- `GET /testsuites?goalId={goalId}` - List goal journeys
- `POST /testsuites/{journeyId}/checkpoints/attach` - Attach checkpoint to journey
- `GET /testsuites/latest_status` - Get execution status

#### CLI Commands to Implement:
- `create-journey`
- `list-journeys`
- `get-journey`
- `attach-checkpoint`
- `get-journey-status`

### 4. Test Cases (Checkpoints) Management (2 endpoints)
Checkpoint creation and management.

#### Key Endpoints:
- `POST /testcases` - Create new checkpoint
- `DELETE /testcases/{checkpointId}/steps/{stepId}` - Delete checkpoint step

#### CLI Commands to Implement:
- `create-checkpoint`
- `delete-checkpoint-step`

### 5. Test Steps Management (50 endpoints)
Individual test step creation - largest category.

#### Key Endpoints:
- `POST /teststeps` - Create test step (multiple variations)
- `GET /teststeps/{stepId}` - Get step details
- `PUT /teststeps/{stepId}/properties` - Update step properties

#### CLI Commands to Implement:
- `create-step-*` (39 step types already implemented)
- `get-step`
- `update-step`

### 6. Execution Management (5 endpoints)
Test execution monitoring and results.

#### Key Endpoints:
- `GET /executions/{executionId}` - Get execution details
- `GET /executions/analysis/{executionId}` - Get execution analysis
- `POST /executions/{executionId}/failures/explain` - Get AI failure analysis
- `GET /jobs/{executionId}/status` - Get job status

#### CLI Commands to Implement:
- `get-execution`
- `get-execution-analysis`
- `get-execution-status`
- `explain-failures`

## Medium Priority Endpoints

### 1. Test Data Management (13 endpoints)
Test data tables and values management.

#### Key Endpoints:
- `POST /testdata/tables/create` - Create test data table
- `GET /testdata/tables/{tableId}` - Get table details
- `GET /testdata/tables/{tableId}/values` - Get table values
- `POST /testdata/tables/clone` - Clone table

### 2. Scripts/Extensions Management (13 endpoints)
Custom script and extension management.

#### Key Endpoints:
- `POST /scripts` - Create script
- `GET /scripts/{scriptId}` - Get script details
- `PUT /scripts/{scriptId}` - Update script
- `GET /scripts/{scriptId}/versions` - Get script versions

### 3. Environments Management (8 endpoints)
Environment configuration and variables.

#### Key Endpoints:
- `POST /environments` - Create environment
- `GET /environments/{environmentId}` - Get environment details
- `POST /environments/{environmentId}/variables` - Create environment variable

### 4. API Tests Management (9 endpoints)
API test creation and management.

#### Key Endpoints:
- `POST /api-tests/apis` - Create API test
- `GET /api-tests/apis/{apiTestId}` - Get API test details
- `POST /api-tests/folders` - Create test folder

### 5. Integrations Management (3 endpoints)
External integrations and installations.

#### Key Endpoints:
- `GET /integrations` - List integrations
- `GET /integrations/installations` - List installations
- `GET /integrations/interactive` - Interactive integrations

## Low Priority Endpoints

### 1. DELETE Operations (13 endpoints)
Delete operations - dangerous, require careful implementation.

### 2. PUT Operations (11 endpoints)
Update operations - can often be replaced with create operations.

### 3. Healing Operations (2 endpoints)
Element healing - advanced feature.

### 4. Organization Management
User and organization management - less critical for CLI.

## CLI Implementation Recommendations

### Phase 1: Core Commands (Immediate Priority)
1. `create-project` / `list-projects` / `get-project`
2. `create-goal` / `list-goals` / `get-goal`
3. `create-journey` / `list-journeys` / `get-journey`
4. `create-checkpoint` / `list-checkpoints`
5. `execute-goal` / `get-execution-status`

### Phase 2: Enhanced Commands (Medium Priority)
1. `manage-test-data` / `import-test-data` / `export-test-data`
2. `create-environment` / `list-environments` / `manage-environment-variables`
3. `create-api-test` / `list-api-tests`
4. `get-project-metrics` / `get-goal-metrics`
5. `monitor-execution` / `get-execution-logs`

### Phase 3: Advanced Commands (Lower Priority)
1. `create-script` / `list-scripts` / `update-script`
2. `clone-project` / `export-project` / `import-project`
3. `validate-test-structure`
4. `manage-integrations`
5. `healing-operations`

## Currently Implemented vs. Missing Commands

### ✅ Already Implemented (39 commands)
- All step creation commands (`create-step-*`)
- Basic structure creation (`create-project`, `create-goal`, `create-journey`, `create-checkpoint`)
- Listing commands (`list-projects`, `list-goals`, `list-journeys`, `list-checkpoints`)
- Execution commands (`execute-goal`)
- Context management (`set-checkpoint`)

### ❌ Missing High-Value Commands (7 commands)
1. `manage-test-data` - Test data tables management
2. `import-test-data` - Import CSV/Excel data
3. `export-test-data` - Export test data
4. `monitor-execution` - Real-time execution monitoring
5. `get-execution-logs` - Detailed execution logs
6. `get-project-metrics` - Project analytics
7. `get-goal-metrics` - Goal analytics

### 🔄 Enhancement Opportunities (8 commands)
1. `create-batch-structure` - Enhanced batch operations
2. `clone-project` - Project cloning
3. `export-project` - Project export
4. `import-project` - Project import
5. `validate-test-structure` - Structure validation
6. `run-test-suite` - Enhanced test execution
7. `get-test-results` - Results analysis
8. `create-environment` - Environment management

## Authentication & Headers

All endpoints require:
- `Authorization: Bearer <token>`
- `X-Virtuoso-Client-ID: <client-id>`
- `X-Virtuoso-Client-Name: <client-name>`

Base URL: `https://api-app2.virtuoso.qa/api`

## Common Parameters

### Path Parameters
- `{projectId}` - Project identifier
- `{goalId}` - Goal identifier
- `{journeyId}` - Journey identifier
- `{checkpointId}` - Checkpoint identifier
- `{stepId}` - Step identifier
- `{executionId}` - Execution identifier

### Query Parameters
- `envelope` - Response envelope format (true/false)
- `organizationId` - Organization identifier
- `archived` - Include archived items (true/false)
- `detailed` - Include detailed information (true/false)
- `includeBody` - Include body content (true/false)

## Request/Response Patterns

### Standard Response Format
```json
{
  "success": true,
  "item": {
    "id": 123,
    "name": "...",
    "...": "..."
  }
}
```

### Error Response Format
```json
{
  "success": false,
  "error": {
    "message": "Error description",
    "code": "ERROR_CODE"
  }
}
```

## Next Steps

1. **Implement Missing High-Value Commands**: Focus on test data management, execution monitoring, and metrics
2. **Enhance Existing Commands**: Add better error handling, output formatting, and parameter validation
3. **Add Batch Operations**: Implement bulk operations for efficiency
4. **Improve Documentation**: Add comprehensive help text and examples
5. **Add Integration Tests**: Create comprehensive test suites for all commands

## Files Generated

- `virtuoso_api_endpoints.json` - Raw endpoint data
- `virtuoso_api_analysis.json` - Detailed analysis and recommendations
- `extract_endpoints.py` - Endpoint extraction script
- `analyze_endpoints.py` - Endpoint analysis script

This catalog provides a comprehensive foundation for expanding the Virtuoso CLI with additional high-value commands and features.