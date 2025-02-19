locals {
  test_spec                                       = yamldecode(file("${path.module}/../../../distributions/${var.distro}/test-spec.yaml"))
  releases_bucket_name                            = "nr-releases"
  required_permissions_boundary_arn_for_new_roles = "arn:aws:iam::${var.aws_account_id}:policy/resource-provisioner-boundary"
}

resource "random_string" "deploy_id" {
  length  = 6
  special = false
}


data "aws_ecr_repository" "ecr_repo" {
  name = var.distro
}

resource "helm_release" "ci_e2e_nightly" {
  name  = "ci-e2etest-nightly-${var.distro}"
  chart = "../../charts/nr_backend"

  create_namespace = true
  namespace        = "nightly-${var.distro}"

  set {
    name  = "image.repository"
    value = data.aws_ecr_repository.ecr_repo.repository_url
  }

  set {
    name  = "image.tag"
    value = "nightly@${var.nightly_docker_manifest_sha}"
  }

  set {
    name  = "image.pullPolicy"
    value = "Always"
  }

  set {
    name  = "secrets.nrBackendUrl"
    value = var.nr_backend_url
  }

  set {
    name  = "secrets.nrIngestKey"
    value = var.nr_ingest_key
  }

  set {
    name  = "collector.hostname"
    value = "${var.test_environment}-${random_string.deploy_id.result}-${var.distro}-k8s_node"
  }
}

module "ci_e2e_ec2" {
  count                = local.test_spec.nightly.ec2.enabled ? 1 : 0
  source               = "../modules/ec2"
  releases_bucket_name = local.releases_bucket_name
  collector_distro     = var.distro
  nr_ingest_key        = var.nr_ingest_key
  # reuse vpc to avoid having to pay for second NAT gateway for this simple use case
  vpc_id              = data.aws_eks_cluster.eks_cluster.vpc_config[0].vpc_id
  deploy_id           = random_string.deploy_id.result
  permission_boundary = local.required_permissions_boundary_arn_for_new_roles
}
