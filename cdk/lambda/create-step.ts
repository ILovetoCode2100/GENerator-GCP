import { APIGatewayProxyEventV2, APIGatewayProxyResultV2, Context } from 'aws-lambda';
import { BaseHandler, createHandler } from './shared/base-handler';
import { VirtuosoStep } from './shared/types';

interface CreateStepRequest {
  action: string;
  target: string;
  value?: string;
  checkpointId: string;
}

class CreateStepHandler extends BaseHandler {
  async handle(event: APIGatewayProxyEventV2, context: Context): Promise<APIGatewayProxyResultV2> {
    return this.handleRequest(event, context, async () => {
      const authToken = this.extractAuthToken(event);
      const requestBody = this.parseRequestBody<CreateStepRequest>(event);
      
      if (!authToken) {
        return this.createErrorResponse(401, 'Missing authorization token');
      }

      if (!requestBody || !requestBody.action || !requestBody.target || !requestBody.checkpointId) {
        return this.createErrorResponse(400, 'Missing required fields: action, target, checkpointId');
      }

      const step = await this.makeVirtuosoRequest<VirtuosoStep>(
        'POST',
        '/steps',
        requestBody,
        authToken
      );

      return this.createResponse(201, step);
    });
  }
}

export const handler = createHandler(CreateStepHandler);