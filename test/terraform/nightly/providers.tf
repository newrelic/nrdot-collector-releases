terraform {
  required_version = "~> 1.14.0"
  required_providers {
    aws = {
      version = "5.81.0"
    }
  }
}

terraform {
  backend "s3" {
    bucket       = "nrdot-terraform-states"
    encrypt      = true
    use_lockfile = true
    region       = "us-east-1"
    key          = "newrelic/nrdot-collector-releases/nightly/terraform.tfstate"
  }
}

provider "aws" {
  region              = var.aws_region
  allowed_account_ids = [var.aws_account_id]
  # necessary role is already assumed as part of nightly workflow
}
