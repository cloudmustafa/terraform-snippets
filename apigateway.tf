
resource "aws_api_gateway_rest_api" "MaxCloud-Dev" {
  name        = "MaxCloud-Dev"
  description = "API for the Max Cloud Data Service"
}


# Build API Gateway
resource "aws_api_gateway_resource" "MaxApiSubscription-Resource-Subscription" {
  rest_api_id = "${aws_api_gateway_rest_api.MaxCloud-Dev.id}"
  parent_id   = "${aws_api_gateway_rest_api.MaxCloud-Dev.root_resource_id}"
  path_part   = "subscription"

}
# Build /subscription resource
resource "aws_api_gateway_method" "SubscriptionMethodPost" {
  rest_api_id      = "${aws_api_gateway_rest_api.MaxCloud-Dev.id}"
  resource_id      = "${aws_api_gateway_resource.MaxApiSubscription-Resource-Subscription.id}"
  http_method      = "POST"
  authorization    = "NONE"
  api_key_required = "true"

}

# Build subscription/POST method
resource "aws_api_gateway_integration" "SubscriptionIntegration-POST" {
  rest_api_id             = aws_api_gateway_rest_api.MaxCloud-Dev.id
  resource_id             = aws_api_gateway_resource.MaxApiSubscription-Resource-Subscription.id
  http_method             = aws_api_gateway_method.SubscriptionMethodPost.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = "arn:aws:apigateway:us-east-1:lambda:path/2015-03-31/functions/${aws_lambda_function.Provision.arn}/invocations"

}

# Build subscription/POST method response
resource "aws_api_gateway_method_response" "SubscriptionMethodResponse-POST-200" {
  rest_api_id = aws_api_gateway_rest_api.MaxCloud-Dev.id
  resource_id = aws_api_gateway_resource.MaxApiSubscription-Resource-Subscription.id
  http_method = aws_api_gateway_method.SubscriptionMethodPost.http_method
  status_code = "200"

  response_models = {
    "application/json" = "Empty"
  }
}

# Build subscription/POST intergration response
resource "aws_api_gateway_integration_response" "SubscriptionIntergrationResponse-POST-200" {
  rest_api_id = aws_api_gateway_rest_api.MaxCloud-Dev.id
  resource_id = aws_api_gateway_resource.MaxApiSubscription-Resource-Subscription.id
  http_method = aws_api_gateway_method.SubscriptionMethodPost.http_method
  status_code = aws_api_gateway_method_response.SubscriptionMethodResponse-POST-200.status_code

  response_templates = {
    "application/json" = ""
  }
}

resource "aws_api_gateway_method_response" "SubscriptionMethodResponse-POST-404" {
  rest_api_id = aws_api_gateway_rest_api.MaxCloud-Dev.id
  resource_id = aws_api_gateway_resource.MaxApiSubscription-Resource-Subscription.id
  http_method = aws_api_gateway_method.SubscriptionMethodPost.http_method
  status_code = "404"

  response_models = {
    "application/json" = "Empty"
  }
}

resource "aws_api_gateway_resource" "MaxApiSubscription-Resource-Subscription-ID" {
  rest_api_id = "${aws_api_gateway_rest_api.MaxCloud-Dev.id}"
  parent_id   = "${aws_api_gateway_resource.MaxApiSubscription-Resource-Subscription.id}"
  path_part   = "{subscriptionid}"

}

resource "aws_api_gateway_method" "SubscriptionID-PATCH" {
  rest_api_id      = "${aws_api_gateway_rest_api.MaxCloud-Dev.id}"
  resource_id      = "${aws_api_gateway_resource.MaxApiSubscription-Resource-Subscription-ID.id}"
  http_method      = "PATCH"
  authorization    = "NONE"
  api_key_required = "true"

}
resource "aws_api_gateway_method" "SubscriptionID-PUT" {
  rest_api_id      = "${aws_api_gateway_rest_api.MaxCloud-Dev.id}"
  resource_id      = "${aws_api_gateway_resource.MaxApiSubscription-Resource-Subscription-ID.id}"
  http_method      = "PUT"
  authorization    = "NONE"
  api_key_required = "true"

}

