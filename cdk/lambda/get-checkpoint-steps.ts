import { APIGatewayProxyEventV2, APIGatewayProxyResultV2, Context } from 'aws-lambda';
import { BaseHandler, createHandler } from './shared/base-handler';
import { VirtuosoCheckpointStep } from './shared/types';

class GetCheckpointStepsHandler extends BaseHandler {
  async handle(event: APIGatewayProxyEventV2, context: Context): Promise<APIGatewayProxyResultV2> {
    return this.handleRequest(event, context, async () => {
      const authToken = this.extractAuthToken(event);
      const pathParams = this.extractPathParameters(event);
      
      if (!authToken) {
        return this.createErrorResponse(401, 'Missing authorization token');
      }

      const checkpointId = pathParams.checkpoint_id;
      if (!checkpointId) {
        return this.createErrorResponse(400, 'Missing checkpoint_id path parameter');
      }

      const steps = await this.makeVirtuosoRequest<VirtuosoCheckpointStep[]>(
        'GET',
        `/checkpoints/${checkpointId}/steps`,
        undefined,
        authToken
      );

      return this.createResponse(200, steps);
    });
  }
}

export const handler = createHandler(GetCheckpointStepsHandler);