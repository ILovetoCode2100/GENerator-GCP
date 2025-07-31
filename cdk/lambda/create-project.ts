import { APIGatewayProxyEventV2, APIGatewayProxyResultV2, Context } from 'aws-lambda';
import { BaseHandler, createHandler } from './shared/base-handler';
import { VirtuosoProject } from './shared/types';

interface CreateProjectRequest {
  name: string;
  organizationId: string;
}

class CreateProjectHandler extends BaseHandler {
  async handle(event: APIGatewayProxyEventV2, context: Context): Promise<APIGatewayProxyResultV2> {
    return this.handleRequest(event, context, async () => {
      const authToken = this.extractAuthToken(event);
      const requestBody = this.parseRequestBody<CreateProjectRequest>(event);
      
      if (!authToken) {
        return this.createErrorResponse(401, 'Missing authorization token');
      }

      if (!requestBody || !requestBody.name || !requestBody.organizationId) {
        return this.createErrorResponse(400, 'Missing required fields: name, organizationId');
      }

      const project = await this.makeVirtuosoRequest<VirtuosoProject>(
        'POST',
        '/projects',
        requestBody,
        authToken
      );

      return this.createResponse(201, project);
    });
  }
}

export const handler = createHandler(CreateProjectHandler);