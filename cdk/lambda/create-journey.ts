import { APIGatewayProxyEventV2, APIGatewayProxyResultV2, Context } from 'aws-lambda';
import { BaseHandler, createHandler } from './shared/base-handler';
import { VirtuosoJourney } from './shared/types';

interface CreateJourneyRequest {
  name: string;
  goalId: string;
}

class CreateJourneyHandler extends BaseHandler {
  async handle(event: APIGatewayProxyEventV2, context: Context): Promise<APIGatewayProxyResultV2> {
    return this.handleRequest(event, context, async () => {
      const authToken = this.extractAuthToken(event);
      const requestBody = this.parseRequestBody<CreateJourneyRequest>(event);
      
      if (!authToken) {
        return this.createErrorResponse(401, 'Missing authorization token');
      }

      if (!requestBody || !requestBody.name || !requestBody.goalId) {
        return this.createErrorResponse(400, 'Missing required fields: name, goalId');
      }

      const journey = await this.makeVirtuosoRequest<VirtuosoJourney>(
        'POST',
        '/journeys',
        requestBody,
        authToken
      );

      return this.createResponse(201, journey);
    });
  }
}

export const handler = createHandler(CreateJourneyHandler);