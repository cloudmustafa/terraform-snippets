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
	"github.com/aws/aws-sdk-go/service/ec2"
)

// InputParameters : EC2 instance name passed in that
// will stop the server (not terminate)
type InputParameters struct {
	Ec2Instance string `json:"ec2Instance"`
}

// ResponseData : Data structure for returned data from Lambda function
type ResponseData struct {
	ResponseMessage string `json:"responseMesssage"`
}

func main() {

	// p := InputParameters{
	// 	Ec2Instance: "MaxEdge-55061d05-2be4-4a46-848b-222167e98d75",
	// }
	// LambdaHandler(p)
	lambda.Start(LambdaHandler)

}

// LambdaHandler : AWS Lambda function handler
func LambdaHandler(parameters InputParameters) (ResponseData, error) {

	var res ResponseData

	// Connection information
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	if e := getEc2InstanceID(sess, parameters.Ec2Instance); e != nil {
		fmt.Println(e.Error())
		return res, e
	}

	if e := SetSubActive(sess, parameters.Ec2Instance); e != nil {
		fmt.Println(e.Error())
		return res, e
	}

	res.ResponseMessage = fmt.Sprintln("Instance updated", parameters.Ec2Instance)

	return res, nil

}

func getEc2InstanceID(sess *session.Session, ec2Name string) error {

	svc := ec2.New(sess)

	// Passing in the tag:Name <key> for the instance
	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:Name"),
				Values: []*string{
					aws.String(ec2Name),
				},
			},
		},
	}

	// Getting the metadata from the EC2 instance
	result, err := svc.DescribeInstances(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
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

	//Getting the Instance ID
	// Declaring the variable to store Instance ID
	var instanceID string

	// Looping through the result metadata by first ranging Reservations
	// Then Ranging Instances to get to the instance ID
	for idx := range result.Reservations {
		for _, inst := range result.Reservations[idx].Instances {
			instanceID = *inst.InstanceId
		}
	}

	fmt.Println(result)
	fmt.Println(instanceID)

	if e := startEc2(sess, instanceID); e != nil {
		fmt.Println(e.Error())
		return e
	}
	return err
}

// startEc2 :
// The service name is the "id" from the DynamoDB table and is the
// the CloudFormation StackName
func startEc2(sess *session.Session, ec2InstanceID string) error {

	svc := ec2.New(sess)

	instanceID := ec2InstanceID

	input := &ec2.StartInstancesInput{
		InstanceIds: []*string{
			aws.String(instanceID),
		},
	}

	result, err := svc.StartInstances(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return err
	}
	fmt.Println(result)

	return err
}

// SetSubActive : Sets the 'subscription' table status key to "Restarted"
func SetSubActive(sess *session.Session, subscriptionid string) error {

	svc := dynamodb.New(sess)

	// The table name in DynamoDB
	var tableName = os.Getenv("SUBSCRIPTION_METADATA_TABLE")

	//Added
	// Updating the document with the value "Stopped"
	// for the subscriptionstatus attribute
	update := expression.Set(
		expression.Name("subscriptionstatus"),
		expression.Value("Restarted"),
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

	result, err := svc.UpdateItem(input) //Updating the item
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
	fmt.Println(result)

	return err
}
