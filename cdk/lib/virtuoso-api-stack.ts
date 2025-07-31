import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as apigatewayv2 from 'aws-cdk-lib/aws-apigatewayv2';
import * as apigatewayv2Integrations from 'aws-cdk-lib/aws-apigatewayv2-integrations';
import * as apigatewayv2Authorizers from 'aws-cdk-lib/aws-apigatewayv2-authorizers';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as lambdaNodejs from 'aws-cdk-lib/aws-lambda-nodejs';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as logs from 'aws-cdk-lib/aws-logs';
import * as secretsmanager from 'aws-cdk-lib/aws-secretsmanager';
import * as path from 'path';

export class VirtuosoApiStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // Create Secrets Manager secret for Virtuoso API key
    const apiKeySecret = new secretsmanager.Secret(this, 'VirtuosoApiKey', {
      description: 'Virtuoso API authentication key',
      secretName: 'virtuoso-api-key',
    });

    // Common Lambda environment variables
    const commonEnv = {
      VIRTUOSO_API_BASE_URL: process.env.VIRTUOSO_API_BASE_URL || 'https://api.virtuoso.qa',
      VIRTUOSO_API_KEY_SECRET_NAME: apiKeySecret.secretName,
      NODE_OPTIONS: '--enable-source-maps',
    };

    // Common Lambda configuration
    const commonLambdaProps = {
      runtime: lambda.Runtime.NODEJS_20_X,
      architecture: lambda.Architecture.ARM_64,
      timeout: cdk.Duration.seconds(30),
      memorySize: 512,
      environment: commonEnv,
      bundling: {
        minify: true,
        sourceMap: true,
        target: 'node20',
        format: lambdaNodejs.OutputFormat.ESM,
        mainFields: ['module', 'main'],
        externalModules: ['@aws-sdk/*'],
      },
      logRetention: logs.RetentionDays.ONE_WEEK,
    };

    // Create custom authorizer Lambda
    const authorizerLambda = new lambdaNodejs.NodejsFunction(this, 'AuthorizerFunction', {
      ...commonLambdaProps,
      entry: path.join(__dirname, '../lambda/authorizer.ts'),
      handler: 'handler',
      functionName: 'virtuoso-api-authorizer',
    });

    // Grant read access to secrets
    apiKeySecret.grantRead(authorizerLambda);

    // Create HTTP API
    const httpApi = new apigatewayv2.HttpApi(this, 'VirtuosoApi', {
      apiName: 'virtuoso-simplified-api',
      description: 'Simplified proxy API for Virtuoso',
      corsPreflight: {
        allowOrigins: ['*'],
        allowMethods: [
          apigatewayv2.CorsHttpMethod.GET,
          apigatewayv2.CorsHttpMethod.POST,
          apigatewayv2.CorsHttpMethod.PUT,
          apigatewayv2.CorsHttpMethod.DELETE,
          apigatewayv2.CorsHttpMethod.OPTIONS,
        ],
        allowHeaders: ['Authorization', 'Content-Type', 'X-Api-Key'],
        maxAge: cdk.Duration.hours(24),
      },
    });

    // Create custom authorizer
    const authorizer = new apigatewayv2Authorizers.HttpLambdaAuthorizer('CustomAuthorizer', authorizerLambda, {
      identitySource: ['$request.header.Authorization'],
      authorizerName: 'BearerTokenAuthorizer',
      responseTypes: [apigatewayv2Authorizers.HttpLambdaResponseType.SIMPLE],
      resultsCacheTtl: cdk.Duration.minutes(5),
    });

    // Define all API endpoints with their corresponding handler files
    const endpoints = [
      { method: 'GET', path: '/api/user', handler: 'get-user' },
      { method: 'GET', path: '/api/projects', handler: 'list-projects' },
      { method: 'GET', path: '/api/projects/{project_id}/goals', handler: 'list-goals' },
      { method: 'POST', path: '/api/projects', handler: 'create-project' },
      { method: 'POST', path: '/api/goals', handler: 'create-goal' },
      { method: 'GET', path: '/api/goals/{goal_id}/versions', handler: 'get-goal-versions' },
      { method: 'POST', path: '/api/goals/{goal_id}/execute', handler: 'execute-goal' },
      { method: 'POST', path: '/api/goals/{goal_id}/snapshots/{snapshot_id}/execute', handler: 'execute-snapshot' },
      { method: 'POST', path: '/api/journeys', handler: 'create-journey' },
      { method: 'POST', path: '/api/checkpoints', handler: 'create-checkpoint' },
      { method: 'GET', path: '/api/checkpoints/{checkpoint_id}/steps', handler: 'get-checkpoint-steps' },
      { method: 'POST', path: '/api/steps', handler: 'create-step' },
      { method: 'POST', path: '/api/executions', handler: 'start-execution' },
      { method: 'GET', path: '/api/executions/{execution_id}', handler: 'get-execution-status' },
      { method: 'GET', path: '/api/executions/{execution_id}/analysis', handler: 'get-execution-analysis' },
      { method: 'POST', path: '/api/library/checkpoints', handler: 'create-library-checkpoint' },
      { method: 'GET', path: '/api/library/checkpoints', handler: 'list-library-checkpoints' },
      { method: 'POST', path: '/api/testdata/tables', handler: 'create-test-data-table' },
      { method: 'POST', path: '/api/environments', handler: 'create-environment' },
    ];

    // Create Lambda functions and API routes for each endpoint
    endpoints.forEach(({ method, path: routePath, handler }) => {
      // Create Lambda function
      const lambdaFunction = new lambdaNodejs.NodejsFunction(this, `${handler}Function`, {
        ...commonLambdaProps,
        entry: path.join(__dirname, `../lambda/${handler}.ts`),
        handler: 'handler',
        functionName: `virtuoso-${handler}`,
      });

      // Grant read access to secrets
      apiKeySecret.grantRead(lambdaFunction);

      // Create Lambda integration
      const integration = new apigatewayv2Integrations.HttpLambdaIntegration(
        `${handler}Integration`,
        lambdaFunction
      );

      // Add route to API
      httpApi.addRoutes({
        path: routePath,
        methods: [method as apigatewayv2.HttpMethod],
        integration,
        authorizer,
      });
    });

    // Add throttling
    const throttle = new apigatewayv2.CfnStage(this, 'ThrottleStage', {
      apiId: httpApi.httpApiId,
      stageName: '$default',
      autoDeploy: true,
      defaultRouteSettings: {
        throttlingRateLimit: 1000,
        throttlingBurstLimit: 2000,
      },
    });

    // Output the API endpoint
    new cdk.CfnOutput(this, 'ApiEndpoint', {
      value: httpApi.url!,
      description: 'HTTP API Gateway endpoint URL',
    });

    new cdk.CfnOutput(this, 'ApiKeySecretName', {
      value: apiKeySecret.secretName,
      description: 'Secrets Manager secret name for Virtuoso API key',
    });
  }
}