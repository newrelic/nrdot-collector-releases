locals {
  fips_str             = var.fips ? "-fips" : ""
  test_env_name        = "${var.test_env_prefix}${local.fips_str}"
  releases_bucket_name = "nrdot-test-artifacts"
}

module "ci_e2e_ec2_ubuntu" {
  count                = var.platform == "ubuntu" ? 1 : 0
  source               = "../modules/ec2/ubuntu"
  test_environment     = local.test_env_name
  releases_bucket_name = local.releases_bucket_name
  collector_distro     = var.distro
  nr_ingest_key        = var.nr_ingest_key
  vpc_id              = var.vpc_id
  subnet_id           = var.subnet_id
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
  vpc_id              = var.vpc_id
  subnet_id           = var.subnet_id
  test_key            = var.test_key
  nrdot_version       = var.nrdot_version
  commit_sha_short    = var.commit_sha_short
  platform_version    = var.platform_version
}