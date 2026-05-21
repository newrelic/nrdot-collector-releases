locals {
  fips_str             = var.fips ? "-fips" : ""
  test_env_name        = "${var.test_env_prefix}${local.fips_str}"
  releases_bucket_name = "nr-releases"
}

data "aws_eks_cluster" "eks_cluster" {
  name = "aws-ci-e2etest"
}

module "ci_e2e_ec2_ubuntu" {
  count                = var.platform == "ubuntu" ? 1 : 0
  source               = "../modules/ec2/ubuntu"
  test_environment     = local.test_env_name
  releases_bucket_name = local.releases_bucket_name
  collector_distro     = var.distro
  nr_ingest_key        = var.nr_ingest_key
  # reuse vpc to avoid having to pay for second NAT gateway for this simple use case
  vpc_id              = data.aws_eks_cluster.eks_cluster.vpc_config[0].vpc_id
  test_key            = var.test_key
  nrdot_version       = var.nrdot_version
  commit_sha_short    = var.commit_sha_short
  platform_version    = var.platform_version
}

module "ci_e2e_ec2_windows" {
  count                = var.platform == "windows" ? 1 : 0
  source               = "../modules/ec2/windows"
  test_environment     = local.test_env_name
  releases_bucket_name = local.releases_bucket_name
  collector_distro     = var.distro
  nr_ingest_key        = var.nr_ingest_key
  # reuse vpc to avoid having to pay for second NAT gateway for this simple use case
  vpc_id              = data.aws_eks_cluster.eks_cluster.vpc_config[0].vpc_id
  test_key            = var.test_key
  nrdot_version       = var.nrdot_version
  commit_sha_short    = var.commit_sha_short
  platform_version    = var.platform_version
}