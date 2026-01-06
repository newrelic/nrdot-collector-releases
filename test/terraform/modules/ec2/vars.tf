variable "permission_boundary" {
  description = "ARN of the IAM policy that is used to set the permissions boundary for the IAM roles created by this module"
  type        = string
}

variable "releases_bucket_name" {
  type = string
}

variable "test_environment" {
  type        = string
  description = "Name of test environment to distinguish entities"
}

variable "vpc_id" {
  description = "The ID of the VPC where the instance will be deployed to (in one of the private subnets)"
  type        = string
}

variable "nr_ingest_key" {
  description = "New Relic ingest license key"
  type        = string
  sensitive   = true
}

variable "collector_distro" {
  description = "Name of the distribution of NRDOT to install"
  type        = string
}

variable "nrdot_version" {
  description = "Version of NRDOT to install"
  type        = string
}

variable "commit_sha_short" {
  description = "Short commit SHA (7 chars) for S3 artifact path"
  type        = string
}

variable "test_key" {
  description = "Test key for scoping queries"
  type        = string
}