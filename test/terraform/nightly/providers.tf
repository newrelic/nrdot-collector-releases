terraform {
  required_version = "1.9.8"
  required_providers {
    aws = {
      version = "5.81.0"
    }
    helm = {
      version = "2.17.0"
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
  assume_role {
    role_arn = "arn:aws:iam::${var.aws_account_id}:role/resource-provisioner"
  }
  # role "arn:aws:iam::${var.aws_account_id}:role/resource-provisioner" is expected to have been assumed

  # Use profile if provided, otherwise expect AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY as env vars
#   profile = var.aws_profile

  # Only assume role if not using a profile (legacy behavior)
  # When using a profile, credentials should already be assumed via profile configuration
#   dynamic "assume_role" {
#     for_each = var.aws_profile == "" ? [1] : []
#     content {
#       role_arn = "arn:aws:iam::${var.aws_account_id}:role/resource-provisioner"
#     }
#   }
}

data "aws_eks_cluster" "eks_cluster" {
  name = "aws-ci-e2etest"
}

data "aws_eks_cluster_auth" "eks_cluster_auth" {
  name = "aws-ci-e2etest"
}

provider "helm" {
  kubernetes {
    host                   = data.aws_eks_cluster.eks_cluster.endpoint
    cluster_ca_certificate = base64decode(data.aws_eks_cluster.eks_cluster.certificate_authority[0].data)
    token                  = data.aws_eks_cluster_auth.eks_cluster_auth.token
  }
}
