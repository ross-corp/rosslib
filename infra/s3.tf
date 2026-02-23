resource "aws_s3_bucket" "app" {
  bucket = "${local.name}-app"
  tags   = local.tags
}

resource "aws_s3_bucket_public_access_block" "app" {
  bucket                  = aws_s3_bucket.app.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_server_side_encryption_configuration" "app" {
  bucket = aws_s3_bucket.app.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# Uncomment if you want a separate bucket for Terraform state.
# Create this first (manually or via a bootstrap script) before
# enabling the S3 backend in main.tf.
#
# resource "aws_s3_bucket" "state" {
#   bucket = "${local.name}-terraform-state"
#   tags   = local.tags
# }
#
# resource "aws_s3_bucket_versioning" "state" {
#   bucket = aws_s3_bucket.state.id
#   versioning_configuration { status = "Enabled" }
# }
