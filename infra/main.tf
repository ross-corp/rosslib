terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  # After first apply, create the state bucket manually (or via the s3.tf state bucket),
  # then uncomment this block and run `terraform init` to migrate local state to S3.
  # backend "s3" {
  #   bucket = "rosslib-terraform-state"
  #   key    = "infra/terraform.tfstate"
  #   region = "us-east-1"
  # }
}

provider "aws" {
  region = var.region
}

locals {
  name = var.app_name
  tags = {
    Project     = var.app_name
    ManagedBy   = "terraform"
  }
}
