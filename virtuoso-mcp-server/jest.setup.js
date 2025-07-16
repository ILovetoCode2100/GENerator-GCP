// Jest setup file for additional configurations
import { jest } from "@jest/globals";

// Increase timeout for longer running tests
jest.setTimeout(10000);

// Add custom matchers or global test utilities here
global.testHelpers = {
  // Helper function to create mock CLI responses
  createMockResponse: (data, success = true) => {
    return {
      success,
      data,
      raw: JSON.stringify(data),
    };
  },

  // Helper to create mock error responses
  createMockError: (message) => {
    return {
      success: false,
      error: message,
      raw: "",
    };
  },
};

// Mock console methods to reduce noise during tests
global.console = {
  ...console,
  error: jest.fn(),
  warn: jest.fn(),
  log: jest.fn(),
  // Keep info and debug for test debugging
  info: console.info,
  debug: console.debug,
};
