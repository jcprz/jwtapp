# Deployment Guide

This guide explains how to deploy the JWT App to AWS Lambda using AWS CDK and GitHub Actions.

## Architecture

The application is deployed using the following AWS services:

- **AWS Lambda** with Function URL for the API
- **Amazon RDS PostgreSQL** for the database
- **Amazon ElastiCache Redis** for session caching
- **Amazon VPC** for networking
- **AWS Secrets Manager** for secure credential storage
- **Amazon Route53** for DNS (optional subdomain setup)

## Prerequisites

1. AWS Account with appropriate permissions
2. GitHub repository with the code
3. AWS CLI installed and configured (for local deployment)
4. Go 1.23 or later
5. Node.js and npm (for CDK CLI)

## Setup Instructions

### 1. Configure AWS Account

You need an AWS account with a GitHub OIDC provider and an IAM role named `github-actions-role` with `PowerUserAccess` policy.

If you haven't set this up yet, follow these steps:

```bash
# Create OIDC provider for GitHub Actions
aws iam create-open-id-connect-provider \
  --url https://token.actions.githubusercontent.com \
  --client-id-list sts.amazonaws.com \
  --thumbprint-list 6938fd4d98bab03faadb97b34396831e3780aea1

# Create IAM role (replace YOUR_GITHUB_ORG and YOUR_REPO)
cat > trust-policy.json <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::YOUR_ACCOUNT_ID:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": "repo:YOUR_GITHUB_ORG/YOUR_REPO:*"
        }
      }
    }
  ]
}
EOF

aws iam create-role \
  --role-name github-actions-role \
  --assume-role-policy-document file://trust-policy.json

# Attach PowerUserAccess policy
aws iam attach-role-policy \
  --role-name github-actions-role \
  --policy-arn arn:aws:iam::aws:policy/PowerUserAccess
```

### 2. Configure GitHub Secrets

Add the following secrets to your GitHub repository (Settings > Secrets and variables > Actions):

- `AWS_ACCOUNT_ID`: Your AWS account ID (e.g., 123456789012)

### 3. Deploy via GitHub Actions

The deployment happens automatically when you push to the `main` branch. You can also trigger it manually:

1. Go to **Actions** tab in your GitHub repository
2. Select **Deploy to AWS Lambda** workflow
3. Click **Run workflow**

### 4. Local Deployment (Optional)

To deploy from your local machine:

```bash
# Install CDK CLI
npm install -g aws-cdk

# Configure AWS credentials
export AWS_ACCOUNT_ID=your-account-id
export AWS_REGION=us-east-1
export CDK_DEFAULT_ACCOUNT=$AWS_ACCOUNT_ID
export CDK_DEFAULT_REGION=$AWS_REGION

# Bootstrap CDK (first time only)
cd cdk
cdk bootstrap aws://$AWS_ACCOUNT_ID/$AWS_REGION

# Deploy the stack
cdk deploy --all

# Get the Function URL
aws cloudformation describe-stacks \
  --stack-name JwtAppStack \
  --query 'Stacks[0].Outputs[?OutputKey==`FunctionUrl`].OutputValue' \
  --output text
```

## Infrastructure Components

### Lambda Function

- Runtime: Go 1.23 (custom runtime)
- Memory: 512 MB
- Timeout: 30 seconds
- Function URL: Enabled with CORS support

### RDS PostgreSQL

- Engine: PostgreSQL 15
- Instance Type: db.t3.small
- Storage: 20 GB (auto-scaling up to 100 GB)
- Backup: 7-day retention

### ElastiCache Redis

- Engine: Redis 7.0
- Node Type: cache.t3.micro
- Nodes: 1 (single node)

### VPC Configuration

- 2 Availability Zones
- Public, Private, and Isolated subnets
- 1 NAT Gateway for Lambda internet access

## Environment Variables

The Lambda function is configured with the following environment variables:

