# 🔄 Rollback Plan

## Overview

This document outlines the procedures to rollback the Virtuoso API deployment in case of critical issues. The rollback process is designed to be quick, safe, and minimize downtime.

## Rollback Triggers

Rollback should be initiated when:

- **Critical Error Rate**: > 20% of requests failing
- **Performance Degradation**: P95 latency > 5 seconds
- **Service Unavailable**: Complete service failure
- **Data Corruption**: Incorrect data being returned
- **Security Breach**: Compromised API keys or unauthorized access

## Pre-Rollback Checklist

Before initiating rollback:

1. [ ] Confirm the issue is deployment-related
2. [ ] Document the specific problem
3. [ ] Notify stakeholders via Slack/Email
4. [ ] Take screenshot of error metrics
5. [ ] Export current logs for analysis

## Rollback Methods

### Method 1: Quick Rollback (Recommended)

**Time to Complete**: 2-5 minutes

```bash
# 1. List current Cloud Run revisions
gcloud run revisions list --service virtuoso-api --region us-central1

# 2. Get the previous stable revision (e.g., virtuoso-api-00001-abc)
PREVIOUS_REVISION="virtuoso-api-00001-abc"

# 3. Route 100% traffic to previous revision
gcloud run services update-traffic virtuoso-api \
  --region us-central1 \
  --to-revisions $PREVIOUS_REVISION=100

# 4. Verify rollback
curl https://virtuoso-api-936111683985.us-central1.run.app/health
```

### Method 2: Container Image Rollback

**Time to Complete**: 5-10 minutes

```bash
# 1. List available container images
gcloud container images list-tags gcr.io/virtuoso-generator/virtuoso-api

# 2. Identify previous stable version
PREVIOUS_IMAGE="gcr.io/virtuoso-generator/virtuoso-api:v1.0.0"

# 3. Deploy previous image
gcloud run deploy virtuoso-api \
  --image $PREVIOUS_IMAGE \
  --region us-central1 \
  --platform managed

# 4. Verify deployment
gcloud run services describe virtuoso-api --region us-central1
```

### Method 3: Git Revert and Redeploy

**Time to Complete**: 15-20 minutes

```bash
# 1. Revert to previous commit
git log --oneline -10  # Find the stable commit
git revert HEAD
git push origin main

# 2. Trigger Cloud Build (automatic on push)
# Monitor build progress
gcloud builds list --limit 5

# 3. Verify new deployment
curl https://virtuoso-api-936111683985.us-central1.run.app/health
```

## Service-Specific Rollbacks

### Firestore Rollback

```bash
# Restore from backup
gcloud firestore import gs://virtuoso-backups/2024-07-24/

# Verify data integrity
gcloud firestore operations list
```

### BigQuery Rollback

```sql
-- Restore tables from snapshots
CREATE OR REPLACE TABLE analytics.command_executions
AS SELECT * FROM analytics.command_executions@1690234800000;

-- Verify row counts
SELECT COUNT(*) FROM analytics.command_executions;
```

### Secret Manager Rollback

```bash
# List secret versions
gcloud secrets versions list virtuoso-api-key

# Rollback to previous version
gcloud secrets versions enable 1 --secret="virtuoso-api-key"
gcloud secrets versions disable 2 --secret="virtuoso-api-key"
```

## Rollback Verification

After rollback, verify:

### 1. Health Checks

```bash
# Basic health
curl https://virtuoso-api-936111683985.us-central1.run.app/health

# Detailed health
curl https://virtuoso-api-936111683985.us-central1.run.app/health?detailed=true

# Readiness
curl https://virtuoso-api-936111683985.us-central1.run.app/health/ready
```

### 2. Core Functionality

```bash
# Test command listing
curl -H "X-API-Key: $API_KEY" \
  https://virtuoso-api-936111683985.us-central1.run.app/api/v1/commands

# Test command execution
curl -X POST -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"args": ["test"]}' \
  https://virtuoso-api-936111683985.us-central1.run.app/api/v1/commands/step/assert/exists
```

### 3. Monitoring Metrics

```bash
# Check error rates
gcloud monitoring time-series list \
  --filter='metric.type="run.googleapis.com/request_count" AND metric.label.response_code_class="5xx"'

# Check latency
gcloud monitoring time-series list \
  --filter='metric.type="run.googleapis.com/request_latencies"'
```

## Post-Rollback Actions

### Immediate (Within 1 hour)

1. **Confirm Stability**

   - Monitor error rates for 30 minutes
   - Check user reports/feedback
   - Verify all endpoints responding

