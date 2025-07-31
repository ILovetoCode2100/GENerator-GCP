import { APIGatewayProxyEventV2, APIGatewayProxyResultV2, Context } from 'aws-lambda';
import { BaseHandler, createHandler } from './shared/base-handler';
import { VirtuosoEnvironment } from './shared/types';

interface CreateEnvironmentRequest {
  name: string;
  variables: Record<string, any>;
}

class CreateEnvironmentHandler extends BaseHandler {
  async handle(event: APIGatewayProxyEventV2, context: Context): Promise<APIGatewayProxyResultV2> {
    return this.handleRequest(event, context, async () => {
      const authToken = this.extractAuthToken(event);
      const requestBody = this.parseRequestBody<CreateEnvironmentRequest>(event);
      
      if (!authToken) {
        return this.createErrorResponse(401, 'Missing authorization token');
      }

      if (!requestBody || !requestBody.name || !requestBody.variables) {
        return this.createErrorResponse(400, 'Missing required fields: name, variables');
      }

      const environment = await this.makeVirtuosoRequest<VirtuosoEnvironment>(
        'POST',
        '/environments',
        requestBody,
        authToken
      );

      return this.createResponse(201, environment);
    });
  }
}

export const handler = createHandler(CreateEnvironmentHandler);