output "instance_ip" {
  description = "Elastic IP of the app server"
  value       = aws_eip.main.public_ip
}

output "s3_bucket" {
  description = "S3 bucket name for app assets"
  value       = aws_s3_bucket.app.id
}

output "ssh_command" {
  description = "SSH command to connect to the instance"
  value       = "ssh ubuntu@${aws_eip.main.public_ip}"
}
