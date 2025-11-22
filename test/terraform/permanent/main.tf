locals {
  distros = toset(distinct(flatten([
    for _, v in fileset(path.module, "../../../distributions/*/**") :
    regex("../../../distributions/([^/]*).*", dirname(v))
    if !strcontains(v, "nrdot-collector-experimental")
  ]))) 
  releases_bucket_name                            = "nr-releases"
  required_permissions_boundary_arn_for_new_roles = "arn:aws:iam::${var.aws_account_id}:policy/resource-provisioner-boundary"
}

data "aws_vpc" "this" {
  id = "vpc-015d2f927c8b5dea7"
}


module "ci_e2e_cluster" {
  source = "../modules/eks_cluster"

  name                = "aws-ci-e2etest"
  permission_boundary = local.required_permissions_boundary_arn_for_new_roles
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

module "ecr" {
  depends_on = [module.ci_e2e_cluster]

  for_each = local.distros

  source          = "terraform-aws-modules/ecr/aws"
  version         = "2.3.1"
  repository_name = each.value

  repository_image_tag_mutability = "MUTABLE"

  repository_read_write_access_arns = [data.aws_iam_session_context.current.issuer_arn]
  repository_read_access_arns = [module.ci_e2e_cluster.cluster_iam_role_arn]

  repository_lifecycle_policy = jsonencode({
    rules = [
      {
        rulePriority = 1,
        description  = "Keep last 10 nightly images",
        selection = {
          tagStatus   = "tagged",
          tagPrefixList = ["nightly"],
          countType   = "imageCountMoreThan",
          countNumber = 10
        },
        action = {
          type = "expire"
        }
      }
    ]
  })
}

module "s3_bucket" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = "4.5.0"
  bucket  = local.releases_bucket_name
  acl     = "private"

  control_object_ownership = true
  object_ownership         = "ObjectWriter"

  versioning = {
    enabled = true
  }
}

data "aws_eks_cluster_auth" "this" {
  name = module.ci_e2e_cluster.cluster_name
}