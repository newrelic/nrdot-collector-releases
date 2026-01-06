terraform {
  required_version = "1.9.8"
  required_providers {
    aws = {
      version = "5.81.0"
    }
    helm = {
      version = "2.17.0"
    }
    # Required to delete random resources, can be removed after successful apply
    random = {
      version = "3.7.2"
    }
  }
}

terraform {
  backend "s3" {
    encrypt        = true
    dynamodb_table = "terraform-states-lock"
    region         = "us-east-1"
    key            = "newrelic/nrdot-collector-releases/nightly/terraform.tfstate"
    # 'bucket' and 'role_arn' provided via '-backend-config'
  }
}

provider "aws" {
  region              = var.aws_region
  allowed_account_ids = [var.aws_account_id]
  # necessary role is already assumed as part of nightly workflow
}

