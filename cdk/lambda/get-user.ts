import { APIGatewayProxyEventV2, APIGatewayProxyResultV2, Context } from 'aws-lambda';
import { BaseHandler, createHandler } from './shared/base-handler';
import { VirtuosoUser } from './shared/types';

class GetUserHandler extends BaseHandler {
  async handle(event: APIGatewayProxyEventV2, context: Context): Promise<APIGatewayProxyResultV2> {
    return this.handleRequest(event, context, async () => {
      const authToken = this.extractAuthToken(event);
      
      if (!authToken) {
        return this.createErrorResponse(401, 'Missing authorization token');
      }

      const userData = await this.makeVirtuosoRequest<VirtuosoUser>(
        'GET',
        '/user',
        undefined,
        authToken
      );

      return this.createResponse(200, userData);
    });
  }
}

export const handler = createHandler(GetUserHandler);