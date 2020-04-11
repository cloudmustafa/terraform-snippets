package getphysicalid

// package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// GetPhysicalID : Function that takes in a StackName and returns the PhysicalID
// of the ECS Task Definition
func GetPhysicalID(stackID string) (string, error) {

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	cf := cloudformation.New(sess)

	//Describe the stack to get the resource status by passing stack name
	cfe := &cloudformation.DescribeStackEventsInput{
		StackName: aws.String(stackID),
	}

	// Checks the CloudFormation stack event
getEvents:
	ev, err := cf.DescribeStackEvents(cfe)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case cloudformation.ErrCodeStackInstanceNotFoundException:
				fmt.Println(cloudformation.ErrCodeStackInstanceNotFoundException, aerr.Error())
			default:
				fmt.Println(aerr.Error())

				return "", aerr
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
	}

	// Looping thru the CloudFormation Stack Events to see
	// when the Resource Status is = to "CREATE_COMPLETE"
	for _, event := range ev.StackEvents {
		if *event.LogicalResourceId == "ECSTaskDefinition" && *event.ResourceStatus == "CREATE_COMPLETE" {
			{
				break
			}

		} else {
			goto getEvents
		}
	}

	println(awsutil.StringValue(ev.StackEvents))

	// Gets Information about the stack, base on the ID
	// that is passed in
	params := &cloudformation.DescribeStackResourceInput{
		LogicalResourceId: aws.String("ECSTaskDefinition"), // Required
		StackName:         aws.String(stackID),             // Required
	}

	resp, err := cf.DescribeStackResource(params)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case cloudformation.ErrCodeStackInstanceNotFoundException:
				fmt.Println(cloudformation.ErrCodeStackInstanceNotFoundException, aerr.Error())
			default:
				fmt.Println(aerr.Error())

				return "", aerr
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
	}
	// Pretty-print the response data.
	//fmt.Println(awsutil.StringValue(resp))

	println(awsutil.StringValue(resp.StackResourceDetail.PhysicalResourceId))
	return *resp.StackResourceDetail.PhysicalResourceId, nil

}
