locals {
  ubuntu_codenames = {
    "22.04" = "jammy"
    "24.04" = "noble"
  }
  release_short_name  = local.ubuntu_codenames[var.platform_version]
  instance_identifier = "ec2_ubuntu${replace(var.platform_version, ".", "_")}-0"
}

module "common_infrastructure" {
  source = "../common"

  platform = "ubuntu"
  platform_version = "${var.platform_version}"
  vpc_id = "${var.vpc_id}"
  test_environment = "${var.test_environment}"
  collector_distro = "${var.collector_distro}"
}

data "aws_ami" "ubuntu_ami" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd*/ubuntu-${local.release_short_name}-${var.platform_version}-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

resource "aws_instance" "ubuntu" {
  ami                    = data.aws_ami.ubuntu_ami.id
  instance_type          = "t2.micro"
  subnet_id              = module.common_infrastructure.private_subnet_ids[0]
  vpc_security_group_ids = [module.common_infrastructure.security_group_id]
  iam_instance_profile   = module.common_infrastructure.instance_profile_name

  tags = {
      Name = "${var.test_environment}-${var.collector_distro}-${local.instance_identifier}"
  }

  user_data_replace_on_change = true
  user_data                   = templatefile("${path.module}/userdata.sh.tftpl", {
    releases_bucket_name = var.releases_bucket_name
    collector_distro     = var.collector_distro
    nrdot_version        = var.nrdot_version
    commit_sha_short     = var.commit_sha_short
    nr_ingest_key        = var.nr_ingest_key
    test_key             = var.test_key
  })
}