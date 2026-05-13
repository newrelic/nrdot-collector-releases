variable "platform_version" {
  type = string
  description = "Version of the EC2 platform"

  validation {
    condition     = contains(["2016", "2019", "2022", "2025"], var.platform_version)
    error_message = "platform_version must be one of: '2016', '2019', '2022', '2025'"
  }
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