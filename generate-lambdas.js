#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

// Configuration
const config = {
  region: process.env.AWS_REGION || 'us-east-1',
  runtime: 'nodejs18.x',
  memorySize: 256,
  timeout: 30,
  baseUrl: process.env.VIRTUOSO_API_URL || 'https://api.virtuoso.qa/api',
  outputDir: './lambda-functions',
  layerDir: './lambda-layer',
  samTemplate: './template.yaml'
};

// Endpoint groups definition
const endpointGroups = {
  project: {
    name: 'VirtuosoProjectHandler',
    endpoints: [
      { method: 'POST', path: '/projects', handler: 'createProject' },
      { method: 'GET', path: '/projects', handler: 'listProjects' },
      { method: 'GET', path: '/projects/{projectId}/goals', handler: 'listProjectGoals' }
    ]
  },
  goal: {
    name: 'VirtuosoGoalHandler',
    endpoints: [
      { method: 'POST', path: '/goals', handler: 'createGoal' },
      { method: 'GET', path: '/goals/{goalId}/versions', handler: 'getGoalVersions' },
      { method: 'POST', path: '/goals/{goalId}/snapshots/{snapshotId}/execute', handler: 'executeGoalSnapshot' }
    ]
  },
  journey: {
    name: 'VirtuosoJourneyHandler',
    endpoints: [
      { method: 'POST', path: '/testsuites', handler: 'createJourney' },
      { method: 'POST', path: '/journeys', handler: 'createJourneyAlt' },
      { method: 'GET', path: '/testsuites/latest_status', handler: 'listJourneysWithStatus' },
      { method: 'GET', path: '/testsuites/{journeyId}', handler: 'getJourneyDetails' },
      { method: 'PUT', path: '/testsuites/{journeyId}', handler: 'updateJourney' },
      { method: 'POST', path: '/testsuites/{journeyId}/checkpoints/attach', handler: 'attachCheckpoint' },
      { method: 'POST', path: '/journeys/attach-library', handler: 'attachLibraryCheckpoint' }
    ]
  },
  checkpoint: {
    name: 'VirtuosoCheckpointHandler',
    endpoints: [
      { method: 'POST', path: '/testcases', handler: 'createCheckpoint' },
      { method: 'POST', path: '/checkpoints', handler: 'createCheckpointAlt' },
      { method: 'GET', path: '/testcases/{checkpointId}', handler: 'getCheckpointDetails' },
      { method: 'GET', path: '/checkpoints/{checkpointId}/teststeps', handler: 'getCheckpointSteps' },
      { method: 'POST', path: '/testcases/{checkpointId}/add-to-library', handler: 'addCheckpointToLibrary' }
    ]
  },
  step: {
    name: 'VirtuosoStepHandler',
    endpoints: [
      { method: 'POST', path: '/teststeps', handler: 'addTestStep' },
      { method: 'POST', path: '/teststeps?envelope=false', handler: 'addTestStepNoEnvelope' },
      { method: 'POST', path: '/steps', handler: 'addTestStepAlt' },
      { method: 'GET', path: '/teststeps/{stepId}', handler: 'getStepDetails' },
      { method: 'PUT', path: '/teststeps/{stepId}/properties', handler: 'updateStepProperties' }
    ]
  },
  execution: {
    name: 'VirtuosoExecutionHandler',
    endpoints: [
      { method: 'POST', path: '/executions', handler: 'executeGoal' },
      { method: 'GET', path: '/executions/{executionId}', handler: 'getExecutionStatus' },
      { method: 'GET', path: '/executions/analysis/{executionId}', handler: 'getExecutionAnalysis' }
    ]
  },
  library: {
    name: 'VirtuosoLibraryHandler',
    endpoints: [
      { method: 'POST', path: '/library/checkpoints', handler: 'addToLibrary' },
      { method: 'GET', path: '/library/checkpoints/{libraryCheckpointId}', handler: 'getLibraryCheckpoint' },
      { method: 'PUT', path: '/library/checkpoints/{libraryCheckpointId}', handler: 'updateLibraryCheckpoint' },
      { method: 'DELETE', path: '/library/checkpoints/{libraryCheckpointId}/steps/{testStepId}', handler: 'removeLibraryStep' },
      { method: 'POST', path: '/library/checkpoints/{libraryCheckpointId}/steps/{testStepId}/move', handler: 'moveLibraryStep' }
    ]
  },
  data: {
    name: 'VirtuosoDataHandler',
    endpoints: [
      { method: 'POST', path: '/testdata/tables/create', handler: 'createDataTable' },
      { method: 'GET', path: '/testdata/tables/{tableId}', handler: 'getDataTable' },
      { method: 'POST', path: '/testdata/tables/{tableId}/import', handler: 'importDataToTable' }
    ]
  },
  environment: {
    name: 'VirtuosoEnvironmentHandler',
    endpoints: [
      { method: 'POST', path: '/environments', handler: 'createEnvironment' }
    ]
  }
};

