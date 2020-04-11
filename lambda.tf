
resource "aws_lambda_function" "Provision" {
  s3_bucket     = "maxedgecloudtocloudpoc-sandbox" #verify bucket name
  s3_key        = "main.zip"                       #need to verify file name
  function_name = "Provision-dev-test"
  role          = aws_iam_role.MaxEdgeRole.arn
  handler       = "main"
  runtime       = "go1.x"
  timeout       = "15"

  environment {
    variables = {

      AWS_ACCOUNT_ID                     = "007131566380" # Need to be a variable
      CLOUDFORMATION_ECS_SERVICE_REF     = "ECSREF"
      CLOUDFORMATION_SECRET_REF          = "SECRETREF"
      CLOUDFORMATION_TASK_DEFINITION_REF = "TASKDEFREF"
      CLUSTER_NAME                       = "MaxEdge-POC-Cluster-Dev"
      CONTAINER_IMAGE                    = "007131566380.dkr.ecr.us-east-1.amazonaws.com/cloudtocloud:latest" # Need to be a variable
      ENVIRONMENT                        = "sbx"                                                              # Need to be a variable
      EXECUTION_ROLE                     = "arn:aws:iam::007131566380:role/ecsTaskExecutionRole"
      SECURITY_GROUP                     = "sg-08697827a0774faeb"     # Need to be a variable
      SUBNET_1                           = "subnet-0ac4e0869f688225c" # Need to be a variable
      SUBNET_2                           = "subnet-06f003ebf6501c183" # Need to be a variable
      SUBNET_3                           = "subnet-00534747bcc25a82b" # Need to be a variable
      SUBSCRIPTION_METADATA_TABLE        = "subscriptions"
      TASK_ROLE                          = "arn:aws:iam::007131566380:role/ecsTaskExecutionRole"
    }
  }
}

resource "aws_lambda_permission" "lambda_permission" {
  statement_id  = "AllowProvisionAPIInvoke"
  action        = "lambda:InvokeFunction"
  function_name = "Provision-dev-test"
  principal     = "apigateway.amazonaws.com"


  # The /*/*/* part allows invocation from any stage, method and resource path
  # within API Gateway REST API.
  source_arn = "${aws_api_gateway_rest_api.MaxCloud-Dev.execution_arn}/*/*/*"
}
