package stoptask

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
)

// TargetMeta : Struct here to map to DynamoDB table attributes (struct in a struct)
type TargetMeta struct {
	AccesKey   string `json:"accessKey"`
	DeviceID   string `json:"deviceID"`
	Host       string `json:"host"`
	PolicyName string `json:"policyName"`
}

// APISubscription : Struct here to map to DynamoDB table attributes (main struct)
type APISubscription struct {
	AssetID     string     `json:"assetID"`
	ID          string     `json:"id"`
	Tags        []string   `json:"tags"`
	TargetMeta  TargetMeta `json:"targetMeta"`
	TargetName  string     `json:"targetName"`
	Version     int        `json:"version"`
	SubStatus   string     `json:"subStatus"`
	Environment string     `json:"environment"`
}

// ResponseData : Data structure for returned data from Lambda function
type ResponseData struct {
	ResponseMessage string `json:"responseMesssage"`
}

// StopTasks : Updates the desired count based on the service name.
// The service name is the "id" from the DynamoDB table and is the
// the CloudFormation StackName
func StopTasks(sess *session.Session, tasks []*string) error {

	svc := ecs.New(sess)

	var varClusterName = os.Getenv("CLUSTER_NAME")

	for _, task := range tasks {
		vTask := &ecs.StopTaskInput{
			Cluster: aws.String(varClusterName),
			Task:    task,
		}

		_, err := svc.StopTask(vTask)
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
	}
	return nil
}
