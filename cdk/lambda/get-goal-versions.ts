import { APIGatewayProxyEventV2, APIGatewayProxyResultV2, Context } from 'aws-lambda';
import { BaseHandler, createHandler } from './shared/base-handler';
import { VirtuosoGoalVersion } from './shared/types';

class GetGoalVersionsHandler extends BaseHandler {
  async handle(event: APIGatewayProxyEventV2, context: Context): Promise<APIGatewayProxyResultV2> {
    return this.handleRequest(event, context, async () => {
      const authToken = this.extractAuthToken(event);
      const pathParams = this.extractPathParameters(event);
      
      if (!authToken) {
        return this.createErrorResponse(401, 'Missing authorization token');
      }

      const goalId = pathParams.goal_id;
      if (!goalId) {
        return this.createErrorResponse(400, 'Missing goal_id path parameter');
      }

      const versions = await this.makeVirtuosoRequest<VirtuosoGoalVersion[]>(
        'GET',
        `/goals/${goalId}/versions`,
        undefined,
        authToken
      );

      return this.createResponse(200, versions);
    });
  }
}

export const handler = createHandler(GetGoalVersionsHandler);