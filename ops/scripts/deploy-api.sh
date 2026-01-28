#!/bin/bash
set -e

# Deploy API to ECS
# Usage: ./deploy-api.sh [tag]
# Default tag is 'latest'

TAG=${1:-latest}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
TERRAFORM_DIR="$PROJECT_ROOT/ops/terraform"

# Detect container runtime (docker or podman)
if command -v docker &> /dev/null; then
    CONTAINER_CMD="docker"
elif command -v podman &> /dev/null; then
    CONTAINER_CMD="podman"
else
    echo "❌ Error: Neither docker nor podman found"
    exit 1
fi

echo "🚀 Deploying API to ECS..."
echo "🐳 Using container runtime: $CONTAINER_CMD"

# Check if terraform directory exists
if [ ! -d "$TERRAFORM_DIR" ]; then
    echo "❌ Error: Terraform directory not found at $TERRAFORM_DIR"
    exit 1
fi

# Get ECR repository URL from Terraform
cd "$TERRAFORM_DIR"
ECR_REPO=$(terraform output -raw ecr_repository_url 2>/dev/null)
if [ -z "$ECR_REPO" ]; then
    echo "❌ Error: Could not get ECR repository URL from Terraform"
    echo "Make sure you've run 'terraform apply' first"
    exit 1
fi

ECS_CLUSTER=$(terraform output -raw ecs_cluster_name 2>/dev/null)
ECS_SERVICE=$(terraform output -raw ecs_service_name 2>/dev/null)

echo "📦 ECR Repository: $ECR_REPO"
echo "🏷️  Image Tag: $TAG"

# Extract region and account from ECR URL
AWS_REGION=$(echo "$ECR_REPO" | cut -d'.' -f4)
ECR_REGISTRY=$(echo "$ECR_REPO" | cut -d'/' -f1)

echo "🔐 Logging into ECR..."
aws ecr get-login-password --region "$AWS_REGION" | $CONTAINER_CMD login --username AWS --password-stdin "$ECR_REGISTRY"

# Build Docker image (Go cross-compiles to AMD64 for AWS Fargate)
# The Dockerfile's GOARCH=amd64 ensures the binary is x86-64 compatible
echo "🏗️  Building Docker image (cross-compiling to AMD64)..."
cd "$PROJECT_ROOT/api"
$CONTAINER_CMD build -t flyhalf-api:"$TAG" --target production .

# Tag image for ECR
echo "🏷️  Tagging image for ECR..."
$CONTAINER_CMD tag flyhalf-api:"$TAG" "$ECR_REPO":"$TAG"

# Push to ECR
echo "⬆️  Pushing image to ECR..."
$CONTAINER_CMD push "$ECR_REPO":"$TAG"

# Update ECS service
echo "🔄 Updating ECS service..."
aws ecs update-service \
    --cluster "$ECS_CLUSTER" \
    --service "$ECS_SERVICE" \
    --force-new-deployment \
    --region "$AWS_REGION" \
    --no-cli-pager

echo "✅ Deployment initiated successfully!"
echo ""
echo "To monitor the deployment:"
echo "  aws ecs describe-services --cluster $ECS_CLUSTER --services $ECS_SERVICE --region $AWS_REGION"
echo ""
echo "To view logs:"
echo "  aws logs tail /ecs/flyhalf-api --follow --region $AWS_REGION"
