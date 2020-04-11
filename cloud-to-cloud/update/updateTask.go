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

// MetaData : Struct here to map to DynamoDB table attributes (struct in a struct)
type MetaData struct {
	HostName             string `json:"hostname"`
	DeviceID             string `json:"deviceid"`
	TopicName            string `json:"topicname"`
	ClientCertificate    string `json:"clientcertificate"`
	ClientKey            string `json:"clientkey"`
	CertificateAuthority string `json:"certificateauthority"`
	AccessKey            string `json:"accesskey"`
	Secret               string `json:"secret"`
	Location             string `json:"location"`
	Frequency            int    `json:"frequency"`
	PrivateKey           string `json:"privatekey"`
	ProjectID            string `json:"projectid"`
	Region               string `json:"region"`
	RegistryID           string `json:"registryid"`
	SasToken             string `json:"sastoken"`
	PolicyName           string `json:"policyname"`
}

// Asset : Struct here to map to DynamoDB table attributes (struct in a struct)
type Asset struct {
	AssetID int      `json:"assetid"`
	Name    string   `json:"name"`
	Tags    []string `json:"tags"`
}

// APISubscription : Struct here to map to DynamoDB table attributes (main struct)
type APISubscription struct {
	SubscriptionID string   `json:"subscriptionID"`
	Name           string   `json:"name"`
	Assets         []Asset  `json:"assets"`
	Target         string   `json:"target"`
	MetaData       MetaData `json:"metadata"`
	StartTIme      int      `json:"starttime"`
	EndTime        int      `json:"endtime"`
}

// ResponseData : Data structure for returned data from Lambda function
type ResponseData struct {
	ResponseMessage string `json:"responseMesssage"`
}

func main() {

	lambda.Start(LambdaHandler)

}

// LambdaHandler : AWS Lambda function handler
func LambdaHandler(subscription APISubscription) (ResponseData, error) {

	var res ResponseData

	// Connection information
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	if e := stopService(sess, subscription); e != nil {
		fmt.Println(e.Error())
		return res, nil
	}

	if e := checkTaskStatus(sess, subscription); e != nil {
		fmt.Println(e.Error())
		return res, e
	}

	if e := updateMetaData(sess, subscription); e != nil {
		fmt.Println(e.Error())
		return res, e
	}

	res.ResponseMessage = fmt.Sprintln("Service updated", subscription.SubscriptionID)

	return res, nil

}

// Stoping the task by setting the ECS service desired count to 0
func stopService(sess *session.Session, subscription APISubscription) error {

	svc := ecs.New(sess)

	var varClusterName = os.Getenv("CLUSTER_NAME")

	updateECS := &ecs.UpdateServiceInput{
		Cluster:            aws.String(varClusterName),
		DesiredCount:       aws.Int64(0),
		ForceNewDeployment: aws.Bool(true),
		Service:            aws.String(subscription.SubscriptionID),
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

// Function that checks the Service RunningCount and once it has the value of 0
// the function calls startService.
func checkTaskStatus(sess *session.Session, subscription APISubscription) error {

	svc := ecs.New(sess)

	var varClusterName = os.Getenv("CLUSTER_NAME")

	statusResults := &ecs.DescribeServicesInput{
		Cluster: aws.String(varClusterName),
		Services: []*string{
			aws.String(subscription.SubscriptionID),
		},
	}

getRunningCount:
	result, err := svc.DescribeServices(statusResults)
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

	// Looping thru the Service until all tasks are stopped
	for _, service := range result.Services {
		if *service.RunningCount == 0 {
			{

				startService(sess, subscription)

				break
			}

		} else {

			goto getRunningCount
		}
	}

	return err
}

// Starting the task by setting the ECS service desired count to 1

func startService(sess *session.Session, subscription APISubscription) error {

	svc := ecs.New(sess)

	var varClusterName = os.Getenv("CLUSTER_NAME")

	updateECS := &ecs.UpdateServiceInput{
		Cluster:            aws.String(varClusterName),
		DesiredCount:       aws.Int64(1),
		ForceNewDeployment: aws.Bool(true),
		Service:            aws.String(subscription.SubscriptionID),
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

func updateMetaData(sess *session.Session, subscription APISubscription) error {

	svc := dynamodb.New(sess)

	// The table name in DynamoDB
	var tableName = os.Getenv("SUBSCRIPTION_METADATA_TABLE")

	update := expression.Set(
		expression.Name("target"),
		expression.Value(subscription.Target),
	).Set(
		expression.Name("metadata"),
		expression.Value(subscription.MetaData),
	).Set(
		expression.Name("assets"),
		expression.Value(subscription.Assets),
	)

	expr, err := expression.NewBuilder().
		WithUpdate(update).
		Build()

	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		Key: map[string]*dynamodb.AttributeValue{
			"subscriptionID": {
				S: aws.String(subscription.SubscriptionID),
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
