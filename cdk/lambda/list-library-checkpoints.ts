import { APIGatewayProxyEventV2, APIGatewayProxyResultV2, Context } from 'aws-lambda';
import { BaseHandler, createHandler } from './shared/base-handler';
import { VirtuosoLibraryCheckpoint } from './shared/types';

class ListLibraryCheckpointsHandler extends BaseHandler {
  async handle(event: APIGatewayProxyEventV2, context: Context): Promise<APIGatewayProxyResultV2> {
    return this.handleRequest(event, context, async () => {
      const authToken = this.extractAuthToken(event);
      
      if (!authToken) {
        return this.createErrorResponse(401, 'Missing authorization token');
      }

      const libraryCheckpoints = await this.makeVirtuosoRequest<VirtuosoLibraryCheckpoint[]>(
        'GET',
        '/library/checkpoints',
        undefined,
        authToken
      );

      return this.createResponse(200, libraryCheckpoints);
    });
  }
}

export const handler = createHandler(ListLibraryCheckpointsHandler);