import { APIGatewayProxyEventV2, APIGatewayProxyResultV2, Context } from 'aws-lambda';
import { BaseHandler, createHandler } from './shared/base-handler';
import { VirtuosoExecution } from './shared/types';

interface StartExecutionRequest {
  goalId: string;
  environment?: string;
}

class StartExecutionHandler extends BaseHandler {
  async handle(event: APIGatewayProxyEventV2, context: Context): Promise<APIGatewayProxyResultV2> {
    return this.handleRequest(event, context, async () => {
      const authToken = this.extractAuthToken(event);
      const requestBody = this.parseRequestBody<StartExecutionRequest>(event);
      
      if (!authToken) {
        return this.createErrorResponse(401, 'Missing authorization token');
      }

      if (!requestBody || !requestBody.goalId) {
        return this.createErrorResponse(400, 'Missing required field: goalId');
      }

      const execution = await this.makeVirtuosoRequest<VirtuosoExecution>(
        'POST',
        '/executions',
        requestBody,
        authToken
      );

      return this.createResponse(201, execution);
    });
  }
}

export const handler = createHandler(StartExecutionHandler);