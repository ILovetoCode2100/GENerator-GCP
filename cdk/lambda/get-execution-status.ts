import { APIGatewayProxyEventV2, APIGatewayProxyResultV2, Context } from 'aws-lambda';
import { BaseHandler, createHandler } from './shared/base-handler';
import { VirtuosoExecution } from './shared/types';

class GetExecutionStatusHandler extends BaseHandler {
  async handle(event: APIGatewayProxyEventV2, context: Context): Promise<APIGatewayProxyResultV2> {
    return this.handleRequest(event, context, async () => {
      const authToken = this.extractAuthToken(event);
      const pathParams = this.extractPathParameters(event);
      
      if (!authToken) {
        return this.createErrorResponse(401, 'Missing authorization token');
      }

      const executionId = pathParams.execution_id;
      if (!executionId) {
        return this.createErrorResponse(400, 'Missing execution_id path parameter');
      }

      const execution = await this.makeVirtuosoRequest<VirtuosoExecution>(
        'GET',
        `/executions/${executionId}`,
        undefined,
        authToken
      );

      return this.createResponse(200, execution);
    });
  }
}

export const handler = createHandler(GetExecutionStatusHandler);