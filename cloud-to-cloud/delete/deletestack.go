package main

import (
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"

	"fmt"
)

// InputParameters : CloudFormation stack ID passed in that
// will delete the CloudFormation stack
type InputParameters struct {
	StackID string `json:"stackID"`
}

// ResponseData : Data structure for returned data from Lambda function
type ResponseData struct {
	ResponseMessage string `json:"responseMesssage"`
}

// Required Lambda Code
func main() {
	lambda.Start(LambdaHandler)
}

// LambdaHandler : AWS Lambda function handler
func LambdaHandler(parameters InputParameters) (ResponseData, error) {

	var res ResponseData
	// Connection information
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	if e := setSubInactive(sess, parameters.StackID); e != nil {
		fmt.Println(e.Error())
		return res, e
	}

	if e := deleteStack(sess, parameters.StackID); e != nil {
		fmt.Printf("Got error deleting stack: %s", e.Error())
		return res, e
	}

	res.ResponseMessage = fmt.Sprintln("Deleted stack", parameters.StackID)

	return res, nil

}

// DeleteStack : Deletes the CloudFormation Stack based in its ID
func deleteStack(sess *session.Session, stackID string) error {

	// Create CloudFormation client in region
	svc := cloudformation.New(sess)

	delInput := &cloudformation.DeleteStackInput{StackName: aws.String(stackID)}
	_, err := svc.DeleteStack(delInput)
	if err != nil {
		return err
	}
	return nil
}

// SetSubInactive : Sets the 'subscription' table status key to "DELETED"
func setSubInactive(sess *session.Session, subscriptionid string) error {

	svc := dynamodb.New(sess)

	// The table name in DynamoDB
	var tableName = os.Getenv("SUBSCRIPTION_METADATA_TABLE")

	//Added
	// Updating the document with the value "Deleted"
	// for the subscriptionstatus attribute
	update := expression.Set(
		expression.Name("subscriptionstatus"),
		expression.Value("Deleted"),
	)

	expr, err := expression.NewBuilder().
		WithUpdate(update).
		Build()

	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		Key: map[string]*dynamodb.AttributeValue{
			"subscriptionID": {
				S: aws.String(subscriptionid),
			},
		},
		TableName:        aws.String(tableName),
		UpdateExpression: expr.Update(),
	}

	results, err := svc.UpdateItem(input) //Updating the item
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				fmt.Println(dynamodb.ErrCodeConditionalCheckFailedException, aerr.Error())
			case dynamodb.ErrCodeProvisionedThroughputExceededException:
				fmt.Println(dynamodb.ErrCodeProvisionedThroughputExceededException, aerr.Error())
			case dynamodb.ErrCodeResourceNotFoundException:
				fmt.Println(dynamodb.ErrCodeResourceNotFoundException, aerr.Error())
			case dynamodb.ErrCodeItemCollectionSizeLimitExceededException:
				fmt.Println(dynamodb.ErrCodeItemCollectionSizeLimitExceededException, aerr.Error())
			case dynamodb.ErrCodeTransactionConflictException:
				fmt.Println(dynamodb.ErrCodeTransactionConflictException, aerr.Error())
			case dynamodb.ErrCodeRequestLimitExceeded:
				fmt.Println(dynamodb.ErrCodeRequestLimitExceeded, aerr.Error())
			case dynamodb.ErrCodeInternalServerError:
				fmt.Println(dynamodb.ErrCodeInternalServerError, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return err
	}
	fmt.Println(results)

	return err
}
