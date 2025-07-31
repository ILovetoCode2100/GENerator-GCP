import { APIGatewayProxyEventV2, APIGatewayProxyResultV2, Context } from 'aws-lambda';
import { BaseHandler, createHandler } from './shared/base-handler';
import { VirtuosoCheckpoint } from './shared/types';

interface CreateCheckpointRequest {
  name: string;
  journeyId: string;
}

class CreateCheckpointHandler extends BaseHandler {
  async handle(event: APIGatewayProxyEventV2, context: Context): Promise<APIGatewayProxyResultV2> {
    return this.handleRequest(event, context, async () => {
      const authToken = this.extractAuthToken(event);
      const requestBody = this.parseRequestBody<CreateCheckpointRequest>(event);
      
      if (!authToken) {
        return this.createErrorResponse(401, 'Missing authorization token');
      }

      if (!requestBody || !requestBody.name || !requestBody.journeyId) {
        return this.createErrorResponse(400, 'Missing required fields: name, journeyId');
      }

      const checkpoint = await this.makeVirtuosoRequest<VirtuosoCheckpoint>(
        'POST',
        '/checkpoints',
        requestBody,
        authToken
      );

      return this.createResponse(201, checkpoint);
    });
  }
}

export const handler = createHandler(CreateCheckpointHandler);