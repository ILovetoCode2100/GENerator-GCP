import { APIGatewayProxyEventV2, APIGatewayProxyResultV2, Context } from 'aws-lambda';
import { BaseHandler, createHandler } from './shared/base-handler';
import { VirtuosoExecutionJob } from './shared/types';

class ExecuteSnapshotHandler extends BaseHandler {
  async handle(event: APIGatewayProxyEventV2, context: Context): Promise<APIGatewayProxyResultV2> {
    return this.handleRequest(event, context, async () => {
      const authToken = this.extractAuthToken(event);
      const pathParams = this.extractPathParameters(event);
      
      if (!authToken) {
        return this.createErrorResponse(401, 'Missing authorization token');
      }

      const goalId = pathParams.goal_id;
      const snapshotId = pathParams.snapshot_id;
      
      if (!goalId || !snapshotId) {
        return this.createErrorResponse(400, 'Missing required path parameters: goal_id, snapshot_id');
      }

      const execution = await this.makeVirtuosoRequest<VirtuosoExecutionJob>(
        'POST',
        `/goals/${goalId}/snapshots/${snapshotId}/execute`,
        {},
        authToken
      );

      return this.createResponse(200, execution);
    });
  }
}

export const handler = createHandler(ExecuteSnapshotHandler);