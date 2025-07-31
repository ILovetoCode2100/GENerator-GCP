import { APIGatewayRequestSimpleAuthorizerHandlerV2 } from 'aws-lambda';

/**
 * Custom authorizer for Bearer token validation
 * 
 * This is a simple authorizer that validates Bearer tokens.
 * In production, you would validate against a user database,
 * JWT verification, or external auth service.
 */
export const handler: APIGatewayRequestSimpleAuthorizerHandlerV2 = async (event) => {
  console.log('Authorizer event:', JSON.stringify(event, null, 2));

  try {
    // Extract authorization header
    const authHeader = event.headers?.authorization || event.headers?.Authorization;
    
    if (!authHeader) {
      console.log('No authorization header found');
      return { isAuthorized: false };
    }

    // Check for Bearer token format
    if (!authHeader.startsWith('Bearer ')) {
      console.log('Invalid authorization header format');
      return { isAuthorized: false };
    }

    const token = authHeader.substring(7); // Remove 'Bearer ' prefix
    
    // Validate token (simplified for example)
    // In production, implement proper token validation:
    // - JWT verification
    // - Database lookup
    // - External auth service call
    // - Token expiration check
    
    if (!token || token.length < 10) {
      console.log('Invalid token format');
      return { isAuthorized: false };
    }

    // For demo purposes, accept any token that starts with 'vrt_'
    // In production, properly validate the token
    const isValid = token.startsWith('vrt_') || process.env.ALLOW_ANY_TOKEN === 'true';
    
    if (!isValid) {
      console.log('Token validation failed');
      return { isAuthorized: false };
    }

    console.log('Token validated successfully');
    
    // Return authorization response with context
    return {
      isAuthorized: true,
      context: {
        // Add user information that can be used in Lambda functions
        userId: 'user-' + token.substring(0, 8),
        tokenType: 'bearer',
        // Additional context can be added here
      },
    };
    
  } catch (error) {
    console.error('Authorizer error:', error);
    // On error, deny access
    return { isAuthorized: false };
  }
};