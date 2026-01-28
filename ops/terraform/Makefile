.PHONY: help init plan apply destroy output fmt validate

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

init: ## Initialize Terraform with backend configuration
	terraform init -backend-config=backend.hcl

plan: ## Show execution plan
	terraform plan

apply: ## Apply Terraform configuration
	terraform apply

destroy: ## Destroy all Terraform-managed resources
	@echo "⚠️  WARNING: This will destroy all resources!"
	@echo "Make sure to backup any data before proceeding."
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		terraform destroy; \
	fi

output: ## Show Terraform outputs
	terraform output

fmt: ## Format Terraform files
	terraform fmt -recursive

validate: ## Validate Terraform configuration
	terraform validate

clean-s3: ## Empty S3 web bucket (required before destroy)
	@echo "Emptying S3 bucket..."
	aws s3 rm s3://$$(terraform output -raw web_bucket_name)/ --recursive

refresh: ## Refresh Terraform state
	terraform refresh
