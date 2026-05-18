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

variable "nrdot_version" {
  description = "Version of NRDOT to test during nightly"
  type        = string
}

variable "commit_sha_short" {
  description = "Short commit SHA (7 chars) for S3 artifact path"
  type        = string
}

variable "nr_ingest_key" {
  description = "New Relic ingest license key"
  type        = string
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

variable "test_key" {
  description = "Test key for scoping queries (used by action-based tests, Go tests use generated pattern)"
  type        = string
  default     = ""
}

/* 
  EC2 config options
*/ 
variable "platform" {
  type = string
  description = "EC2 platform to test on - ubuntu or windows"
  default = "ubuntu"
}

variable "platform_version" {
  type = string
  description = "Version of the EC2 platform"
  default = "24.04"
}