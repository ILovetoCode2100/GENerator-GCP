# Manual Steps for D365 Virtuoso Test Deployment

This document outlines the manual steps that cannot be automated and must be performed by a human operator.

## Pre-Deployment Manual Steps

### 1. Obtain Virtuoso API Token

1. Log in to your Virtuoso platform account
2. Navigate to Settings or Profile section
3. Find API Keys or API Access section
4. Generate a new API token or copy existing one
5. Store it securely - you'll need it for deployment

### 2. Identify Your D365 Instance

1. Log in to your D365 environment
2. Check the URL - it should be in format: `https://[instance-name].crm.dynamics.com`
3. Extract the instance name (the part before .crm.dynamics.com)
4. Examples:
   - URL: `https://contoso-dev.crm.dynamics.com` → Instance: `contoso-dev`
   - URL: `https://acme-test.crm.dynamics.com` → Instance: `acme-test`

### 3. D365 Test Environment Preparation

#### Create Test Users

1. In D365, go to Settings → Security → Users
2. Create dedicated test users with appropriate roles:
   - Sales Test User (Sales Manager role)
   - Service Test User (Customer Service Rep role)
   - Marketing Test User (Marketing Manager role)
   - Finance Test User (Finance Manager role)
   - HR Test User (HR Manager role)
3. Document usernames and passwords for test configuration

#### Configure Test Data

1. Create test customers/accounts that won't interfere with production
2. Set up test products in the product catalog
3. Configure test price lists
4. Create test territories and business units if needed

#### Security Configuration

1. Ensure test users have appropriate permissions for their modules
2. Configure any custom security roles needed for tests
3. Set up any required team assignments

## Post-Deployment Manual Steps

### 1. Virtuoso Platform Configuration

#### Test Execution Settings

1. Log in to Virtuoso platform
2. Navigate to your newly created project
3. For each goal/module:
   - Set execution frequency (daily, weekly, on-demand)
   - Configure execution environment
   - Set timeout values based on your D365 performance
   - Configure retry policies

#### Notifications Setup

1. Go to Project Settings → Notifications
2. Configure email notifications for:
   - Test failures
   - Completion reports
   - Error thresholds
3. Add team members who should receive notifications
4. Set up Slack/Teams integration if desired

#### Team Access

1. Navigate to Project Settings → Team
2. Add team members with appropriate roles:
   - Viewers (read-only access)
   - Contributors (can modify tests)
   - Administrators (full access)
3. Configure SSO if your organization uses it

### 2. Initial Test Verification

#### Run Sample Tests

1. Select 2-3 tests from each module
2. Run them individually to verify:
   - D365 connection works
   - Login credentials are correct
   - Basic navigation functions
   - Data creation/modification works

#### Common Issues to Check

1. **Slow Page Loads**: Increase wait times in tests
2. **Element Not Found**: D365 UI might have customizations
3. **Permission Errors**: Verify test user roles
4. **Data Conflicts**: Ensure test data is isolated

### 3. Test Customization

Based on your D365 customizations, you may need to modify tests for:

#### Custom Fields

1. Identify custom fields in your D365 instance
2. Update relevant tests to include these fields
3. Add validation for custom business rules

#### Custom Workflows

1. Document any custom workflows that affect standard processes
2. Modify tests to account for additional approval steps
3. Add tests for custom workflow scenarios

#### Custom Entities

1. If you have custom entities, create new tests
2. Follow the same pattern as existing tests
3. Ensure proper relationships are tested

### 4. Performance Optimization

#### Execution Timing

1. Monitor initial test runs for duration
2. Identify long-running tests
3. Optimize by:
   - Reducing unnecessary waits
   - Combining related test steps
   - Using more efficient selectors

#### Parallel Execution

1. In Virtuoso platform, configure parallel execution:
   - Start with 2-3 parallel threads
   - Monitor for conflicts
   - Increase gradually based on D365 performance

### 5. Maintenance Planning

#### Regular Reviews

1. Schedule monthly reviews of test results
2. Identify frequently failing tests
3. Update tests for D365 updates/changes

#### Version Control

1. Keep track of D365 version/updates
2. Document which test version works with which D365 version
3. Plan test updates alongside D365 updates

## Verification Checklist

Before considering deployment complete, verify:

- [ ] All 169 tests are visible in Virtuoso platform
- [ ] Sample tests from each module run successfully
- [ ] Test users can log in and have correct permissions
- [ ] Notifications are received for test failures
- [ ] Team members have appropriate access
- [ ] Execution schedule is configured
- [ ] Performance is acceptable (tests complete in reasonable time)
- [ ] No conflicts with production data
- [ ] Custom fields/entities are handled correctly
- [ ] Documentation is complete and accessible to team

## Ongoing Manual Tasks

### Daily

- Review test execution dashboard
- Investigate any failed tests
- Check for D365 environment issues

### Weekly

- Review test coverage metrics
- Update tests for any D365 changes
- Optimize slow-running tests

### Monthly

- Full test suite review
- Update test data as needed
- Review and update test documentation
- Plan for upcoming D365 updates

## Support Contacts

Document your support contacts:

- **D365 Administrator**: [Name, Email, Phone]
- **Virtuoso Support**: support@virtuoso.qa
- **Test Team Lead**: [Name, Email]
- **Development Team**: [Contact for custom D365 features]

## Notes Section

Use this section to document:

- Specific customizations in your D365 instance
- Known issues and workarounds
- Test data management procedures
- Any organization-specific requirements
