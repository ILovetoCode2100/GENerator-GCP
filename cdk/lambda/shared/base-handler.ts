import { APIGatewayProxyEventV2, APIGatewayProxyResultV2, Context } from 'aws-lambda';
import { SecretsManagerClient, GetSecretValueCommand } from '@aws-sdk/client-secrets-manager';
import axios, { AxiosResponse, AxiosError } from 'axios';
import { VirtuosoApiConfig, ApiResponse, ErrorResponse, LambdaHandler } from './types';

export abstract class BaseHandler {
  private secretsManager: SecretsManagerClient;
  private apiConfig: VirtuosoApiConfig | null = null;

  constructor() {
    this.secretsManager = new SecretsManagerClient({
      region: process.env.AWS_REGION || 'us-east-1'
    });
  }

  protected async getApiConfig(): Promise<VirtuosoApiConfig> {
    if (this.apiConfig) {
      return this.apiConfig;
    }

    try {
      const secretName = process.env.VIRTUOSO_API_KEY_SECRET_NAME;
      if (!secretName) {
        throw new Error('VIRTUOSO_API_KEY_SECRET_NAME environment variable not set');
      }

      const command = new GetSecretValueCommand({
        SecretId: secretName
      });

      const result = await this.secretsManager.send(command);

      if (!result.SecretString) {
        throw new Error('Secret value is empty');
      }

      // For simple API key storage, the secret might just be the key itself
      // or a JSON object with the key
      let apiKey: string;
      try {
        const parsed = JSON.parse(result.SecretString);
        apiKey = parsed.apiKey || parsed.key || result.SecretString;
      } catch {
        apiKey = result.SecretString;
      }

      this.apiConfig = {
        virtuosoApiBaseUrl: process.env.VIRTUOSO_API_BASE_URL || 'https://api.virtuoso.qa',
        organizationId: process.env.VIRTUOSO_ORGANIZATION_ID || 'default',
        apiKey
      };
      
      return this.apiConfig;
    } catch (error) {
      console.error('Failed to retrieve API configuration:', error);
      throw new Error('Failed to retrieve API configuration');
    }
  }

  protected async makeVirtuosoRequest<T = any>(
    method: string,
    path: string,
    data?: any,
    authToken?: string
  ): Promise<T> {
    const config = await this.getApiConfig();
    if (!config) {
      throw new Error('API configuration not available');
    }
    const url = `${config.virtuosoApiBaseUrl}${path}`;
    
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
      'User-Agent': 'VirtuosoAPIProxy/1.0'
    };

    // Use authorization token from request if provided, otherwise use configured API key
    if (authToken) {
      headers['Authorization'] = authToken;
    } else if (config.apiKey) {
      headers['Authorization'] = `Bearer ${config.apiKey}`;
    }

    const maxRetries = parseInt(process.env.RETRY_ATTEMPTS || '3');
    let lastError: any;

    for (let attempt = 0; attempt <= maxRetries; attempt++) {
      try {
        console.log(`Making request to ${url} (attempt ${attempt + 1}/${maxRetries + 1})`);
        
        const response: AxiosResponse<T> = await axios({
          method: method.toLowerCase() as any,
          url,
          data,
          headers,
          timeout: parseInt(process.env.TIMEOUT_MS || '30000'),
          validateStatus: (status: number) => status < 500 // Don't retry on 4xx errors
        });

        return response.data;
      } catch (error) {
        lastError = error;
        const axiosError = error as AxiosError;
        
        // Don't retry on authentication errors or client errors
        if (axiosError.response?.status && axiosError.response.status < 500) {
          throw error;
        }

        // Log retry attempt
        if (attempt < maxRetries) {
          console.warn(`Request failed (attempt ${attempt + 1}), retrying...`, {
            error: axiosError.message,
            status: axiosError.response?.status
          });
          
          // Exponential backoff
          await new Promise(resolve => setTimeout(resolve, Math.pow(2, attempt) * 1000));
        }
      }
    }

    throw lastError;
  }

  protected createResponse<T>(
    statusCode: number,
    body: T,
    headers?: Record<string, string>
  ): APIGatewayProxyResultV2 {
    const defaultHeaders = {
      'Content-Type': 'application/json',
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Headers': 'Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token',
      'Access-Control-Allow-Methods': 'GET,POST,PUT,DELETE,OPTIONS'
    };

    return {
      statusCode,
      headers: { ...defaultHeaders, ...headers },
      body: JSON.stringify(body)
    };
  }

  protected createErrorResponse(
    statusCode: number,
    error: string,
    message?: string,
    requestId?: string
  ): APIGatewayProxyResultV2 {
    const errorBody: ErrorResponse = {
      error,
      message,
      requestId,
      timestamp: new Date().toISOString()
    };

    return this.createResponse(statusCode, errorBody);
  }

  protected extractAuthToken(event: APIGatewayProxyEventV2): string | undefined {
    return event.headers?.authorization || event.headers?.Authorization;
  }

  protected extractPathParameters(event: APIGatewayProxyEventV2): Record<string, string> {
    return event.pathParameters ? 
      Object.fromEntries(
        Object.entries(event.pathParameters).filter(([_, v]) => v !== undefined)
      ) as Record<string, string> : {};
  }

  protected extractQueryParameters(event: APIGatewayProxyEventV2): Record<string, string> {
    return event.queryStringParameters ?
      Object.fromEntries(
        Object.entries(event.queryStringParameters).filter(([_, v]) => v !== undefined)
      ) as Record<string, string> : {};
  }

  protected parseRequestBody<T = any>(event: APIGatewayProxyEventV2): T | null {
    if (!event.body) {
      return null;
    }

    try {
      return JSON.parse(event.body) as T;
    } catch (error) {
      console.error('Failed to parse request body:', error);
      return null;
    }
  }

  protected async handleRequest(
    event: APIGatewayProxyEventV2,
    context: Context,
    handler: () => Promise<APIGatewayProxyResultV2>
  ): Promise<APIGatewayProxyResultV2> {
    try {
      console.log('Processing request:', {
        httpMethod: event.requestContext.http.method,
        path: event.requestContext.http.path,
        requestId: context.awsRequestId
      });

      return await handler();
    } catch (error) {
      console.error('Request processing failed:', error);
      
      const axiosError = error as AxiosError;
      if (axiosError.response) {
        // Forward Virtuoso API errors
        const statusCode = axiosError.response.status;
        const responseData = axiosError.response.data as any;
        
        return this.createErrorResponse(
          statusCode,
          responseData?.error || 'Virtuoso API error',
          responseData?.message || axiosError.message,
          context.awsRequestId
        );
      }

      // Internal server error
      return this.createErrorResponse(
        500,
        'Internal server error',
        error instanceof Error ? error.message : 'Unknown error',
        context.awsRequestId
      );
    }
  }

  abstract handle(event: APIGatewayProxyEventV2, context: Context): Promise<APIGatewayProxyResultV2>;
}

export function createHandler(HandlerClass: new () => BaseHandler): LambdaHandler {
  const handlerInstance = new HandlerClass();
  return (event: APIGatewayProxyEventV2, context: Context) => handlerInstance.handle(event, context);
}