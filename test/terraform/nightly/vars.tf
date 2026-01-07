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

variable "test_env_prefix" {
  type        = string
  description = "Prefix of test environment to distinguish entities"
}

variable "fips"  {
  type        = bool
  description = "Is FIPS compliant or not"
  default     = false
}

variable "image_tag" {
  description = "Tag of the nightly docker image"
  type        = string
}

variable "test_key" {
  description = "Test key for scoping queries (used by action-based tests, Go tests use generated pattern)"
  type        = string
  default     = ""
}
