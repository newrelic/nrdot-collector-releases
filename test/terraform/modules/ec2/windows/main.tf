locals {
  test_key_prefix = "ec2_windows_server_${var.platform_version}"
}

module "shared" {
  source = "../shared"

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
  subnet_id = module.shared.private_subnet_ids[0]
  vpc_security_group_ids = [module.shared.security_group_id]
  iam_instance_profile = module.shared.instance_profile_name

  tags = {
    Name = "${var.test_environment}-${var.collector_distro}-${local.test_key_prefix}"
  }

  user_data_replace_on_change = true
  user_data                   = <<-EOF
              <powershell>
                # Start transcript to capture all output (Windows EC2 does not print logs to console)
                $logFile = "C:\Windows\Temp\install.log"
                $s3LogPath = "s3://${var.logs_bucket_name}/${var.collector_distro}/nightly-windows-${var.platform_version}.log"
                Start-Transcript -Path $logFile -Append

                try {
                  Write-Host "=========================================="
                  Write-Host "Transcribing logs for nightly run ${var.collector_distro}-windows-${var.platform_version} on $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
                  Write-Host "=========================================="

                  Write-Host "📋 Fetching MSI from s3"
                  Start-Process -Wait -PassThru msiexec.exe -ArgumentList '/i', 'https://awscli.amazonaws.com/AWSCLIV2.msi', '/qn'
                  $msi_package_basepath = "s3://${var.releases_bucket_name}/nrdot-collector-releases/${var.collector_distro}/${var.nrdot_version}/${var.commit_sha_short}/"
                  $latest_msi_filename = aws s3 ls $msi_package_basepath |
                    Sort-Object -Descending |
                    Where-Object { $_ -match "${var.collector_distro}" -and $_ -match "\.msi$" } |
                    Select-Object -First 1 |
                    ForEach-Object { ($_ -split '\s+')[-1] }
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
                } finally {
                  # Always stop transcript and upload log to S3, even on failure
                  Stop-Transcript
                  aws s3 cp $logFile $s3LogPath
                }
              </powershell>
              EOF
}