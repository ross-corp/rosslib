data "aws_iam_policy_document" "ec2_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "main" {
  name               = local.name
  assume_role_policy = data.aws_iam_policy_document.ec2_assume_role.json
  tags               = local.tags
}

data "aws_iam_policy_document" "s3_app" {
  statement {
    actions = [
      "s3:GetObject",
      "s3:PutObject",
      "s3:DeleteObject",
    ]
    resources = ["${aws_s3_bucket.app.arn}/*"]
  }

  statement {
    actions   = ["s3:ListBucket"]
    resources = [aws_s3_bucket.app.arn]
  }
}

resource "aws_iam_role_policy" "s3_app" {
  name   = "s3-app"
  role   = aws_iam_role.main.id
  policy = data.aws_iam_policy_document.s3_app.json
}

resource "aws_iam_instance_profile" "main" {
  name = local.name
  role = aws_iam_role.main.name
  tags = local.tags
}
