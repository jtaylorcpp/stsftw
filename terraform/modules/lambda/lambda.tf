data "aws_caller_identity" "current" {}

data "archive_file" "init" {
  type        = "zip"
  source_file = "${var.bin_path}${var.bin_name}"
  output_path = "${var.bin_path}${var.bin_name}.zip"
}

resource "aws_iam_role" "iam_for_lambda" {
  name = "${var.app_name}_lambda_role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_policy" "lambda_logging" {
  name        = "${var.app_name}_lambda_logging"
  path        = "/"
  description = "IAM policy for logging from a lambda"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:*:*:*",
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "lambda_logs" {
  role       = aws_iam_role.iam_for_lambda.name
  policy_arn = aws_iam_policy.lambda_logging.arn
}

resource "aws_iam_policy" "lambda_vpc" {
  name        = "${var.app_name}_lambda_vpc"
  path        = "/"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:CreateNetworkInterface",
        "ec2:DescribeNetworkInterfaces",
        "ec2:DeleteNetworkInterface"
      ],
      "Resource": "*",
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "lambda_vpc" {
  role       = aws_iam_role.iam_for_lambda.name
  policy_arn = aws_iam_policy.lambda_vpc.arn
}

resource "aws_iam_policy" "lambda_dynamo_sts_sns" {
  name        = "${var.app_name}_lambda_dynamo_sts_sns"
  path        = "/"

  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "Dynamo",
            "Effect": "Allow",
            "Action": [
                "dynamodb:GetItem",
                "dynamodb:Scan",
                "dynamodb:Query"
            ],
            "Resource": "arn:aws:dynamodb:${var.region}:${data.aws_caller_identity.current.account_id}:table/${var.auth_table_name}"
        },
        {
            "Sid": "STS",
            "Effect": "Allow",
            "Action": [
              "sts:AssumeRole",
              "sts:TagSession"
            ],
            "Resource": "*"
        },
        {
            "Sid": "SNS",
            "Effect": "Allow",
            "Action": "sns:Publish",
            "Resource": "${var.sns_arn}"
        }
    ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "lambda_dynamo_sts_sns" {
  role       = aws_iam_role.iam_for_lambda.name
  policy_arn = aws_iam_policy.lambda_dynamo_sts_sns.arn
}

resource aws_security_group outbound {
  name        = "lambda_allow_outbound"
  vpc_id      = var.vpc_id

  
  egress {
      from_port        = 0
      to_port          = 0
      protocol         = "-1"
      cidr_blocks      = ["0.0.0.0/0"]
      ipv6_cidr_blocks = ["::/0"]
    }
}

resource "aws_lambda_function" "sts_app" {
  filename      = "${var.bin_path}${var.bin_name}.zip"
  function_name = var.app_name
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "${var.bin_name}"

  # The filebase64sha256() function is available in Terraform 0.11.12 and later
  # For Terraform 0.11.11 and earlier, use the base64sha256() function and the file() function:
  # source_code_hash = "${base64sha256(file("lambda_function_payload.zip"))}"
  source_code_hash = filebase64sha256("${var.bin_path}${var.bin_name}.zip")

  runtime = "go1.x"

  vpc_config {
    subnet_ids = var.lambda_subnets
    security_group_ids = [aws_security_group.outbound.id]
  }

  environment {
    variables = {
      STS_TABLE_NAME = var.auth_table_name
      STS_SNS_ARN = var.sns_arn
      STS_REGION = var.region
    }
  }
}