#!/usr/bin/env node

const https = require('https');
const { URL } = require('url');

// Configuration
const API_GATEWAY_ENDPOINT = 'https://4sswk1wyv9.execute-api.us-east-1.amazonaws.com/Prod/virtuoso';
const API_TOKEN = 'f7a55516-5cc4-4529-b2ae-8e106a7d164e';

// Test data storage
const testData = {
  projectId: null,
  goalId: null,
  journeyId: null,
  checkpointId: null,
  stepId: null,
  executionId: null,
  libraryCheckpointId: null,
  dataTableId: null,
  environmentId: null
};

// Test results
const testResults = {
  passed: 0,
  failed: 0,
  tests: []
};

// Colors for console output
const colors = {
  reset: '\x1b[0m',
  bright: '\x1b[1m',
  green: '\x1b[32m',
  red: '\x1b[31m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  cyan: '\x1b[36m'
};

// Helper function to make API calls via API Gateway endpoints
async function callVirtuosoAPI(method, path, body = null) {
  return new Promise((resolve, reject) => {
    const url = new URL(`${API_GATEWAY_ENDPOINT}${path}`);
    
    const options = {
      hostname: url.hostname,
      port: 443,
      path: url.pathname + (url.search || ''),
      method: method,
      headers: {
        'Content-Type': 'application/json'
      }
    };

    if (body) {
      const requestData = JSON.stringify(body);
      options.headers['Content-Length'] = requestData.length;
    }

    const req = https.request(options, (res) => {
      let data = '';

      res.on('data', (chunk) => {
        data += chunk;
      });

      res.on('end', () => {
        try {
          const response = JSON.parse(data);
          if (res.statusCode >= 200 && res.statusCode < 300) {
            resolve(response);
          } else {
            reject(new Error(`API call failed: ${res.statusCode} - ${JSON.stringify(response)}`));
          }
        } catch (e) {
          reject(new Error(`Failed to parse response: ${data}`));
        }
      });
    });

    req.on('error', (error) => {
      reject(error);
    });

    if (body) {
      req.write(JSON.stringify(body));
    }
    req.end();
  });
}

// Direct API call to test the holy grail endpoint
async function callDirectAPI(method, path, body = null) {
  return new Promise((resolve, reject) => {
    const url = new URL(`${API_GATEWAY_ENDPOINT}${path}`);
    
    const options = {
      hostname: url.hostname,
      port: 443,
      path: url.pathname,
      method: method,
      headers: {
        'Content-Type': 'application/json'
      }
    };

    if (body) {
      const requestData = JSON.stringify(body);
      options.headers['Content-Length'] = requestData.length;
    }

    const req = https.request(options, (res) => {
      let data = '';

      res.on('data', (chunk) => {
        data += chunk;
      });

      res.on('end', () => {
        try {
          const response = JSON.parse(data);
          resolve({ status: res.statusCode, data: response });
        } catch (e) {
          resolve({ status: res.statusCode, data: data });
        }
      });
    });

    req.on('error', (error) => {
      reject(error);
    });

    if (body) {
      req.write(JSON.stringify(body));
    }
    req.end();
  });
}

// Test runner
async function runTest(testName, testFn) {
  process.stdout.write(`${colors.blue}Running test: ${testName}...${colors.reset} `);
  
  try {
    const startTime = Date.now();
    const result = await testFn();
    const duration = Date.now() - startTime;
    
    console.log(`${colors.green}✓ PASSED${colors.reset} (${duration}ms)`);
    testResults.passed++;
    testResults.tests.push({
      name: testName,
      status: 'passed',
      duration: duration,
      result: result
    });
    
    return result;
  } catch (error) {
    console.log(`${colors.red}✗ FAILED${colors.reset}`);
    console.error(`  ${colors.red}Error: ${error.message}${colors.reset}`);
    testResults.failed++;
    testResults.tests.push({
      name: testName,
      status: 'failed',
      error: error.message
    });
    
    throw error;
  }
}

// Test suite
async function runTestSuite() {
  console.log(`${colors.bright}${colors.cyan}🚀 Virtuoso API Lambda Test Suite${colors.reset}\n`);
  console.log(`API Endpoint: ${API_GATEWAY_ENDPOINT}\n`);
  
  // 1. Create Project
  await runTest('Create Project', async () => {
    const result = await callVirtuosoAPI('POST', '/projects', {
      name: `Test Project ${Date.now()}`,
      description: 'Automated test project'
    });
    
    // Extract project ID from response or use a mock ID
    testData.projectId = result.id || 'test-project-' + Date.now();
    return result;
  });

  // 2. List Projects
  await runTest('List Projects', async () => {
    const result = await callVirtuosoAPI('GET', '/projects');
    return result;
  });

  // 3. Create Goal
  await runTest('Create Goal', async () => {
    const result = await callVirtuosoAPI('POST', '/goals', {
      projectId: testData.projectId,
      name: `Test Goal ${Date.now()}`,
      description: 'Automated test goal'
    });
    
    testData.goalId = result.id || 'test-goal-' + Date.now();
    return result;
  });

  // 4. Create Journey (TestSuite)
  await runTest('Create Journey', async () => {
    const result = await callVirtuosoAPI('POST', '/testsuites', {
      goalId: testData.goalId,
      name: `Test Journey ${Date.now()}`,
      description: 'Automated test journey'
    });
    
    testData.journeyId = result.id || 'test-journey-' + Date.now();
    return result;
  });

  // 5. Create Checkpoint
  await runTest('Create Checkpoint', async () => {
    const result = await callVirtuosoAPI('POST', '/testcases', {
      journeyId: testData.journeyId,
      name: `Test Checkpoint ${Date.now()}`,
      description: 'Automated test checkpoint'
    });
    
    testData.checkpointId = result.id || 'test-checkpoint-' + Date.now();
    return result;
  });

  // 6. Add Test Step
  await runTest('Add Test Step', async () => {
    const result = await callVirtuosoAPI('POST', '/teststeps', {
      checkpointId: testData.checkpointId,
      action: 'navigate',
      target: 'https://example.com',
      description: 'Navigate to example.com'
    });
    
    testData.stepId = result.id || 'test-step-' + Date.now();
    return result;
  });

  // 7. Create Environment
  await runTest('Create Environment', async () => {
    const result = await callVirtuosoAPI('POST', '/environments', {
      name: `Test Environment ${Date.now()}`,
      baseUrl: 'https://test.example.com'
    });
    
    testData.environmentId = result.id || 'test-env-' + Date.now();
    return result;
  });

  // 8. Create Test Data Table
  await runTest('Create Test Data Table', async () => {
    const result = await callVirtuosoAPI('POST', '/testdata/tables/create', {
      name: `Test Data Table ${Date.now()}`,
      columns: ['username', 'password', 'email']
    });
    
    testData.dataTableId = result.id || 'test-table-' + Date.now();
    return result;
  });

  // 9. Execute Goal
  await runTest('Execute Goal', async () => {
    const result = await callVirtuosoAPI('POST', '/executions', {
      goalId: testData.goalId,
      environment: testData.environmentId
    });
    
    testData.executionId = result.id || 'test-execution-' + Date.now();
    return result;
  });

  // 10. Get Execution Status
  await runTest('Get Execution Status', async () => {
    const result = await callVirtuosoAPI('GET', `/executions/${testData.executionId}`);
    return result;
  });

  console.log(`\n${colors.bright}${colors.yellow}🎯 HOLY GRAIL TEST${colors.reset}\n`);

  // 11. HOLY GRAIL TEST - Get Journey Details (including steps and checkpoints)
  await runTest('GET /testsuites/{journeyId} - Holy Grail Test', async () => {
    // Call the API Gateway endpoint directly
    const result = await callVirtuosoAPI('GET', `/testsuites/${testData.journeyId}`);
    
    console.log(`\n${colors.cyan}API Gateway Response:${colors.reset}`);
    console.log(JSON.stringify(result, null, 2));
    
    return result;
  });

  // 12. Add Checkpoint to Library
  await runTest('Add Checkpoint to Library', async () => {
    const result = await callVirtuosoAPI('POST', '/library/checkpoints', {
      checkpointId: testData.checkpointId,
      name: `Library Checkpoint ${Date.now()}`
    });
    
    testData.libraryCheckpointId = result.id || 'test-lib-checkpoint-' + Date.now();
    return result;
  });

  // Print summary
  console.log(`\n${colors.bright}${colors.cyan}📊 Test Summary${colors.reset}`);
  console.log(`${colors.bright}=================${colors.reset}\n`);
  console.log(`Total tests: ${testResults.passed + testResults.failed}`);
  console.log(`${colors.green}Passed: ${testResults.passed}${colors.reset}`);
  console.log(`${colors.red}Failed: ${testResults.failed}${colors.reset}`);
  
  console.log(`\n${colors.bright}Test Data Created:${colors.reset}`);
  Object.entries(testData).forEach(([key, value]) => {
    if (value) {
      console.log(`  ${key}: ${value}`);
    }
  });

  // Save test results
  const fs = require('fs');
  const reportPath = './test-results.json';
  fs.writeFileSync(reportPath, JSON.stringify({
    timestamp: new Date().toISOString(),
    endpoint: API_GATEWAY_ENDPOINT,
    summary: {
      total: testResults.passed + testResults.failed,
      passed: testResults.passed,
      failed: testResults.failed
    },
    testData: testData,
    tests: testResults.tests
  }, null, 2));
  
  console.log(`\n${colors.cyan}Test results saved to: ${reportPath}${colors.reset}`);
  
  return testResults.failed === 0;
}

// Main execution
async function main() {
  try {
    const success = await runTestSuite();
    process.exit(success ? 0 : 1);
  } catch (error) {
    console.error(`\n${colors.red}Fatal error: ${error.message}${colors.reset}`);
    process.exit(1);
  }
}

// Run the tests
if (require.main === module) {
  main();
}