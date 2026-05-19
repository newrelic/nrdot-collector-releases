# Module for infrastructure common to all EC2s

data "aws_vpc" "ec2_vpc" {
  id = var.vpc_id
}

data "aws_subnets" "private_subnets" {
  filter {
    name   = "vpc-id"
    values = [var.vpc_id]
  }
  filter {
    name   = "tag:Name"
    values = ["*private*"]
  }
}

# Shared IAM resources pre-created by bootstrap script — not managed by Terraform.
data "aws_iam_instance_profile" "s3_read_access" {
  name = "nrdot-ec2-s3-nr-releases-read-access"
}

resource "aws_security_group" "ec2_allow_all_egress" {
  name        = "${var.test_environment}-${var.collector_distro}-${var.platform}-${var.platform_version}-ec2-all-egress"
  description = "Allow all outbound traffic"
  vpc_id      = data.aws_vpc.ec2_vpc.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}