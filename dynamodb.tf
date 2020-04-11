resource "aws_dynamodb_table" "dynamodb_subscriptions" {
  name="subscriptionsdev"
  billing_mode="PAY_PER_REQUEST"
  hash_key="subscriptionID"

  attribute {
      name="subscriptionID"
      type= "S"
  }
}

resource "aws_dynamodb_table" "dynamodb_subscriptions_health" {
  name="subscriptions_healthdev"
   billing_mode="PAY_PER_REQUEST"
  hash_key="subscriptionID"

  attribute {
      name="subscriptionID"
      type= "S"
  }
}

resource "aws_dynamodb_table" "dynamodb_status" {
  name="statusdev"
 billing_mode="PAY_PER_REQUEST"
  hash_key="subscriptionID"

  attribute {
      name="subscriptionID"
      type= "S"
  }
}