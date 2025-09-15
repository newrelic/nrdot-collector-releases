locals {
  collector_reported_hostname_prefix = "${var.test_environment}-${var.deploy_id}-${var.collector_distro}"
  instance_config = [
    {
      hostname_suffix    = "ec2_ubuntu22_04-0"
      release_verion     = "22.04"
      release_short_name = "jammy"
    },
    {
      hostname_suffix    = "ec2_ubuntu24_04-0"
      release_verion     = "24.04"
      release_short_name = "noble"
    },
  ]
  collector_version_specifier = "${var.collector_version == "" ? "" : "="}${var.collector_version}"
}

data "aws_ami" "ubuntu_ami" {
  count       = length(local.instance_config)
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd*/ubuntu-${local.instance_config[count.index].release_short_name}-${local.instance_config[count.index].release_verion}-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

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

resource "aws_security_group" "ec2_allow_all_egress" {
  name        = "${var.test_environment}-${var.collector_distro}-ec2-all-egress"
  description = "Allow all outbound traffic"
  vpc_id      = data.aws_vpc.ec2_vpc.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_iam_role" "ec2_role" {
  name = "${var.test_environment}-${var.collector_distro}-ec2-s3-read-access-role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })
  permissions_boundary = var.permission_boundary
}


resource "aws_iam_policy" "s3_nr_releases_read_policy" {
  name = "${var.test_environment}-${var.collector_distro}-s3-read-policy"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:ListBucket"
        ]
        Resource = [
          "arn:aws:s3:::${var.releases_bucket_name}",
          "arn:aws:s3:::${var.releases_bucket_name}/*",
        ]
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "s3_read_policy" {
  role       = aws_iam_role.ec2_role.name
  policy_arn = aws_iam_policy.s3_nr_releases_read_policy.arn
}

resource "aws_iam_instance_profile" "s3_read_access" {
  name = "${var.test_environment}-${var.collector_distro}-ec2-s3-nr-releases-read-access"
  role = aws_iam_role.ec2_role.name
}

resource "aws_instance" "ubuntu" {
  count                  = length(local.instance_config)
  ami                    = data.aws_ami.ubuntu_ami[count.index].id
  instance_type          = "t2.micro"
  subnet_id              = data.aws_subnets.private_subnets.ids[0]
  vpc_security_group_ids = [aws_security_group.ec2_allow_all_egress.id]
  iam_instance_profile   = aws_iam_instance_profile.s3_read_access.name

  tags = {
      Name = "${var.test_environment}-${var.collector_distro}-${local.instance_config[count.index].hostname_suffix}"
  }
  
  user_data_replace_on_change = true
  user_data                   = <<-EOF
              #!/bin/bash
              ################################################
              echo 'Installing Collector'
              export DEBIAN_FRONTEND=noninteractive
              apt update
              apt install unzip
              curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
              unzip -q awscliv2.zip
              ./aws/install
              deb_package_basepath='s3://${var.releases_bucket_name}/nrdot-collector-releases/${var.collector_distro}/${var.collector_version}/'
              latest_deb_package_filename=$(aws s3 ls $${deb_package_basepath} | sort -r | grep '${var.collector_distro}' | grep 'amd64.deb$' | head -n1 | awk '{print $NF}')
              echo "Installing collector from: $${deb_package_basepath}$${latest_deb_package_filename}"
              aws s3 cp "$${deb_package_basepath}$${latest_deb_package_filename}" /tmp/collector.deb
              dpkg -i /tmp/collector.deb
              ################################################
              echo 'Configuring Collector'
              echo 'NEW_RELIC_LICENSE_KEY=${var.nr_ingest_key}' >> /etc/${var.collector_distro}/${var.collector_distro}.conf
              echo "OTEL_RESOURCE_ATTRIBUTES='host.name=${local.collector_reported_hostname_prefix}-${local.instance_config[count.index].hostname_suffix}'" >> /etc/${var.collector_distro}/${var.collector_distro}.conf
              systemctl reload-or-restart ${var.collector_distro}.service
              EOF
}
