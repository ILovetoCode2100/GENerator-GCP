const pRetry = require('p-retry');

exports.retryableRequest = async (fn, options = {}) => {
  return pRetry(fn, {
    retries: 3,
    minTimeout: 1000,
    maxTimeout: 5000,
    onFailedAttempt: error => {
      console.log(`Attempt ${error.attemptNumber} failed. ${error.retriesLeft} retries left.`);
    },
    ...options
  });
};