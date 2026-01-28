#!/bin/bash
set -e

# Deploy web files to S3 and invalidate CloudFront cache
# Usage: ./deploy-web.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
TERRAFORM_DIR="$PROJECT_ROOT/ops/terraform"
WEB_DIR="$PROJECT_ROOT/web"

echo "🚀 Deploying web files to S3 + CloudFront..."

# Check if terraform directory exists
if [ ! -d "$TERRAFORM_DIR" ]; then
    echo "❌ Error: Terraform directory not found at $TERRAFORM_DIR"
    exit 1
fi

# Check if web directory exists
if [ ! -d "$WEB_DIR" ]; then
    echo "❌ Error: Web directory not found at $WEB_DIR"
    exit 1
fi

# Get S3 bucket and CloudFront distribution from Terraform
cd "$TERRAFORM_DIR"
S3_BUCKET=$(terraform output -raw web_bucket_name 2>/dev/null)
if [ -z "$S3_BUCKET" ]; then
    echo "❌ Error: Could not get S3 bucket name from Terraform"
    echo "Make sure you've run 'terraform apply' first"
    exit 1
fi

CF_DISTRIBUTION=$(terraform output -raw cloudfront_distribution_id 2>/dev/null)
if [ -z "$CF_DISTRIBUTION" ]; then
    echo "❌ Error: Could not get CloudFront distribution ID from Terraform"
    exit 1
fi

AWS_REGION=$(terraform output -raw aws_region 2>/dev/null || echo "us-east-1")

echo "📦 S3 Bucket: $S3_BUCKET"
echo "🌐 CloudFront Distribution: $CF_DISTRIBUTION"

# Prepare production configuration
echo "🔧 Preparing production configuration..."
cd "$WEB_DIR"

# Backup original config and use production config
if [ -f "js/config.js" ]; then
    cp js/config.js js/config.js.backup
fi
if [ -f "js/config.production.js" ]; then
    cp js/config.production.js js/config.js
    echo "✓ Using production API URL: https://api.flyhalf.app/api/v1"
else
    echo "⚠️  Warning: config.production.js not found, using existing config.js"
fi

# Sync files to S3
echo "⬆️  Syncing files to S3..."
aws s3 sync . s3://"$S3_BUCKET"/ \
    --delete \
    --region "$AWS_REGION" \
    --exclude "*.conf" \
    --exclude ".DS_Store" \
    --exclude "*.md"

# Set content types for specific files
echo "🔧 Setting content types..."
aws s3 cp s3://"$S3_BUCKET"/index.html s3://"$S3_BUCKET"/index.html \
    --content-type "text/html; charset=utf-8" \
    --metadata-directive REPLACE \
    --region "$AWS_REGION" || true

# Invalidate CloudFront cache
echo "🔄 Invalidating CloudFront cache..."
INVALIDATION_ID=$(aws cloudfront create-invalidation \
    --distribution-id "$CF_DISTRIBUTION" \
    --paths "/*" \
    --query 'Invalidation.Id' \
    --output text)

echo "✅ Deployment completed successfully!"

# Restore original config if backup exists
if [ -f "$WEB_DIR/js/config.js.backup" ]; then
    mv "$WEB_DIR/js/config.js.backup" "$WEB_DIR/js/config.js"
    echo "✓ Restored local development config"
fi

echo ""
echo "CloudFront invalidation ID: $INVALIDATION_ID"
echo ""
echo "To check invalidation status:"
echo "  aws cloudfront get-invalidation --distribution-id $CF_DISTRIBUTION --id $INVALIDATION_ID"
echo ""
echo "Your site will be available at:"
cd "$TERRAFORM_DIR"
terraform output web_url
