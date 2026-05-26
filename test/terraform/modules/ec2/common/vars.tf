variable "vpc_id" {
  description = "The ID of the VPC where the instance will be deployed to (in one of the private subnets)"
  type        = string
}

variable "test_environment" {
  type        = string
  description = "Name of test environment to distinguish entities"
}

variable "collector_distro" {
  description = "Name of the distribution of NRDOT to install"
  type        = string
}

variable "platform" {
  type = string
  description = "EC2 platform to test on - ubuntu or windows"
}

variable "platform_version" {
  type = string
  description = "Version of the EC2 platform"
}