2. **Communication**

   - Send rollback confirmation to stakeholders
   - Update status page
   - Post in #virtuoso-api Slack channel

3. **Initial Analysis**
   - Export logs from failed deployment
   - Create incident ticket
   - Assign investigation owner

### Short-term (Within 24 hours)

1. **Root Cause Analysis**

   - Review deployment logs
   - Analyze error patterns
   - Check for configuration changes

2. **Fix Development**

   - Create hotfix branch
   - Implement necessary fixes
   - Test thoroughly in staging

3. **Documentation**
   - Update rollback procedures if needed
   - Document lessons learned
   - Update runbooks

### Long-term (Within 1 week)

1. **Post-Mortem**

   - Conduct blameless post-mortem
   - Identify process improvements
   - Update deployment procedures

2. **Testing Enhancement**

   - Add tests for failure scenario
   - Update integration tests
   - Improve monitoring coverage

3. **Re-deployment Planning**
   - Schedule re-deployment window
   - Prepare rollout strategy
   - Notify stakeholders

## Emergency Contacts

### Primary On-Call

- **Name**: Platform Engineer
- **Phone**: +1-XXX-XXX-XXXX
- **Email**: oncall@virtuoso.dev

### Escalation Path

1. Primary On-Call Engineer
2. Platform Team Lead
3. Engineering Manager
4. CTO

### External Support

- **GCP Support**: 1-877-355-5787
- **Case Priority**: P1 for production issues

## Rollback Tags

To facilitate easy rollback, we maintain tagged revisions:

```bash
# Tag current stable revision
gcloud run services describe virtuoso-api --region us-central1 --format "value(status.latestReadyRevisionName)" > stable-revision.txt
STABLE_REVISION=$(cat stable-revision.txt)

# Create named tag
gcloud run services update-traffic virtuoso-api \
  --region us-central1 \
  --tag stable-2024-07-24 \
  --to-revisions $STABLE_REVISION=100
```

Access tagged versions:

- Stable: https://stable-2024-07-24---virtuoso-api-936111683985.us-central1.run.app
- Previous: https://stable-2024-07-23---virtuoso-api-936111683985.us-central1.run.app

## Automated Rollback

For critical metrics, automated rollback can be triggered:

```yaml
# monitoring-policy.yaml
displayName: "Auto Rollback on High Error Rate"
conditions:
  - displayName: "Error rate > 50%"
    conditionThreshold:
      filter: 'resource.type="cloud_run_revision" AND metric.type="run.googleapis.com/request_count" AND metric.label.response_code_class="5xx"'
      comparison: COMPARISON_GT
      thresholdValue: 0.5
      duration: 300s
notificationChannels:
  - projects/virtuoso-generator/notificationChannels/auto-rollback
```

## Testing Rollback Procedures

Rollback procedures should be tested monthly:

1. **Staging Environment**

   - Deploy bad version intentionally
   - Execute rollback procedure
   - Verify recovery time

2. **Production Drill**

   - During low-traffic window
   - Route 10% traffic to previous version
   - Verify metrics and logs

3. **Documentation Review**
   - Update procedures based on test results
   - Train new team members
   - Update emergency contacts

## Rollback Decision Matrix

| Severity | Error Rate | User Impact | Action                  | Timeline |
| -------- | ---------- | ----------- | ----------------------- | -------- |
| Critical | >50%       | All users   | Immediate rollback      | <5 min   |
| High     | 20-50%     | Many users  | Quick rollback          | <15 min  |
| Medium   | 5-20%      | Some users  | Evaluate, then rollback | <30 min  |
| Low      | <5%        | Few users   | Monitor, hotfix         | <2 hours |

## Appendix: Common Issues and Solutions

### Issue: API Key Not Working After Rollback

```bash
# Verify secret version
gcloud secrets versions list virtuoso-api-key
# Enable correct version
gcloud secrets versions enable 1 --secret="virtuoso-api-key"
```

### Issue: Database Connection Errors

```bash
# Check Firestore status
gcloud firestore operations list
# Verify service account permissions
gcloud projects get-iam-policy virtuoso-generator
```

### Issue: High Memory Usage

```bash
# Scale up temporarily
gcloud run services update virtuoso-api --memory 1Gi --region us-central1
```

### Issue: SSL Certificate Errors

```bash
# Check certificate status
gcloud compute ssl-certificates list
# Force certificate refresh
gcloud run services update virtuoso-api --region us-central1
```

---

**Last Updated**: July 24, 2024
**Version**: 1.0
**Owner**: Platform Team
