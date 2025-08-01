/**
 * Platform Factory
 * Creates platform-specific implementations of interfaces
 */

// AWS implementations
const { LambdaRuntime } = require('./implementations/aws/lambda-runtime');
const { SSMSecretManager } = require('./implementations/aws/ssm-secret-manager');
const { PowerToolsLogger } = require('./implementations/aws/powertools-logger');

// Generic implementations
const { GenericRuntime } = require('./implementations/generic/generic-runtime');
const { EnvSecretManager } = require('./implementations/generic/env-secret-manager');
const { ConsoleLogger } = require('./implementations/generic/console-logger');
const { JsonConfigManager } = require('./implementations/generic/json-config-manager');

// Import interfaces for type checking
const { RuntimeInterface } = require('./interfaces/runtime.interface');
const { SecretManagerInterface } = require('./interfaces/secret.interface');
const { LoggerInterface } = require('./interfaces/logger.interface');
const { ConfigManagerInterface } = require('./interfaces/config.interface');

/**
 * Platform types enumeration
 */
const PlatformType = {
  AWS_LAMBDA: 'aws-lambda',
  BEDROCK: 'bedrock',
  GENERIC: 'generic',
  EXPRESS: 'express',
  FASTIFY: 'fastify',
  KOA: 'koa'
};

/**
 * Factory for creating platform-specific implementations
 */
class PlatformFactory {
  constructor() {
    this.platform = this.detectPlatform();
    this.instances = new Map();
  }

  /**
   * Detect the current platform
   * @returns {string} Platform identifier
   */
  detectPlatform() {
    // Check for explicit platform configuration
    if (process.env.VIRTUOSO_PLATFORM) {
      return process.env.VIRTUOSO_PLATFORM;
    }

    // Auto-detect AWS Lambda
    if (process.env.AWS_LAMBDA_FUNCTION_NAME) {
      return PlatformType.AWS_LAMBDA;
    }

    // Default to generic
    return PlatformType.GENERIC;
  }

  /**
   * Set the platform type
   * @param {string} platform - Platform identifier
   */
  setPlatform(platform) {
    this.platform = platform;
    this.instances.clear(); // Clear cached instances
  }

  /**
   * Get runtime implementation
   * @param {Object} config - Runtime configuration
   * @returns {RuntimeInterface} Runtime implementation
   */
  getRuntime(config = {}) {
    const key = `runtime_${JSON.stringify(config)}`;
    
    if (!this.instances.has(key)) {
      let runtime;
      
      switch (this.platform) {
        case PlatformType.AWS_LAMBDA:
          runtime = new LambdaRuntime(config);
          break;
        
        case PlatformType.BEDROCK:
          // For bedrock, we can use the generic runtime
          // or create a specific bedrock runtime later
          runtime = new GenericRuntime(config);
          break;
        
        case PlatformType.EXPRESS:
        case PlatformType.FASTIFY:
        case PlatformType.KOA:
        case PlatformType.GENERIC:
        default:
          runtime = new GenericRuntime(config);
          break;
      }
      
      this.instances.set(key, runtime);
    }
    
    return this.instances.get(key);
  }

  /**
   * Get secret manager implementation
   * @param {Object} config - Secret manager configuration
   * @returns {SecretManagerInterface} Secret manager implementation
   */
  getSecretManager(config = {}) {
    const key = `secret_${JSON.stringify(config)}`;
    
    if (!this.instances.has(key)) {
      let secretManager;
      
      switch (this.platform) {
        case PlatformType.AWS_LAMBDA:
          secretManager = new SSMSecretManager(config);
          break;
        
        case PlatformType.BEDROCK:
          // Bedrock can use SSM if in AWS, or env variables
          if (process.env.AWS_REGION) {
            secretManager = new SSMSecretManager(config);
          } else {
            secretManager = new EnvSecretManager(config);
          }
          break;
        
        case PlatformType.EXPRESS:
        case PlatformType.FASTIFY:
        case PlatformType.KOA:
        case PlatformType.GENERIC:
        default:
          secretManager = new EnvSecretManager(config);
          break;
      }
      
      this.instances.set(key, secretManager);
    }
    
    return this.instances.get(key);
  }

  /**
   * Get logger implementation
   * @param {Object} config - Logger configuration
   * @returns {LoggerInterface} Logger implementation
   */
  getLogger(config = {}) {
    const key = `logger_${JSON.stringify(config)}`;
    
    if (!this.instances.has(key)) {
      let logger;
      
      switch (this.platform) {
        case PlatformType.AWS_LAMBDA:
          // Check if PowerTools is available
          try {
            require.resolve('@aws-lambda-powertools/logger');
            logger = new PowerToolsLogger(config);
          } catch (e) {
            // Fallback to console logger
            logger = new ConsoleLogger(config);
          }
          break;
        
        case PlatformType.BEDROCK:
        case PlatformType.EXPRESS:
        case PlatformType.FASTIFY:
        case PlatformType.KOA:
        case PlatformType.GENERIC:
        default:
          logger = new ConsoleLogger(config);
          break;
      }
      
      this.instances.set(key, logger);
    }
    
    return this.instances.get(key);
  }

  /**
   * Get configuration manager implementation
   * @param {Object} config - Configuration manager options
   * @returns {ConfigManagerInterface} Configuration manager implementation
   */
  getConfigManager(config = {}) {
    const key = `config_${JSON.stringify(config)}`;
    
    if (!this.instances.has(key)) {
      let configManager;
      
      switch (this.platform) {
        case PlatformType.AWS_LAMBDA:
          // Can use environment variables or parameter store
          configManager = new JsonConfigManager({
            ...config,
            sources: ['env', 'ssm']
          });
          break;
        
        case PlatformType.BEDROCK:
        case PlatformType.EXPRESS:
        case PlatformType.FASTIFY:
        case PlatformType.KOA:
        case PlatformType.GENERIC:
        default:
          configManager = new JsonConfigManager(config);
          break;
      }
      
      this.instances.set(key, configManager);
    }
    
    return this.instances.get(key);
  }

  /**
   * Create all platform services
   * @param {Object} config - Configuration for all services
   * @returns {Object} Object containing all platform services
   */
  createPlatformServices(config = {}) {
    return {
      runtime: this.getRuntime(config.runtime),
      secretManager: this.getSecretManager(config.secrets),
      logger: this.getLogger(config.logger),
      configManager: this.getConfigManager(config.config),
      platform: this.platform
    };
  }

  /**
   * Clear all cached instances
   */
  clearCache() {
    this.instances.clear();
  }

  /**
   * Get current platform
   * @returns {string} Current platform
   */
  getPlatform() {
    return this.platform;
  }

  /**
   * Check if running on AWS
   * @returns {boolean} True if running on AWS
   */
  isAWS() {
    return this.platform === PlatformType.AWS_LAMBDA || 
           (this.platform === PlatformType.BEDROCK && process.env.AWS_REGION);
  }

  /**
   * Check if running in serverless environment
   * @returns {boolean} True if serverless
   */
  isServerless() {
    return this.platform === PlatformType.AWS_LAMBDA || 
           this.platform === PlatformType.BEDROCK;
  }
}

// Create singleton instance
const platformFactory = new PlatformFactory();

// Export factory and types
module.exports = {
  platformFactory,
  PlatformFactory,
  PlatformType
};