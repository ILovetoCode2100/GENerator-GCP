import { APIGatewayProxyEventV2, APIGatewayProxyResultV2, Context } from 'aws-lambda';
import { BaseHandler, createHandler } from './shared/base-handler';
import { VirtuosoLibraryCheckpoint } from './shared/types';

interface CreateLibraryCheckpointRequest {
  name: string;
  steps: any[];
}

class CreateLibraryCheckpointHandler extends BaseHandler {
  async handle(event: APIGatewayProxyEventV2, context: Context): Promise<APIGatewayProxyResultV2> {
    return this.handleRequest(event, context, async () => {
      const authToken = this.extractAuthToken(event);
      const requestBody = this.parseRequestBody<CreateLibraryCheckpointRequest>(event);
      
      if (!authToken) {
        return this.createErrorResponse(401, 'Missing authorization token');
      }

      if (!requestBody || !requestBody.name || !requestBody.steps) {
        return this.createErrorResponse(400, 'Missing required fields: name, steps');
      }

      const libraryCheckpoint = await this.makeVirtuosoRequest<VirtuosoLibraryCheckpoint>(
        'POST',
        '/library/checkpoints',
        requestBody,
        authToken
      );

      return this.createResponse(201, libraryCheckpoint);
    });
  }
}

export const handler = createHandler(CreateLibraryCheckpointHandler);