import { APIGatewayProxyEventV2, APIGatewayProxyResultV2, Context } from 'aws-lambda';

export interface VirtuosoApiConfig {
  virtuosoApiBaseUrl: string;
  organizationId: string;
  apiKey: string;
}

export interface LambdaHandler {
  (event: APIGatewayProxyEventV2, context: Context): Promise<APIGatewayProxyResultV2>;
}

export interface ApiResponse<T = any> {
  statusCode: number;
  body: T;
  headers?: Record<string, string>;
}

export interface ErrorResponse {
  error: string;
  message?: string;
  requestId?: string;
  timestamp?: string;
}

export interface VirtuosoUser {
  id: string;
  name: string;
  email: string;
}

export interface VirtuosoProject {
  id: string;
  name: string;
  status: string;
}

export interface VirtuosoGoal {
  id: string;
  name: string;
  lastRun?: string;
}

export interface VirtuosoGoalVersion {
  id: string;
  version: string;
  createdAt: string;
}

export interface VirtuosoJourney {
  id: string;
  name: string;
}

export interface VirtuosoCheckpoint {
  id: string;
  name: string;
}

export interface VirtuosoCheckpointStep {
  id: string;
  action: string;
  target: string;
}

export interface VirtuosoStep {
  id: string;
  action: string;
}

export interface VirtuosoExecution {
  id: string;
  status: string;
  progress?: number;
}

export interface VirtuosoExecutionAnalysis {
  passed: number;
  failed: number;
  errors: number;
}

export interface VirtuosoLibraryCheckpoint {
  id: string;
  name: string;
  stepCount: number;
}

export interface VirtuosoTestDataTable {
  id: string;
  name: string;
}

export interface VirtuosoEnvironment {
  id: string;
  name: string;
}

export interface VirtuosoExecutionJob {
  jobId: string;
  status: string;
}