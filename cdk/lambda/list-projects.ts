import { APIGatewayProxyEventV2, APIGatewayProxyResultV2, Context } from 'aws-lambda';
import { BaseHandler, createHandler } from './shared/base-handler';
import { VirtuosoProject } from './shared/types';

class ListProjectsHandler extends BaseHandler {
  async handle(event: APIGatewayProxyEventV2, context: Context): Promise<APIGatewayProxyResultV2> {
    return this.handleRequest(event, context, async () => {
      const authToken = this.extractAuthToken(event);
      const queryParams = this.extractQueryParameters(event);
      
      if (!authToken) {
        return this.createErrorResponse(401, 'Missing authorization token');
      }

      // Build query string for organizationId if provided
      let path = '/projects';
      if (queryParams.organizationId) {
        path += `?organizationId=${encodeURIComponent(queryParams.organizationId)}`;
      }

      const projects = await this.makeVirtuosoRequest<VirtuosoProject[]>(
        'GET',
        path,
        undefined,
        authToken
      );

      return this.createResponse(200, projects);
    });
  }
}

export const handler = createHandler(ListProjectsHandler);