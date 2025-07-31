#!/bin/bash

# Virtuoso API to AWS Lambda - Quick Start Script

echo "🚀 Virtuoso API to AWS Lambda Converter"
echo "======================================="
echo ""

# Check prerequisites
echo "📋 Checking prerequisites..."

# Check Node.js
if ! command -v node &> /dev/null; then
    echo "❌ Node.js is not installed. Please install Node.js 18+ first."
    exit 1
fi

# Check AWS CLI
if ! command -v aws &> /dev/null; then
    echo "❌ AWS CLI is not installed. Please install AWS CLI first."
    echo "   Visit: https://aws.amazon.com/cli/"
    exit 1
fi

# Check SAM CLI
if ! command -v sam &> /dev/null; then
    echo "⚠️  AWS SAM CLI is not installed (optional but recommended)."
    echo "   Visit: https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html"
    echo ""
fi

echo "✅ Prerequisites check complete!"
echo ""

# Generate Lambda functions
echo "⚡ Generating Lambda functions..."
node generate-lambdas.js

if [ $? -ne 0 ]; then
    echo "❌ Failed to generate Lambda functions"
    exit 1
fi

echo ""
echo "✅ Lambda functions generated successfully!"
echo ""
echo "📁 Generated structure:"
echo "   - lambda-functions/    (9 Lambda function handlers)"
echo "   - lambda-layer/        (Shared utilities and dependencies)"
echo "   - template.yaml        (SAM/CloudFormation template)"
echo "   - deploy.sh            (Deployment script)"
echo "   - README-LAMBDA.md     (Documentation)"
echo ""
echo "🎯 Next steps:"
echo ""
echo "1. Review the generated Lambda functions in ./lambda-functions/"
echo "2. Configure your AWS credentials:"
echo "   aws configure"
echo ""
echo "3. Deploy to AWS:"
echo "   ./deploy.sh YOUR_VIRTUOSO_API_TOKEN"
echo ""
echo "4. Optional: Deploy to specific region:"
echo "   AWS_REGION=eu-west-1 ./deploy.sh YOUR_API_TOKEN"
echo ""
echo "5. Test your deployment:"
echo "   aws lambda invoke --function-name VirtuosoProjectHandler --payload '{\"action\":\"listProjects\"}' output.json"
echo ""
echo "📚 See MEGA_PROMPT_LAMBDA_GENERATOR.md for the complete documentation"
echo "📖 See README-LAMBDA.md for usage instructions"
echo ""
echo "Happy coding! 🎉"