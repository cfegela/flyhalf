terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    # Configuration should be provided via backend config file or CLI flags
    # terraform init -backend-config="bucket=YOUR_BUCKET" -backend-config="key=flyhalf/terraform.tfstate" -backend-config="region=us-east-1"
    # Or create a backend.hcl file with these values
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = var.project_name
      Environment = var.environment
      ManagedBy   = "Terraform"
    }
  }
}