- `APP_PORT`: Application port (8080)
- `DB_HOST`: RDS endpoint (auto-configured)
- `DB_PORT`: Database port (auto-configured)
- `DB_NAME`: Database name (jwtappdb)
- `DB_USER`: Database username (jwtappuser)
- `DB_DIALECT`: Database dialect (postgres)
- `REDIS_HOST`: ElastiCache endpoint (auto-configured)
- `REDIS_PORT`: Redis port (auto-configured)
- `REDIS_PASSWORD`: Redis password (empty for development)
- `DB_PASSWORD_SECRET_ARN`: ARN of the DB password secret
- `JWT_SECRET_ARN`: ARN of the JWT secret

## API Endpoints

After deployment, the Lambda Function URL will be available. The following endpoints are exposed:

- `POST /signup` - Create a new user account
- `POST /login` - Authenticate and receive a JWT token
- `DELETE /delete` - Delete user account (requires authentication)
- `GET /protected` - Protected endpoint (requires authentication)
- `GET /healthz` - Health check endpoint

## Testing the Deployment

```bash
# Get the Function URL from CloudFormation outputs
FUNCTION_URL=$(aws cloudformation describe-stacks \
  --stack-name JwtAppStack \
  --query 'Stacks[0].Outputs[?OutputKey==`FunctionUrl`].OutputValue' \
  --output text)

# Health check
curl ${FUNCTION_URL}healthz

# Sign up
curl -X POST ${FUNCTION_URL}signup \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Login
TOKEN=$(curl -X POST ${FUNCTION_URL}login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' \
  | jq -r '.token')

# Access protected endpoint
curl ${FUNCTION_URL}protected \
  -H "Authorization: Bearer $TOKEN"
```

## Route53 Custom Domain (Optional)

The CDK stack includes configuration for a custom subdomain `jwtapp.juliops.com`. To enable this:

1. Ensure you have a hosted zone for `juliops.com` in Route53
2. Update the CDK code to create an API Gateway Custom Domain or CloudFront distribution
3. Lambda Function URLs don't natively support custom domains, so you'll need to add:
   - CloudFront distribution in front of the Function URL
   - ACM certificate for HTTPS
   - Route53 record pointing to CloudFront

## Cost Estimates

Monthly costs (approximate, us-east-1):

- Lambda: $0 (within free tier for light usage)
- RDS db.t3.small: ~$30
- ElastiCache cache.t3.micro: ~$12
- NAT Gateway: ~$32
- Data transfer: Variable

**Total: ~$75/month** (can be reduced by using smaller instances)

## Cleanup

To delete all resources:

```bash
cd cdk
cdk destroy --all
```

**Warning:** This will delete the RDS database. Make sure to back up any important data first.

## Troubleshooting

### Lambda Cold Starts

Initial requests may be slow due to Lambda cold starts. Consider:
- Increasing memory allocation
- Using provisioned concurrency (additional cost)
- Implementing connection pooling

### Database Connection Issues

If Lambda can't connect to RDS:
- Check security groups allow traffic from Lambda to RDS
- Verify Lambda is in the correct VPC and subnets
- Check RDS is in isolated subnets and Lambda is in private subnets with NAT

### Secret Access Issues

If Lambda can't read secrets:
- Verify IAM role has `secretsmanager:GetSecretValue` permission
- Check secret ARNs are correctly configured in environment variables

## Security Considerations

1. **Secrets Management**: All sensitive credentials are stored in AWS Secrets Manager
2. **Network Isolation**: Database and Redis are in isolated subnets with no internet access
3. **Authentication**: JWT-based authentication for API endpoints
4. **HTTPS**: Function URLs use HTTPS by default
5. **Least Privilege**: IAM roles follow least privilege principle

## Monitoring and Logging

- Lambda logs are sent to CloudWatch Logs automatically
- Set up CloudWatch Alarms for:
  - Lambda errors
  - Lambda duration
  - RDS CPU/connections
  - ElastiCache metrics

## Next Steps

1. Set up CloudWatch Dashboards for monitoring
2. Configure CloudWatch Alarms for critical metrics
3. Implement CloudFront + Custom Domain for production
4. Set up automated database backups
5. Implement CI/CD for blue/green deployments
6. Add WAF rules for API protection
