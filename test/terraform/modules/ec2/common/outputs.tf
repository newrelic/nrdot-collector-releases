output "private_subnet_ids" {
  description = "List of private subnet IDs"
  value       = data.aws_subnets.private_subnets.ids
}

output "security_group_id" {
  description = "Security group ID for EC2 instances"
  value       = aws_security_group.ec2_allow_all_egress.id
}

output "instance_profile_name" {
  description = "IAM instance profile name for S3 access"
  value       = data.aws_iam_instance_profile.s3_read_access.name
}
