import { APIGatewayProxyEventV2, APIGatewayProxyResultV2, Context } from 'aws-lambda';
import { BaseHandler, createHandler } from './shared/base-handler';
import { VirtuosoGoal } from './shared/types';

class ListGoalsHandler extends BaseHandler {
  async handle(event: APIGatewayProxyEventV2, context: Context): Promise<APIGatewayProxyResultV2> {
    return this.handleRequest(event, context, async () => {
      const authToken = this.extractAuthToken(event);
      const pathParams = this.extractPathParameters(event);
      
      if (!authToken) {
        return this.createErrorResponse(401, 'Missing authorization token');
      }

      const projectId = pathParams.project_id;
      if (!projectId) {
        return this.createErrorResponse(400, 'Missing project_id path parameter');
      }

      const goals = await this.makeVirtuosoRequest<VirtuosoGoal[]>(
        'GET',
        `/projects/${projectId}/goals`,
        undefined,
        authToken
      );

      return this.createResponse(200, goals);
    });
  }
}

export const handler = createHandler(ListGoalsHandler);