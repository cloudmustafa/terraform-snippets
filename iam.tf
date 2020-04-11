provider "aws" {
  region = "us-east-1"

}

data "aws_iam_policy_document" "MaxPocLambdaPolicyDoc" {
  statement {
    effect = "Allow"
    actions = [
      "iam:DeleteAccessKey",
      "secretsmanager:DescribeSecret",
      "secretsmanager:PutSecretValue",
      "secretsmanager:CreateSecret",
      "iam:TagRole",
      "secretsmanager:DeleteSecret",
      "iam:CreateRole",
      "iam:PutRolePolicy",
      "iam:CreateUser",
      "iam:CreateAccessKey",
      "ecs:DeregisterTaskDefinition",
      "ecs:UpdateService",
      "iam:PassRole",
      "ecs:CreateService",
      "iam:DeleteRolePolicy",
      "ecs:RegisterTaskDefinition",
      "ecs:DescribeServices",
      "iam:ListAccessKeys",
      "iam:GetRole",
      "dynamodb:PutItem",
      "iam:DeleteUserPolicy",
      "ecs:DeleteService",
      "iam:DeleteRole",
      "iam:DeleteUser",
      "dynamodb:UpdateItem",
      "secretsmanager:UpdateSecret",
      "s3:PutObject",
      "cloudformation:CreateStack",
      "iam:PutUserPolicy",
      "cloudformation:DeleteStack",
      "iam:GetUser",
      "secretsmanager:TagResource"
    ]
    resources = ["*"]
  }
}
resource "aws_iam_policy" "MaxPolicy" {
  name   = "MaxEdgePOCDataServicePolicyDev"
  policy = "${data.aws_iam_policy_document.MaxPocLambdaPolicyDoc.json}"
}

resource "aws_iam_role" "MaxEdgeRole" {
  name = "MaxEdgePOCDataServiceLambdaRoleDev"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "MaxPOCRolePolicy" {
  role = aws_iam_role.MaxEdgeRole.name
  policy_arn = aws_iam_policy.MaxPolicy.arn
  
}

resource "aws_iam_role_policy_attachment" "MaxPOCRoleAwsPolicy" {
  role = aws_iam_role.MaxEdgeRole.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  
}

