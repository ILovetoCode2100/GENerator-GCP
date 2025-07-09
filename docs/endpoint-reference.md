# Virtuoso API Endpoints - Complete List

## Base URL
`https://api-app2.virtuoso.qa/api`

## Currently Implemented Endpoints

### 1. Projects
- **Create Project**: `POST {{baseURL}}/projects`
  - Full URL: `https://api-app2.virtuoso.qa/api/projects`
  - Working: ✅

### 2. Goals
- **Create Goal**: `POST {{baseURL}}/goals`
  - Full URL: `https://api-app2.virtuoso.qa/api/goals`
  - Working: ✅

- **Get Goal Snapshot**: `GET {{baseURL}}/goals/{{goalId}}/versions`
  - Full URL: `https://api-app2.virtuoso.qa/api/goals/{goalId}/versions`
  - Working: ✅

### 3. Journeys (Testsuites)
- **Create Journey**: `POST {{baseURL}}/testsuites?envelope=false`
  - Full URL: `https://api-app2.virtuoso.qa/api/testsuites?envelope=false`
  - Working: ✅

- **Get Journeys**: `GET {{baseURL}}/testsuites/latest_status?versionId={{snapshotId}}&goalId={{goalId}}`
  - Full URL: `https://api-app2.virtuoso.qa/api/testsuites/latest_status?versionId={snapshotId}&goalId={goalId}`
  - Not implemented in CLI yet

### 4. Checkpoints (Testcases)
- **Create Checkpoint**: `POST {{baseURL}}/testcases`
  - Full URL: `https://api-app2.virtuoso.qa/api/testcases`
  - Working: ✅

- **Attach Checkpoint**: `POST {{baseURL}}/testsuites/{{journeyId}}/checkpoints/attach`
  - Full URL: `https://api-app2.virtuoso.qa/api/testsuites/{journeyId}/checkpoints/attach`
  - Working: ✅

### 5. Steps (Teststeps)
- **Add Step**: `POST /teststeps` (NO /api prefix!)
  - Full URL: `https://api-app2.virtuoso.qa/teststeps`
  - Currently broken: ❌ (using wrong path with /api prefix)

## From Postman Collection (List Operations)

### 6. List Operations (Not yet implemented)
- **List Projects**: `GET {{baseURL}}/projects?organizationId={{organizationId}}`
  - Full URL: `https://api-app2.virtuoso.qa/api/projects?organizationId=2242`

- **List Goals**: `GET {{baseURL}}/projects/{{projectId}}/goals?archived=false`
  - Full URL: `https://api-app2.virtuoso.qa/api/projects/{projectId}/goals?archived=false`

- **Goal Details**: `GET {{baseURL}}/goals/{{goalId}}`
  - Full URL: `https://api-app2.virtuoso.qa/api/goals/{goalId}`

- **List Journeys with Status**: `GET {{baseURL}}/testsuites/latest_status?snapshotId={{snapshotId}}&goalId={{goalId}}&includeSequencesDetails=true`
  - Full URL: `https://api-app2.virtuoso.qa/api/testsuites/latest_status?snapshotId={snapshotId}&goalId={goalId}&includeSequencesDetails=true`

## Summary

### Pattern for most endpoints:
`https://api-app2.virtuoso.qa/api/{endpoint}`

### Exception:
- **Steps endpoint**: `https://api-app2.virtuoso.qa/teststeps` (no /api prefix)

### Headers (all endpoints):
```
X-Virtuoso-Client-ID: api-cli-generator
X-Virtuoso-Client-Name: api-cli-generator
Authorization: Bearer f7a55516-5cc4-4529-b2ae-8e106a7d164e
Content-Type: application/json
```
