const { SSMClient, GetParameterCommand } = require('@aws-sdk/client-ssm');

const ssm = new SSMClient();

exports.getApiToken = async () => {
  const command = new GetParameterCommand({
    Name: process.env.API_TOKEN_PARAM || '/virtuoso/api-token',
    WithDecryption: true
  });
  const response = await ssm.send(command);
  return response.Parameter.Value;
};