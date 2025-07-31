import * as cdk from 'aws-cdk-lib';
import { Template } from 'aws-cdk-lib/assertions';
import { VirtuosoApiStack } from '../lib/virtuoso-api-stack';

describe('VirtuosoApiStack', () => {
  test('creates all required Lambda functions', () => {
    const app = new cdk.App();
    const stack = new VirtuosoApiStack(app, 'TestStack');
    const template = Template.fromStack(stack);

    // Test that all 19 endpoint Lambda functions are created
    const expectedFunctions = [
      'virtuoso-get-user',
      'virtuoso-list-projects', 
      'virtuoso-list-goals',
      'virtuoso-create-project',
      'virtuoso-create-goal',
      'virtuoso-get-goal-versions',
      'virtuoso-execute-goal',
      'virtuoso-execute-snapshot',
      'virtuoso-create-journey',
      'virtuoso-create-checkpoint',
      'virtuoso-get-checkpoint-steps',
      'virtuoso-create-step',
      'virtuoso-start-execution',
      'virtuoso-get-execution-status',
      'virtuoso-get-execution-analysis',
      'virtuoso-create-library-checkpoint',
      'virtuoso-list-library-checkpoints',
      'virtuoso-create-test-data-table',
      'virtuoso-create-environment'
    ];

    expectedFunctions.forEach(functionName => {
      template.hasResourceProperties('AWS::Lambda::Function', {
        FunctionName: functionName,
        Runtime: 'nodejs20.x',
        Architectures: ['arm64']
      });
    });

    // Plus the authorizer function
    template.hasResourceProperties('AWS::Lambda::Function', {
      FunctionName: 'virtuoso-api-authorizer',
      Runtime: 'nodejs20.x',
      Architectures: ['arm64']
    });
  });

  test('creates HTTP API Gateway with correct configuration', () => {
    const app = new cdk.App();
    const stack = new VirtuosoApiStack(app, 'TestStack');
    const template = Template.fromStack(stack);

    // Test API Gateway creation
    template.hasResourceProperties('AWS::ApiGatewayV2::Api', {
      Name: 'virtuoso-api-proxy',
      ProtocolType: 'HTTP'
    });

    // Test CORS configuration
    template.hasResourceProperties('AWS::ApiGatewayV2::Api', {
      CorsConfiguration: {
        AllowCredentials: true,
        AllowHeaders: [
          'Content-Type',
          'X-Amz-Date',
          'Authorization',
          'X-Api-Key',
          'X-Amz-Security-Token',
          'X-Amz-User-Agent',
          'X-Requested-With'
        ],
        AllowMethods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
        AllowOrigins: ['*']
      }
    });
  });

  test('creates Secrets Manager secret', () => {
    const app = new cdk.App();
    const stack = new VirtuosoApiStack(app, 'TestStack');
    const template = Template.fromStack(stack);

    template.hasResourceProperties('AWS::SecretsManager::Secret', {
      Name: 'virtuoso-api-config',
      Description: 'Configuration secrets for Virtuoso API proxy'
    });
  });

  test('creates IAM role with correct permissions', () => {
    const app = new cdk.App();
    const stack = new VirtuosoApiStack(app, 'TestStack');
    const template = Template.fromStack(stack);

    template.hasResourceProperties('AWS::IAM::Role', {
      AssumeRolePolicyDocument: {
        Statement: [
          {
            Action: 'sts:AssumeRole',
            Effect: 'Allow',
            Principal: {
              Service: 'lambda.amazonaws.com'
            }
          }
        ]
      },
      ManagedPolicyArns: [
        'arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole',
        'arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole'
      ]
    });
  });

  test('creates CloudWatch Log Groups', () => {
    const app = new cdk.App();
    const stack = new VirtuosoApiStack(app, 'TestStack');
    const template = Template.fromStack(stack);

    // Test that log groups are created for Lambda functions
    template.hasResourceProperties('AWS::Logs::LogGroup', {
      LogGroupName: '/aws/lambda/virtuoso-get-user',
      RetentionInDays: 7
    });

    template.hasResourceProperties('AWS::Logs::LogGroup', {
      LogGroupName: '/aws/lambda/virtuoso-api-authorizer',
      RetentionInDays: 7
    });
  });

  test('creates API Gateway routes with correct paths', () => {
    const app = new cdk.App();
    const stack = new VirtuosoApiStack(app, 'TestStack');
    const template = Template.fromStack(stack);

    // Test some key routes
    template.hasResourceProperties('AWS::ApiGatewayV2::Route', {
      RouteKey: 'GET /api/user'
    });

    template.hasResourceProperties('AWS::ApiGatewayV2::Route', {
      RouteKey: 'POST /api/projects'
    });

    template.hasResourceProperties('AWS::ApiGatewayV2::Route', {
      RouteKey: 'GET /api/projects/{project_id}/goals'
    });
  });

  test('creates API Gateway stage with throttling', () => {
    const app = new cdk.App();
    const stack = new VirtuosoApiStack(app, 'TestStack');
    const template = Template.fromStack(stack);

    template.hasResourceProperties('AWS::ApiGatewayV2::Stage', {
      StageName: 'prod',
      AutoDeploy: true,
      ThrottleSettings: {
        RateLimit: 1000,
        BurstLimit: 2000
      }
    });
  });

  test('outputs important values', () => {
    const app = new cdk.App();
    const stack = new VirtuosoApiStack(app, 'TestStack');
    const template = Template.fromStack(stack);

    // Test that outputs are created
    template.hasOutput('ApiGatewayUrl', {});
    template.hasOutput('ApiGatewayId', {});
    template.hasOutput('SecretsManagerArn', {});
    template.hasOutput('LambdaExecutionRoleArn', {});
  });
});