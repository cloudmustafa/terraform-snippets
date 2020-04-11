package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/aws/aws-sdk-go/service/ecs"
)

// InputParameters : Service name passed in that
// will update the desired count to 0 and will stop
// all tasks
type InputParameters struct {
	ServiceName string `json:"serviceName"`
}

// ResponseData : Data structure for returned data from Lambda function
type ResponseData struct {
	ResponseMessage string `json:"responseMesssage"`
}

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

	if e := updateService(sess, parameters.ServiceName); e != nil {
		fmt.Println(e.Error())
		return res, e
	}

	if e := SetSubInactive(sess, parameters.ServiceName); e != nil {
		fmt.Println(e.Error())
		return res, e
	}

	res.ResponseMessage = fmt.Sprintln("Service updated", parameters.ServiceName)

	return res, nil

}

// updateService : Updates the desired count based on the service name.
// The service name is the "id" from the DynamoDB table and is the
// the CloudFormation StackName
func updateService(sess *session.Session, serviceName string) error {

	svc := ecs.New(sess)

	var varClusterName = os.Getenv("CLUSTER_NAME")

	updateECS := &ecs.UpdateServiceInput{
		Cluster:            aws.String(varClusterName),
		DesiredCount:       aws.Int64(0),
		ForceNewDeployment: aws.Bool(true),
		Service:            aws.String(serviceName),
	}
	_, err := svc.UpdateService(updateECS)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ecs.ErrCodeServerException:
				fmt.Println(ecs.ErrCodeServerException, aerr.Error())
			case ecs.ErrCodeClientException:
				fmt.Println(ecs.ErrCodeClientException, aerr.Error())
			case ecs.ErrCodeInvalidParameterException:
				fmt.Println(ecs.ErrCodeInvalidParameterException, aerr.Error())
			case ecs.ErrCodeClusterNotFoundException:
				fmt.Println(ecs.ErrCodeClusterNotFoundException, aerr.Error())
			case ecs.ErrCodeServiceNotFoundException:
				fmt.Println(ecs.ErrCodeServiceNotFoundException, aerr.Error())
			case ecs.ErrCodeServiceNotActiveException:
				fmt.Println(ecs.ErrCodeServiceNotActiveException, aerr.Error())
			case ecs.ErrCodePlatformUnknownException:
				fmt.Println(ecs.ErrCodePlatformUnknownException, aerr.Error())
			case ecs.ErrCodePlatformTaskDefinitionIncompatibilityException:
				fmt.Println(ecs.ErrCodePlatformTaskDefinitionIncompatibilityException, aerr.Error())
			case ecs.ErrCodeAccessDeniedException:
				fmt.Println(ecs.ErrCodeAccessDeniedException, aerr.Error())
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

	return err
}

// SetSubInactive : Sets the 'subscription' table status key to "Stop"
func SetSubInactive(sess *session.Session, subscriptionid string) error {

	svc := dynamodb.New(sess)

	// The table name in DynamoDB
	var tableName = os.Getenv("SUBSCRIPTION_METADATA_TABLE")

	//Added
	// Updating the document with the value "Stopped"
	// for the subscriptionstatus attribute
	update := expression.Set(
		expression.Name("subscriptionstatus"),
		expression.Value("Stopped"),
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
