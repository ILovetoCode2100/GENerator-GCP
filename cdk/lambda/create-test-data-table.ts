import { APIGatewayProxyEventV2, APIGatewayProxyResultV2, Context } from 'aws-lambda';
import { BaseHandler, createHandler } from './shared/base-handler';
import { VirtuosoTestDataTable } from './shared/types';

interface CreateTestDataTableRequest {
  name: string;
  columns: any[];
}

class CreateTestDataTableHandler extends BaseHandler {
  async handle(event: APIGatewayProxyEventV2, context: Context): Promise<APIGatewayProxyResultV2> {
    return this.handleRequest(event, context, async () => {
      const authToken = this.extractAuthToken(event);
      const requestBody = this.parseRequestBody<CreateTestDataTableRequest>(event);
      
      if (!authToken) {
        return this.createErrorResponse(401, 'Missing authorization token');
      }

      if (!requestBody || !requestBody.name || !requestBody.columns) {
        return this.createErrorResponse(400, 'Missing required fields: name, columns');
      }

      const testDataTable = await this.makeVirtuosoRequest<VirtuosoTestDataTable>(
        'POST',
        '/testdata/tables',
        requestBody,
        authToken
      );

      return this.createResponse(201, testDataTable);
    });
  }
}

export const handler = createHandler(CreateTestDataTableHandler);