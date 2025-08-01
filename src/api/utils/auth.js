/**
 * Authentication Utility
 * Platform-agnostic authentication helper using the abstraction layer
 */

/**
 * Get API token using platform-specific secret manager
 * @param {Object} secretManager - Secret manager instance from platform factory
 * @returns {Promise<string>} API token
 */
async function getApiToken(secretManager) {
  if (!secretManager) {
    throw new Error('Secret manager not provided');
  }
  
  return secretManager.getApiToken();
}

/**
 * Get tenant-specific API token
 * @param {Object} secretManager - Secret manager instance
 * @param {string} tenantId - Tenant identifier
 * @returns {Promise<string>} Tenant API token
 */
async function getTenantApiToken(secretManager, tenantId) {
  if (!secretManager) {
    throw new Error('Secret manager not provided');
  }
  
  if (!tenantId) {
    throw new Error('Tenant ID not provided');
  }
  
  return secretManager.getTenantSecret(tenantId, 'api-token');
}

/**
 * Create authentication headers
 * @param {string} token - API token
 * @param {Object} additionalHeaders - Additional headers to include
 * @returns {Object} Headers object
 */
function createAuthHeaders(token, additionalHeaders = {}) {
  return {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
    ...additionalHeaders
  };
}

/**
 * Validate API token format
 * @param {string} token - API token to validate
 * @returns {boolean} True if valid
 */
function isValidToken(token) {
  if (!token || typeof token !== 'string') {
    return false;
  }
  
  // Basic validation - token should be non-empty
  // Add more specific validation if needed
  return token.trim().length > 0;
}

module.exports = {
  getApiToken,
  getTenantApiToken,
  createAuthHeaders,
  isValidToken
};

// Backward compatibility for AWS Lambda environments
// This allows existing Lambda functions to work without changes
if (typeof exports !== 'undefined') {
  exports.getApiToken = async () => {
    // Try to use the AWS SSM implementation directly for backward compatibility
    try {
      const { SSMClient, GetParameterCommand } = require('@aws-sdk/client-ssm');
      const ssm = new SSMClient();
      const command = new GetParameterCommand({
        Name: process.env.API_TOKEN_PARAM || '/virtuoso/api-token',
        WithDecryption: true
      });
      const response = await ssm.send(command);
      return response.Parameter.Value;
    } catch (error) {
      // If AWS SDK not available, throw error
      throw new Error('AWS SSM not available. Use platform factory for authentication.');
    }
  };
}