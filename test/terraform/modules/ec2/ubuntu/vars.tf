variable "test_environment" {
  type = string
}

variable "collector_distro" {
  type = string
}

variable "releases_bucket_name" {
  type = string
}

variable "nrdot_version" {
  type = string
}

variable "commit_sha_short" {
  type = string
}

variable "nr_ingest_key" {
  type      = string
  sensitive = true
}

variable "test_key" {
  type = string
}
