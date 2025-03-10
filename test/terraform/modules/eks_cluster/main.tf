locals {

  # We need to ensure we don't attempt to use any of the disallowed zones:
  # https://docs.aws.amazon.com/eks/latest/userguide/network-reqs.html
  valid_availability_zones = [
    for zone in data.aws_availability_zones.available.names : zone
    if !contains(data.aws_availability_zones.disallowed.names, zone)
  ]
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_availability_zones" "disallowed" {
  filter {
    name = "zone-id"
    values = ["use1-az3", "usw1-az2", "cac1-az3"]
  }
}


module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "5.17.0"

  name = "${var.name}-vpc"

  cidr = "10.0.0.0/16"

  azs = slice(local.valid_availability_zones, 0, 2)

  private_subnets = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  public_subnets = ["10.0.4.0/24", "10.0.5.0/24", "10.0.6.0/24"]

  enable_nat_gateway = true
  # Will be assigned to the first public subnet, in this case 10.0.4.0/24.  Private subnets will
  # route their internet traffic through it.
  # https://registry.terraform.io/modules/terraform-aws-modules/vpc/aws/latest#single-nat-gateway
  single_nat_gateway = true

  vpc_flow_log_permissions_boundary = var.permission_boundary
}


module "eks_cluster" {
  source  = "terraform-aws-modules/eks/aws"
  version = "20.31.4"

  cluster_name    = var.name
  cluster_version = "1.31"

  cluster_endpoint_private_access = true
  cluster_endpoint_public_access  = true

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  enable_cluster_creator_admin_permissions = true
  iam_role_permissions_boundary            = var.permission_boundary

  cloudwatch_log_group_retention_in_days = 1

  eks_managed_node_group_defaults = {
    iam_role_permissions_boundary = var.permission_boundary

    instance_types = ["m5.large"]

    min_size     = 2
    max_size     = 2
    desired_size = 2
  }

  eks_managed_node_groups = {
    default = {
      # Starting on 1.30, AL2023 is the default AMI type for EKS managed node groups
      ami_type = "AL2023_x86_64_STANDARD"
    }
  }
}