resource "aws_api_gateway_integration" "SubscriptionIntegration-PATCH" {
  rest_api_id             = "${aws_api_gateway_rest_api.MaxCloud-Dev.id}"
  resource_id             = "${aws_api_gateway_resource.MaxApiSubscription-Resource-Subscription-ID.id}"
  http_method             = "${aws_api_gateway_method.SubscriptionID-PATCH.http_method}"
  integration_http_method = "PATCH"
  type                    = "AWS_PROXY"
  uri                     = "arn:aws:apigateway:us-east-1:lambda:path/2015-03-31/functions/${aws_lambda_function.Provision.arn}/invocations"
  #timeout is default to 29,000 milliseconds

}

resource "aws_api_gateway_integration" "SubscriptionIntegration-PUT" {
  rest_api_id             = "${aws_api_gateway_rest_api.MaxCloud-Dev.id}"
  resource_id             = "${aws_api_gateway_resource.MaxApiSubscription-Resource-Subscription-ID.id}"
  http_method             = "${aws_api_gateway_method.SubscriptionID-PUT.http_method}"
  integration_http_method = "PUT"
  type                    = "AWS_PROXY"
  uri                     = "arn:aws:apigateway:us-east-1:lambda:path/2015-03-31/functions/${aws_lambda_function.Provision.arn}/invocations"
  #timeout is default to 29,000 milliseconds

}

resource "aws_api_gateway_method_response" "SubscriptionIDMethodResponse-PATCH-200" {
  rest_api_id = aws_api_gateway_rest_api.MaxCloud-Dev.id
  resource_id = aws_api_gateway_resource.MaxApiSubscription-Resource-Subscription-ID.id
  http_method = aws_api_gateway_method.SubscriptionID-PATCH.http_method
  status_code = "200"

  response_models = {
    "application/json" = "Empty"
  }
}


resource "aws_api_gateway_integration_response" "SubscriptionIDIntergrationResponse-PATCH-200" {
  rest_api_id = aws_api_gateway_rest_api.MaxCloud-Dev.id
  resource_id = aws_api_gateway_resource.MaxApiSubscription-Resource-Subscription-ID.id
  http_method = aws_api_gateway_method.SubscriptionID-PATCH.http_method
  status_code = aws_api_gateway_method_response.SubscriptionIDMethodResponse-PATCH-200.status_code

  response_templates = {
    "application/json" = ""
  }
}

resource "aws_api_gateway_method_response" "SubscriptionIDMethodResponse-PUT-200" {
  rest_api_id = aws_api_gateway_rest_api.MaxCloud-Dev.id
  resource_id = aws_api_gateway_resource.MaxApiSubscription-Resource-Subscription-ID.id
  http_method = aws_api_gateway_method.SubscriptionID-PUT.http_method
  status_code = "200"

  response_models = {
    "application/json" = "Empty"
  }
}


resource "aws_api_gateway_integration_response" "SubscriptionIDIntergrationResponse-PUT-200" {
  rest_api_id = aws_api_gateway_rest_api.MaxCloud-Dev.id
  resource_id = aws_api_gateway_resource.MaxApiSubscription-Resource-Subscription-ID.id
  http_method = aws_api_gateway_method.SubscriptionID-PUT.http_method
  status_code = aws_api_gateway_method_response.SubscriptionIDMethodResponse-PUT-200.status_code

  response_templates = {
    "application/json" = ""
  }
}

resource "aws_api_gateway_deployment" "MaxCloudStage-dev" {
  rest_api_id = aws_api_gateway_rest_api.MaxCloud-Dev.id
  stage_name  = "Development"
  depends_on = [
    "aws_api_gateway_integration.SubscriptionIntegration-POST",
    "aws_api_gateway_integration.SubscriptionIntegration-PUT",
    "aws_api_gateway_integration.SubscriptionIntegration-PATCH",
  ]
}

resource "aws_api_gateway_usage_plan" "MaxCloudUsagePlan-dev" {
  name = "DevelopmentUsagePlan"

  api_stages {
    api_id = aws_api_gateway_rest_api.MaxCloud-Dev.id
    stage  = aws_api_gateway_deployment.MaxCloudStage-dev.stage_name
  }
}

resource "aws_api_gateway_api_key" "DevKey" {
  name = "temp"
}

resource "aws_api_gateway_usage_plan_key" "DevUsageKey" {
  key_id        = "${aws_api_gateway_api_key.DevKey.id}"
  key_type      = "API_KEY"
  usage_plan_id = "${aws_api_gateway_usage_plan.MaxCloudUsagePlan-dev.id}"
}