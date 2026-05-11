locals {
  windows_config = [
    {
      test_key_prefix = "ec2_windows_server_2025"
      server_version = "2025"
    }
  ]
}

module "shared" {
  source = "../shared"

  vpc_id = "${var.vpc_id}"
  test_environment = "${var.test_environment}"
  collector_distro = "${var.collector_distro}"
}

data "aws_ami" "windows_ami" {
  count = length(local.windows_config)
  most_recent = true

  filter {
    name = "name"
    values = ["Windows_Server-${local.windows_config[count.index].server_version}-Core-Base-*"]
  }

  filter {
    name = "virtualization-type"
    values  = ["hvm"]
  }

  owners = ["801119661308"] # Amazon (Windows AMI)
}

resource "aws_instance" "windows" {
  count = length(local.windows_config)
  ami = data.aws_ami.windows_ami[count.index].id
  instance_type = "t2.micro"
  subnet_id = module.shared.private_subnet_ids[0]
  vpc_security_group_ids = [module.shared.security_group_id]
  iam_instance_profile = module.shared.instance_profile_name

  tags = {
    Name = "${var.test_environment}-${var.collector_distro}-${local.windows_config[count.index].test_key_prefix}"
  }

  user_data_replace_on_change = true
  user_data                   = <<-EOF
              <powershell>
                Write-Host "📋 Fetching MSI from s3"
                Start-Process -Wait -PassThru msiexec.exe -ArgumentList '/i', 'https://awscli.amazonaws.com/AWSCLIV2.msi', '/qn'
                $msi_package_basepath = "s3://${var.releases_bucket_name}/nrdot-collector-releases/${var.collector_distro}/${var.nrdot_version}/${var.commit_sha_short}/"
                $latest_msi_filename = aws s3 ls $msi_package_basepath | Sort-Object -Descending | Where-Object { $_ -match "${var.collector_distro}" -and $_ -match "\.msi$" } | Select-Object -First 1 | ForEach-Object { ($_ -split '\s+')[-1] }
                $msi_path = Join-Path $env:TEMP "collector.msi"
                aws s3 cp "$msi_package_basepath$latest_msi_filename" $msi_path

                # Set nrdot config environment variables.
                Set-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment' -Name 'NEW_RELIC_LICENSE_KEY' -Value "${var.nr_ingest_key}"
                Set-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment' -Name 'OTEL_RESOURCE_ATTRIBUTES' -Value "${var.test_key}"

                Write-Host "📋 Installing collector"
                $log_path = Join-Path $env:TEMP "msi-install.log"
                $msi_args = @(
                    '/i',
                    $msi_path,
                    '/qn',
                    '/l*',
                    $log_path
                )
                $process = Start-Process -Wait -PassThru msiexec.exe -ArgumentList $msi_args

                # Validate install successful
                Write-Host '`nInstallation Log (Last 200 lines):'
                Get-Content $log_path | Select-Object -Last 200
                if ($process.ExitCode -ne 0) {
                  Write-Host "❌ MSI installation failed with exit code $($process.ExitCode)"
                  if (Test-Path $log_path) {
                    Write-Host '`nInstallation Log - Errors and Warnings:'
                    Get-Content $log_path | Select-String -Pattern 'error|warning|failed|exception|fatal' -Context 2,2
                    Write-Host ''
                  }
                  exit $process.ExitCode
                }

                Write-Host "⏳ Waiting 30 seconds for collector to spool up"
                Start-Sleep -Seconds 30
                
                # Dump collector logs
                Write-Host "`nCollector logs from Windows Event Log:"
                Get-WinEvent -LogName Application -MaxEvents 100 -ErrorAction SilentlyContinue | Where-Object { $_.ProviderName -eq "${var.collector_distro}" } | Select-Object -ExpandProperty Message

                # Check if service is running
                $service = Get-Service -Name "${var.collector_distro}"
                if ($service -and $service.Status -eq 'Running') {
                  Write-Host "✅ Service nrdot-collector is running"
                } else {
                  Write-Error "❌ Service is not running"
                }
              </powershell>
              EOF
}