import { APIGatewayProxyEventV2, APIGatewayProxyResultV2, Context } from 'aws-lambda';
import { BaseHandler, createHandler } from './shared/base-handler';
import { VirtuosoExecutionAnalysis } from './shared/types';

class GetExecutionAnalysisHandler extends BaseHandler {
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

      const analysis = await this.makeVirtuosoRequest<VirtuosoExecutionAnalysis>(
        'GET',
        `/executions/${executionId}/analysis`,
        undefined,
        authToken
      );

      return this.createResponse(200, analysis);
    });
  }
}

export const handler = createHandler(GetExecutionAnalysisHandler);