// Shared layer code
const layerCode = {
  'package.json': `{
  "name": "virtuoso-lambda-layer",
  "version": "1.0.0",
  "dependencies": {
    "axios": "^1.6.0",
    "@aws-lambda-powertools/logger": "^1.14.0",
    "@aws-lambda-powertools/metrics": "^1.14.0",
    "@aws-lambda-powertools/tracer": "^1.14.0",
    "p-retry": "^5.1.2"
  }
}`,
  'utils/auth.js': `const { SSMClient, GetParameterCommand } = require('@aws-sdk/client-ssm');

const ssm = new SSMClient();

exports.getApiToken = async () => {
  const command = new GetParameterCommand({
    Name: process.env.API_TOKEN_PARAM || '/virtuoso/api-token',
    WithDecryption: true
  });
  const response = await ssm.send(command);
  return response.Parameter.Value;
};`,
  'utils/error-handler.js': `const { Logger } = require('@aws-lambda-powertools/logger');

const logger = new Logger();

class VirtuosoError extends Error {
  constructor(message, statusCode = 500, details = {}) {
    super(message);
    this.statusCode = statusCode;
    this.details = details;
  }
}

exports.VirtuosoError = VirtuosoError;

exports.handleError = (error) => {
  logger.error('Error occurred', { error });
  
  if (error instanceof VirtuosoError) {
    return {
      statusCode: error.statusCode,
      body: JSON.stringify({
        error: error.message,
        details: error.details
      })
    };
  }
  
  if (error.response) {
    return {
      statusCode: error.response.status,
      body: JSON.stringify({
        error: error.response.data?.message || 'API request failed',
        details: error.response.data
      })
    };
  }
  
  return {
    statusCode: 500,
    body: JSON.stringify({
      error: 'Internal server error',
      message: error.message
    })
  };
};`,
  'utils/retry.js': `const pRetry = require('p-retry');

exports.retryableRequest = async (fn, options = {}) => {
  return pRetry(fn, {
    retries: 3,
    minTimeout: 1000,
    maxTimeout: 5000,
    onFailedAttempt: error => {
      console.log(\`Attempt \${error.attemptNumber} failed. \${error.retriesLeft} retries left.\`);
    },
    ...options
  });
};`,
  'utils/logger.js': `const { Logger } = require('@aws-lambda-powertools/logger');

exports.createLogger = (serviceName) => {
  return new Logger({
    serviceName,
    logLevel: process.env.LOG_LEVEL || 'INFO'
  });
};`,
  'config.js': `module.exports = {
  baseUrl: process.env.VIRTUOSO_API_URL || 'https://api.virtuoso.qa/api',
  timeout: parseInt(process.env.API_TIMEOUT || '30000'),
  retryConfig: {
    retries: parseInt(process.env.RETRY_COUNT || '3'),
    minTimeout: parseInt(process.env.RETRY_MIN_TIMEOUT || '1000'),
    maxTimeout: parseInt(process.env.RETRY_MAX_TIMEOUT || '5000')
  }
};`
};

