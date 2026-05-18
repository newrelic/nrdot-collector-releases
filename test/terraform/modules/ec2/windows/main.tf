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
                # Proof of life - write error event immediately
                New-EventLog -LogName System -Source "E2EUserDataScript" -ErrorAction SilentlyContinue
                Write-EventLog -LogName System -Source "E2EUserDataScript" -EntryType Error -EventId 1000 -Message "UserData script started execution"

                # We need to send events to New Relic as events because Windows EC2 only "includes the last three system event log errors"
                # Define helper function for sending NR events.
                function Send-NREvent {
                    param([string]$Message)
                    $accountId = "${var.nr_account_id}"
                    $event = @{
                        eventType = "NightlyLog"
                        platform = "windows"
                        version = "${var.nrdot_version}"
                        message = $Message
                    } | ConvertTo-Json -Compress
                    newrelic events post --accountId $accountId --event $event
                }

                # Install newrelic cli (source: newrelic-cli readme doc)
                [Net.ServicePointManager]::SecurityProtocol = 'tls12, tls'
                (New-Object System.Net.WebClient).DownloadFile(
                  "https://download.newrelic.com/install/newrelic-cli/scripts/install.ps1",
                  "$env:TEMP\install.ps1")
                & $env:TEMP\install.ps1
                # Refresh PATH to include newrelic-cli
                $env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")
                # Set newrelic cli profile
                newrelic profiles add --profile test --accountId ${var.nr_account_id} --apiKey ${var.nr_api_key} --licenseKey ${var.nr_ingest_key} -r us
                newrelic profiles default --profile test

                Send-NREvent "Fetching MSI from s3"
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
                Set-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment' -Name 'OTEL_RESOURCE_ATTRIBUTES' -Value "testKey=${var.test_key}"

                Send-NREvent "Installing collector"
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
                  Send-NREvent "MSI installation failed with exit code $($process.ExitCode)"
                  if (Test-Path $log_path) {
                    Write-Host '`nInstallation Log - Errors and Warnings:'
                    Get-Content $log_path | Select-String -Pattern 'error|warning|failed|exception|fatal' -Context 2,2
                    Write-Host ''
                  }
                  exit $process.ExitCode
                }

                Send-NREvent "Waiting 30 seconds for collector to spool up"
                Start-Sleep -Seconds 30
                
                # Dump collector logs
                Write-Host "`nCollector logs from Windows Event Log:"
                Get-WinEvent -LogName Application -MaxEvents 100 -ErrorAction SilentlyContinue | Where-Object { $_.ProviderName -eq "${var.collector_distro}" } | Select-Object -ExpandProperty Message

                # Check if service is running
                $service = Get-Service -Name "${var.collector_distro}"
                if ($service -and $service.Status -eq 'Running') {
                  Send-NREvent "Service nrdot-collector is running"
                } else {
                  Send-NREvent "Service is not running"
                  exit 1
                }
              </powershell>
              EOF
}