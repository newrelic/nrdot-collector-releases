locals {
  fips_str                                        = var.fips ? "-fips" : ""
  test_env_name                                   = "${var.k8s_namespace_prefix}${local.fips_str}"
  test_spec                                       = yamldecode(file("${path.module}/../../../distributions/${var.distro}/test/spec-nightly.yaml"))
  ec2_enabled                                     = local.test_spec.nightly.ec2.enabled && !var.fips
  chart_name                                      = local.test_spec.nightly.collectorChart.name
  chart_version                                   = try(local.test_spec.nightly.collectorChart.version, null)
  releases_bucket_name                            = "nr-releases"
  required_permissions_boundary_arn_for_new_roles = "arn:aws:iam::${var.aws_account_id}:policy/resource-provisioner-boundary"
  k8s_namespace                                   = "${var.k8s_namespace_prefix}${local.fips_str}-${var.distro}"
}

resource "random_string" "deploy_id" {
  length  = 6
  special = false
}


data "aws_ecr_repository" "ecr_repo" {
  name = var.distro
}

resource "helm_release" "ci_e2e_nightly_nr_backend" {
  count = local.chart_name == "nr_backend" ? 1 : 0
  name  = "${local.test_env_name}-nr-backend-${var.distro}"
  chart = "../../charts/nr_backend"

  create_namespace  = true
  namespace         = local.k8s_namespace
  dependency_update = true

  set {
    name  = "image.repository"
    value = data.aws_ecr_repository.ecr_repo.repository_url
  }

  set {
    name  = "image.tag"
    value = var.image_tag
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
    name  = "testKey"
    value = "${local.test_env_name}-${random_string.deploy_id.result}-${var.distro}-k8s_node"
  }

  set {
    name  = "clusterName"
    value = data.aws_eks_cluster.eks_cluster.name
  }

  set {
    name  = "demoService.enabled"
    value = "true"
  }
}

resource "helm_release" "ci_e2e_nightly_nr_k8s_otel_collector" {
  count      = local.chart_name == "newrelic/nr-k8s-otel-collector" ? 1 : 0
  name       = "${local.test_env_name}-nr-k8s-otel-${var.distro}"
  repository = "https://helm-charts.newrelic.com"
  chart      = "nr-k8s-otel-collector"
  version    = local.chart_version

  create_namespace  = true
  namespace         = local.k8s_namespace
  dependency_update = true

  set {
    name  = "images.collector.repository"
    value = data.aws_ecr_repository.ecr_repo.repository_url
  }

  set {
    name  = "images.collector.tag"
    value = var.image_tag
  }

  set {
    name  = "nrStaging"
    value = strcontains(var.nr_backend_url, "staging") ? "true" : "false"
  }

  set {
    name  = "licenseKey"
    value = var.nr_ingest_key
  }

  set {
    # populates k8s.cluster.name attribute; fips suffix to distinguish fips/non-fips as we use this as testKey for k8s
    name  = "cluster"
    value = "${data.aws_eks_cluster.eks_cluster.name}${local.fips_str}"
  }

  set {
    name  = "lowDataMode"
    value = "false"
  }

  set {
    // avoid name conflicts for global resources like ClusterRole when fips + non-fips nightly run side by side
    name  = "fullnameOverride"
    // usage of namespace has no particular significance, it's just guaranteed to be different for fips vs non-fips
    value = local.k8s_namespace
  }
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
  deploy_id           = random_string.deploy_id.result
  permission_boundary = local.required_permissions_boundary_arn_for_new_roles
}
