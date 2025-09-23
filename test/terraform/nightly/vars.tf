variable "aws_account_id" {
  type        = string
  description = "AWS account id to deploy to"
}

variable "aws_region" {
  type        = string
  description = "AWS region to deploy to"
  default     = "us-east-1"
}

variable "distro" {
  description = "Distro to test during nightly"
  type        = string
}

variable "nightly_docker_manifest_sha" {
  description = "SHA of the nightly docker manifest"
  type        = string
}

variable "nr_backend_url" {
  type        = string
  description = "NR endpoint used in test cluster"
  sensitive   = true
}

variable "nr_ingest_key" {
  type        = string
  description = "NR ingest key used in test cluster"
  sensitive   = true
}

variable "test_environment" {
  type        = string
  description = "Name of test environment to distinguish entities"
  default     = "nightly"
}

variable "fips"  {
  type        = bool
  description = "Is FIPS compliant or not"
  default     = false
}

