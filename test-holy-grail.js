#!/usr/bin/env node

const https = require('https');
const { URL } = require('url');

// Configuration
const API_GATEWAY_ENDPOINT = 'https://4sswk1wyv9.execute-api.us-east-1.amazonaws.com/Prod/virtuoso';

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

// Direct API call
async function callAPI(method, path, body = null) {
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

    console.log(`${colors.blue}Calling: ${method} ${url.href}${colors.reset}`);

    const req = https.request(options, (res) => {
      let data = '';

      res.on('data', (chunk) => {
        data += chunk;
      });

      res.on('end', () => {
        console.log(`${colors.cyan}Status: ${res.statusCode}${colors.reset}`);
        console.log(`${colors.cyan}Headers:${colors.reset}`, res.headers);
        
        try {
          const response = data ? JSON.parse(data) : {};
          resolve({ status: res.statusCode, data: response, headers: res.headers });
        } catch (e) {
          resolve({ status: res.statusCode, data: data, headers: res.headers });
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

// Test the holy grail endpoint
async function testHolyGrail() {
  console.log(`${colors.bright}${colors.yellow}🎯 HOLY GRAIL TEST - Direct API Calls${colors.reset}\n`);
  console.log(`API Endpoint: ${API_GATEWAY_ENDPOINT}\n`);

  // Test 1: List projects
  console.log(`${colors.bright}Test 1: List Projects${colors.reset}`);
  try {
    const result = await callAPI('GET', '/projects');
    console.log(`${colors.cyan}Response:${colors.reset}`);
    console.log(JSON.stringify(result.data, null, 2));
    console.log();
  } catch (error) {
    console.error(`${colors.red}Error: ${error.message}${colors.reset}\n`);
  }

  // Test 2: Create a test journey (if we had a project ID)
  console.log(`${colors.bright}Test 2: Create Test Data${colors.reset}`);
  try {
    const projectResult = await callAPI('POST', '/projects', {
      name: `Holy Grail Test Project ${Date.now()}`,
      description: 'Testing the holy grail endpoint'
    });
    console.log(`${colors.cyan}Project Response:${colors.reset}`);
    console.log(JSON.stringify(projectResult.data, null, 2));
    console.log();
  } catch (error) {
    console.error(`${colors.red}Error: ${error.message}${colors.reset}\n`);
  }

  // Test 3: Holy Grail - Get Journey Details
  console.log(`${colors.bright}${colors.yellow}Test 3: HOLY GRAIL - GET /testsuites/{journeyId}${colors.reset}`);
  
  // Try with a test journey ID
  const testJourneyIds = [
    '123',  // Test ID
    'test-journey-123',  // Another test ID
    '1'  // Simple test ID
  ];

  for (const journeyId of testJourneyIds) {
    console.log(`\n${colors.blue}Testing with journeyId: ${journeyId}${colors.reset}`);
    try {
      const result = await callAPI('GET', `/testsuites/${journeyId}`);
      console.log(`${colors.cyan}Response:${colors.reset}`);
      console.log(JSON.stringify(result.data, null, 2));
      
      if (result.status === 200) {
        console.log(`${colors.green}✓ SUCCESS: Holy grail endpoint working!${colors.reset}`);
        
        // Check if response contains steps and checkpoints
        if (result.data && (result.data.steps || result.data.checkpoints || result.data.testSteps || result.data.testCases)) {
          console.log(`${colors.green}✓ Response contains test structure data${colors.reset}`);
        }
      }
    } catch (error) {
      console.error(`${colors.red}Error: ${error.message}${colors.reset}`);
    }
  }

  // Test 4: Check API Gateway routes
  console.log(`\n${colors.bright}Test 4: Check Available Routes${colors.reset}`);
  const testRoutes = [
    { method: 'GET', path: '/testsuites' },
    { method: 'GET', path: '/journeys' },
    { method: 'GET', path: '/testsuites/latest_status' },
    { method: 'POST', path: '/testsuites' },
    { method: 'GET', path: '/' }
  ];

  for (const route of testRoutes) {
    try {
      const result = await callAPI(route.method, route.path);
      console.log(`${route.method} ${route.path}: ${result.status}`);
    } catch (error) {
      console.log(`${route.method} ${route.path}: Error - ${error.message}`);
    }
  }

  console.log(`\n${colors.bright}${colors.cyan}Test Summary${colors.reset}`);
  console.log('The holy grail endpoint GET /testsuites/{journeyId} should return:');
  console.log('- Journey/TestSuite details');
  console.log('- List of checkpoints/test cases');
  console.log('- List of steps within each checkpoint');
  console.log('\nIf the Lambda functions are properly deployed and connected to the real Virtuoso API,');
  console.log('this endpoint will return the complete test structure.');
}

// Run the test
testHolyGrail().catch(error => {
  console.error(`\n${colors.red}Fatal error: ${error.message}${colors.reset}`);
  process.exit(1);
});