/**
 * Core API Handlers Index
 * Exports all platform-agnostic Virtuoso API handlers
 */

const { BaseHandler } = require('./base-handler');
const { ProjectHandler } = require('./project-handler');
const { GoalHandler } = require('./goal-handler');
const { JourneyHandler } = require('./journey-handler');
const { CheckpointHandler } = require('./checkpoint-handler');
const { StepHandler } = require('./step-handler');
const { ExecutionHandler } = require('./execution-handler');
const { LibraryHandler } = require('./library-handler');
const { DataHandler } = require('./data-handler');
const { EnvironmentHandler } = require('./environment-handler');

module.exports = {
  BaseHandler,
  ProjectHandler,
  GoalHandler,
  JourneyHandler,
  CheckpointHandler,
  StepHandler,
  ExecutionHandler,
  LibraryHandler,
  DataHandler,
  EnvironmentHandler
};