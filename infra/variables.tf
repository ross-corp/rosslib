variable "region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "app_name" {
  description = "Used to name all resources"
  type        = string
  default     = "rosslib"
}

variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t3.medium"
}

variable "ssh_public_key" {
  description = "SSH public key content for EC2 access (e.g. contents of ~/.ssh/id_ed25519.pub)"
  type        = string
}

variable "domain" {
  description = "Domain name (e.g. rosslib.com) — leave empty to skip Route 53"
  type        = string
  default     = ""
}