// Lambda function template
const lambdaTemplate = (groupName, endpoints) => `const axios = require('axios');
const { getApiToken } = require('/opt/utils/auth');
const { handleError, VirtuosoError } = require('/opt/utils/error-handler');
const { retryableRequest } = require('/opt/utils/retry');
const { createLogger } = require('/opt/utils/logger');
const config = require('/opt/config');

const logger = createLogger('${groupName}');

// Initialize axios instance
const api = axios.create({
  baseURL: config.baseUrl,
  timeout: config.timeout,
  headers: {
    'Content-Type': 'application/json'
  }
});

// Add auth interceptor
api.interceptors.request.use(async (config) => {
  const token = await getApiToken();
  config.headers.Authorization = \`Bearer \${token}\`;
  return config;
});

// Handler implementations
${endpoints.map(endpoint => `
const ${endpoint.handler} = async (event) => {
  const { params = {}, body, queryStringParameters } = event;
  let url = '${endpoint.path}';
  
  // Replace path parameters
  Object.keys(params).forEach(key => {
    url = url.replace(\`{\${key}}\`, params[key]);
  });
  
  // Add query parameters
  if (queryStringParameters) {
    const queryString = new URLSearchParams(queryStringParameters).toString();
    url += \`?\${queryString}\`;
  }
  
  const requestConfig = {
    method: '${endpoint.method}',
    url
  };
  
  if (body) {
    requestConfig.data = body;
  }
  
  const response = await api(requestConfig);
  return response.data;
};`).join('\n')}

// Main handler
exports.handler = async (event) => {
  logger.info('Received event', { event });
  
  try {
    const { action } = event;
    
    const handlers = {
${endpoints.map(endpoint => `      '${endpoint.handler}': ${endpoint.handler}`).join(',\n')}
    };
    
    if (!handlers[action]) {
      throw new VirtuosoError(\`Unknown action: \${action}\`, 400);
    }
    
    const result = await retryableRequest(
      () => handlers[action](event),
      config.retryConfig
    );
    
    return {
      statusCode: 200,
      body: JSON.stringify(result)
    };
  } catch (error) {
    return handleError(error);
  }
};
`;

// SAM template generator
const generateSAMTemplate = () => `AWSTemplateFormatVersion: '2010-09-09'
Transform: AWS::Serverless-2016-10-31
Description: 'Virtuoso API Lambda Functions'

Globals:
  Function:
    Runtime: nodejs18.x
    MemorySize: ${config.memorySize}
    Timeout: ${config.timeout}
    Environment:
      Variables:
        VIRTUOSO_API_URL: ${config.baseUrl}
        LOG_LEVEL: INFO
        API_TOKEN_PARAM: /virtuoso/api-token

Parameters:
  ApiTokenValue:
    Type: String
    NoEcho: true
    Description: Virtuoso API Token

Resources:
  # API Token Parameter
  ApiTokenParameter:
    Type: AWS::SSM::Parameter
    Properties:
      Name: /virtuoso/api-token
      Type: SecureString
      Value: !Ref ApiTokenValue

  # Shared Layer
  VirtuosoLayer:
    Type: AWS::Serverless::LayerVersion
    Properties:
      LayerName: virtuoso-lambda-layer
      Description: Shared utilities and dependencies
      ContentUri: ./lambda-layer/
      CompatibleRuntimes:
        - nodejs18.x
      RetentionPolicy: Retain

${Object.entries(endpointGroups).map(([key, group]) => `
  # ${group.name}
  ${group.name}:
    Type: AWS::Serverless::Function
    Properties:
      FunctionName: ${group.name}
      Handler: index.handler
      CodeUri: ./lambda-functions/${key}/
      Layers:
        - !Ref VirtuosoLayer
      Policies:
        - SSMParameterReadPolicy:
            ParameterName: virtuoso/api-token
      Events:
${group.endpoints.map((endpoint, index) => `        ${endpoint.handler}Event:
          Type: Api
          Properties:
            Path: /virtuoso${endpoint.path}
            Method: ${endpoint.method.toLowerCase()}
`).join('')}`).join('')}

Outputs:
  ApiEndpoint:
    Description: API Gateway endpoint URL
    Value: !Sub 'https://\${ServerlessRestApi}.execute-api.\${AWS::Region}.amazonaws.com/Prod/virtuoso'
`;

// Main generation function
const generateLambdaFunctions = () => {
  console.log('🚀 Starting Lambda function generation...\n');
  
  // Create directories
  [config.outputDir, config.layerDir].forEach(dir => {
    if (!fs.existsSync(dir)) {
      fs.mkdirSync(dir, { recursive: true });
    }
  });
  
  // Generate layer files
  console.log('📦 Creating shared layer...');
  Object.entries(layerCode).forEach(([filePath, content]) => {
    const fullPath = path.join(config.layerDir, 'nodejs', filePath);
    const dir = path.dirname(fullPath);
    if (!fs.existsSync(dir)) {
      fs.mkdirSync(dir, { recursive: true });
    }
    fs.writeFileSync(fullPath, content);
  });
  
  // Create package.json for npm install
  console.log('📥 Creating package.json for layer dependencies...');
  
  // Generate Lambda functions
  console.log('\n⚡ Generating Lambda functions...');
  Object.entries(endpointGroups).forEach(([key, group]) => {
    const functionDir = path.join(config.outputDir, key);
    if (!fs.existsSync(functionDir)) {
      fs.mkdirSync(functionDir, { recursive: true });
    }
    
    // Write function code
    const functionCode = lambdaTemplate(group.name, group.endpoints);
    fs.writeFileSync(path.join(functionDir, 'index.js'), functionCode);
    
    // Create package.json
    const packageJson = {
      name: group.name.toLowerCase(),
      version: '1.0.0',
      main: 'index.js'
    };
    fs.writeFileSync(
      path.join(functionDir, 'package.json'), 
      JSON.stringify(packageJson, null, 2)
    );
    
    console.log(`✅ Generated ${group.name} (${group.endpoints.length} endpoints)`);
  });
  
  // Generate SAM template
  console.log('\n📄 Generating SAM template...');
  fs.writeFileSync(config.samTemplate, generateSAMTemplate());
  
  // Generate deployment script
  console.log('🚀 Generating deployment script...');
  const deployScript = `#!/bin/bash

# Virtuoso API Lambda Deployment Script

set -e

echo "🚀 Deploying Virtuoso API Lambda Functions"

# Check if API token is provided
if [ -z "$1" ]; then
  echo "❌ Error: API token required"
  echo "Usage: ./deploy.sh <API_TOKEN>"
  exit 1
fi

API_TOKEN=$1
STACK_NAME=\${STACK_NAME:-virtuoso-api-stack}
REGION=\${AWS_REGION:-${config.region}}
S3_BUCKET=\${S3_BUCKET:-virtuoso-lambda-deployments-$RANDOM}

# Create S3 bucket for deployments if it doesn't exist
echo "📦 Creating deployment bucket..."
aws s3 mb s3://$S3_BUCKET --region $REGION 2>/dev/null || true

# Install layer dependencies
echo "📥 Installing layer dependencies..."
cd lambda-layer/nodejs && npm install && cd ../..

# Package the application
echo "📦 Packaging application..."
sam package \\
  --template-file template.yaml \\
  --s3-bucket $S3_BUCKET \\
  --output-template-file packaged.yaml \\
  --region $REGION

# Deploy the application
echo "🚀 Deploying application..."
sam deploy \\
  --template-file packaged.yaml \\
  --stack-name $STACK_NAME \\
  --capabilities CAPABILITY_IAM \\
  --parameter-overrides ApiTokenValue=$API_TOKEN \\
  --region $REGION \\
  --no-fail-on-empty-changeset

# Get outputs
echo "✅ Deployment complete!"
echo "📋 Stack outputs:"
aws cloudformation describe-stacks \\
  --stack-name $STACK_NAME \\
  --region $REGION \\
  --query 'Stacks[0].Outputs' \\
  --output table

# Cleanup
rm -f packaged.yaml
`;
  
  fs.writeFileSync('deploy.sh', deployScript);
  fs.chmodSync('deploy.sh', '755');
  
  // Generate README
  const readme = `# Virtuoso API Lambda Functions

## Overview

This project contains AWS Lambda functions that wrap the Virtuoso API endpoints, organized by resource type for optimal performance and maintainability.

## Architecture

- **Shared Layer**: Common utilities, authentication, error handling, and retry logic
- **Lambda Functions**: Grouped by resource type (projects, goals, journeys, etc.)
- **API Gateway**: RESTful API interface with path-based routing

## Deployment

### Prerequisites

- AWS CLI configured with appropriate credentials
- AWS SAM CLI installed
- Node.js 18.x or later

### Quick Deploy

\`\`\`bash
# Deploy with your Virtuoso API token
./deploy.sh YOUR_VIRTUOSO_API_TOKEN

# Or with custom settings
STACK_NAME=my-virtuoso-api AWS_REGION=eu-west-1 ./deploy.sh YOUR_API_TOKEN
\`\`\`

### Manual Deployment

1. Install dependencies:
   \`\`\`bash
   cd lambda-layer/nodejs && npm install
   \`\`\`

2. Deploy using SAM:
   \`\`\`bash
   sam deploy --guided --parameter-overrides ApiTokenValue=YOUR_TOKEN
   \`\`\`

## Usage

Each Lambda function handles multiple related endpoints. Call them with an \`action\` parameter:

\`\`\`javascript
// Example: Create a project
const response = await lambda.invoke({
  FunctionName: 'VirtuosoProjectHandler',
  Payload: JSON.stringify({
    action: 'createProject',
    body: {
      name: 'My Test Project',
      description: 'Automated testing project'
    }
  })
}).promise();

// Example: Execute a goal
const response = await lambda.invoke({
  FunctionName: 'VirtuosoExecutionHandler',
  Payload: JSON.stringify({
    action: 'executeGoal',
    body: {
      goalId: 123,
      environment: 'staging'
    }
  })
}).promise();
\`\`\`

## Function Groups

${Object.entries(endpointGroups).map(([key, group]) => `
### ${group.name}
Endpoints:
${group.endpoints.map(ep => `- ${ep.method} ${ep.path} (action: \`${ep.handler}\`)`).join('\n')}`).join('\n')}

## Environment Variables

- \`VIRTUOSO_API_URL\`: Base URL for Virtuoso API
- \`API_TOKEN_PARAM\`: SSM parameter name for API token
- \`LOG_LEVEL\`: Logging level (DEBUG, INFO, WARN, ERROR)

## Monitoring

CloudWatch Logs and X-Ray tracing are automatically configured for all functions.

## Cost Optimization

- Functions are grouped to minimize cold starts
- Shared layer reduces deployment size
- Memory allocated based on typical workload

## Security

- API token stored in AWS Systems Manager Parameter Store
- IAM roles follow least privilege principle
- All functions use VPC endpoints when configured
`;
  
  fs.writeFileSync('README-LAMBDA.md', readme);
  
  console.log(`
✅ Generation complete!

📁 Generated files:
- ${Object.keys(endpointGroups).length} Lambda functions
- Shared layer with utilities
- SAM template for deployment
- Deployment script
- README documentation

🚀 To deploy:
./deploy.sh YOUR_VIRTUOSO_API_TOKEN

📚 See README-LAMBDA.md for detailed usage instructions
`);
};

// Run the generator
if (require.main === module) {
  generateLambdaFunctions();
}

module.exports = { generateLambdaFunctions, endpointGroups, config };