import { APIGatewayProxyEventV2, APIGatewayProxyResultV2, Context } from 'aws-lambda';
import { SecretsManagerClient, GetSecretValueCommand } from '@aws-sdk/client-secrets-manager';

// Initialize AWS clients
const secretsClient = new SecretsManagerClient({});

// Cache for API key to avoid repeated Secrets Manager calls
let cachedApiKey: string | null = null;
let cacheExpiry = 0;

/**
 * Get Virtuoso API key from Secrets Manager with caching
 */
async function getApiKey(): Promise<string> {
  const now = Date.now();
  
  // Return cached key if still valid (5 minute cache)
  if (cachedApiKey && now < cacheExpiry) {
    return cachedApiKey;
  }

  try {
    const secretName = process.env.VIRTUOSO_API_KEY_SECRET_NAME;
    if (!secretName) {
      throw new Error('VIRTUOSO_API_KEY_SECRET_NAME environment variable not set');
    }

    const command = new GetSecretValueCommand({ SecretId: secretName });
    const response = await secretsClient.send(command);
    
    if (!response.SecretString) {
      throw new Error('Secret value is empty');
    }

    // Handle both JSON and plain string secrets
    let apiKey: string;
    try {
      const secret = JSON.parse(response.SecretString);
      apiKey = secret.apiKey || secret.api_key || secret.key;
    } catch {
      // If not JSON, treat as plain string
      apiKey = response.SecretString;
    }
    
    if (!apiKey) {
      throw new Error('API key not found in secret');
    }
    
    cachedApiKey = apiKey;
    cacheExpiry = now + (5 * 60 * 1000); // Cache for 5 minutes
    
    return cachedApiKey;
  } catch (error) {
    console.error('Failed to retrieve API key:', error);
    throw new Error('Unable to authenticate with Virtuoso API');
  }
}

/**
 * Lambda handler for POST /api/goals/{goal_id}/execute
 * 
 * Simplifications applied:
 * - Request: Only accepts optional startingUrl in body (removed initialData, headers, etc.)
 * - Response: Returns only jobId and status (removed verbose execution details)
 * - Error handling: Simplified error messages
 */
export const handler = async (
  event: APIGatewayProxyEventV2,
  context: Context
): Promise<APIGatewayProxyResultV2> => {
  console.log('Event:', JSON.stringify(event, null, 2));

  try {
    // Extract goal ID from path parameters
    const goalId = event.pathParameters?.goal_id;
    if (!goalId) {
      return {
        statusCode: 400,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ error: 'Goal ID is required' }),
      };
    }

    // Parse request body (may be undefined for empty body)
    let requestBody: any = {};
    if (event.body) {
      try {
        requestBody = JSON.parse(event.body);
      } catch (e) {
        return {
          statusCode: 400,
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ error: 'Invalid JSON in request body' }),
        };
      }
    }

    // Extract simplified parameters
    const { startingUrl } = requestBody;

    // Get API key and base URL
    const apiKey = await getApiKey();
    const baseUrl = process.env.VIRTUOSO_API_BASE_URL || 'https://api.virtuoso.qa';

    // Build Virtuoso API request body
    // The original API expects more fields, but we provide defaults
    const virtuosoRequestBody = {
      goalId,
      startingUrl: startingUrl || null,
      // Default values for fields we're hiding from the simplified API
      includeDataDrivenJourneys: true,
      includeDisabledJourneys: false,
      parallelExecution: true,
      maxParallelExecutions: 5,
      environment: 'production',
    };

    // Make request to Virtuoso API
    const virtuosoUrl = `${baseUrl}/goals/${goalId}/execute`;
    
    console.log('Calling Virtuoso API:', virtuosoUrl);
    
    const response = await fetch(virtuosoUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${apiKey}`,
        'User-Agent': 'VirtuosoSimplifiedAPI/1.0',
      },
      body: JSON.stringify(virtuosoRequestBody),
    });

    const responseText = await response.text();
    console.log('Virtuoso API response:', response.status, responseText);

    // Handle non-2xx responses
    if (!response.ok) {
      let errorMessage = 'Failed to execute goal';
      
      try {
        const errorData = JSON.parse(responseText);
        errorMessage = errorData.message || errorData.error || errorMessage;
      } catch (e) {
        // If response isn't JSON, use status text
        errorMessage = `${response.status} ${response.statusText}`;
      }

      return {
        statusCode: response.status,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ error: errorMessage }),
      };
    }

    // Parse successful response
    let responseData: any;
    try {
      responseData = JSON.parse(responseText);
    } catch (e) {
      console.error('Failed to parse Virtuoso response:', e);
      return {
        statusCode: 502,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ error: 'Invalid response from Virtuoso API' }),
      };
    }

    // Simplify response - extract only essential fields
    const simplifiedResponse = {
      jobId: responseData.jobId || responseData.id || responseData.executionId,
      status: responseData.status || 'started',
    };

    return {
      statusCode: 200,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(simplifiedResponse),
    };

  } catch (error) {
    console.error('Lambda execution error:', error);
    
    return {
      statusCode: 500,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ 
        error: 'Internal server error',
        // Include request ID for debugging
        requestId: context.awsRequestId,
      }),
    };
  }
};