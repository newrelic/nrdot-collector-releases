locals {
  fips_str                                        = var.fips ? "-fips" : ""
  test_env_name                                   = "${var.k8s_namespace_prefix}${local.fips_str}"
  test_spec                                       = yamldecode(file("${path.module}/../../../distributions/${var.distro}/test/spec-nightly-action.yaml"))
  ec2_enabled                                     = try(local.test_spec.terraform.deploy_ec2, false)
  releases_bucket_name                            = "nr-releases"
  required_permissions_boundary_arn_for_new_roles = "arn:aws:iam::${var.aws_account_id}:policy/resource-provisioner-boundary"
  k8s_namespace                                   = "${var.k8s_namespace_prefix}${local.fips_str}-${var.distro}"
}

data "aws_ecr_repository" "ecr_repo" {
  name = var.distro
}

module "ci_e2e_ec2" {
  count                = local.ec2_enabled ? 1 : 0
  source               = "../modules/ec2"
  test_environment     = local.test_env_name
  releases_bucket_name = local.releases_bucket_name
  collector_distro     = var.distro
  nr_ingest_key        = var.nr_ingest_key
  # reuse vpc to avoid having to pay for second NAT gateway for this simple use case
  vpc_id              = data.aws_eks_cluster.eks_cluster.vpc_config[0].vpc_id
  permission_boundary = local.required_permissions_boundary_arn_for_new_roles
  test_key            = var.test_key
}
