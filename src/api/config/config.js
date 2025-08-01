module.exports = {
  baseUrl: process.env.VIRTUOSO_API_URL || 'https://api.virtuoso.qa/api',
  timeout: parseInt(process.env.API_TIMEOUT || '30000'),
  retryConfig: {
    retries: parseInt(process.env.RETRY_COUNT || '3'),
    minTimeout: parseInt(process.env.RETRY_MIN_TIMEOUT || '1000'),
    maxTimeout: parseInt(process.env.RETRY_MAX_TIMEOUT || '5000')
  }
};