locals {
  instance_identifier = "ec2_windows_server_${var.platform_version}"
}

module "common_infrastructure" {
  source = "../common"

  platform = "windows"
  platform_version = "${var.platform_version}"
  vpc_id = "${var.vpc_id}"
  test_environment = "${var.test_environment}"
  collector_distro = "${var.collector_distro}"
}

data "aws_ami" "windows_ami" {
  most_recent = true

  filter {
    name = "name"
    values = ["Windows_Server-${var.platform_version}-English-Core-Base-*"]
  }

  filter {
    name = "virtualization-type"
    values  = ["hvm"]
  }

  owners = ["801119661308"] # Amazon (Windows AMI)
}

resource "aws_instance" "windows" {
  ami = data.aws_ami.windows_ami.id
  instance_type = "t3.micro"
  subnet_id = module.common_infrastructure.private_subnet_ids[0]
  vpc_security_group_ids = [module.common_infrastructure.security_group_id]
  iam_instance_profile = module.common_infrastructure.instance_profile_name

  tags = {
    Name = "${var.test_environment}-${var.collector_distro}-${local.instance_identifier}"
  }

  user_data_replace_on_change = true
  user_data                   = templatefile("${path.module}/userdata.ps1.tftpl", {
    releases_bucket_name = var.releases_bucket_name
    collector_distro     = var.collector_distro
    nrdot_version        = var.nrdot_version
    commit_sha_short     = var.commit_sha_short
    nr_ingest_key        = var.nr_ingest_key
    test_key             = var.test_key
  })
}