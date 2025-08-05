locals {
  test_spec                                       = yamldecode(file("${path.module}/../../../distributions/${var.distro}/spec-nightly.yaml"))
  ec2_enabled                                     = local.test_spec.nightly.ec2.enabled
  chart_name                                      = local.test_spec.nightly.collectorChart.name
  chart_version                                   = local.test_spec.nightly.collectorChart.version
  releases_bucket_name                            = "nr-releases"
  required_permissions_boundary_arn_for_new_roles = "arn:aws:iam::${var.aws_account_id}:policy/resource-provisioner-boundary"
  k8s_namespace                                   = "nightly-${var.distro}"
}

resource "random_string" "deploy_id" {
  length  = 6
  special = false
}


data "aws_ecr_repository" "ecr_repo" {
  name = var.distro
}

resource "helm_release" "ci_e2e_nightly_nr_backend" {
  count   = local.chart_name == "nr_backend" ? 1 : 0
  name    = "nightly-nr-backend-${var.distro}"
  chart   = "../../charts/nr_backend"
  version = local.chart_version

  create_namespace  = true
  namespace         = local.k8s_namespace
  dependency_update = true

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
  name       = "nightly-nr-k8s-otel-${var.distro}"
  repository = "https://helm-charts.newrelic.com"
  chart      = "nr-k8s-otel-collector"
  version    = local.chart_version

  create_namespace  = true
  namespace         = local.k8s_namespace
  dependency_update = true

  set {
    name  = "image.repository"
    value = data.aws_ecr_repository.ecr_repo.repository_url
  }

  set {
    name  = "image.tag"
    value = "nightly@${var.nightly_docker_manifest_sha}"
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
    name  = "cluster"
    value = data.aws_eks_cluster.eks_cluster.name
  }

  set {
    name  = "lowDataMode"
    value = "false"
  }
}

module "ci_e2e_ec2" {
  count                = local.ec2_enabled ? 1 : 0
  source               = "../modules/ec2"
  releases_bucket_name = local.releases_bucket_name
  collector_distro     = var.distro
  nr_ingest_key        = var.nr_ingest_key
  # reuse vpc to avoid having to pay for second NAT gateway for this simple use case
  vpc_id              = data.aws_eks_cluster.eks_cluster.vpc_config[0].vpc_id
  deploy_id           = random_string.deploy_id.result
  permission_boundary = local.required_permissions_boundary_arn_for_new_roles
}
