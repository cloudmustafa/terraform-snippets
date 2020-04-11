package starttask

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"

	"maxedge/cloud-to-cloud/update/gettask"
	"maxedge/cloud-to-cloud/update/stoptask"
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

// Starting the task by setting the ECS service desired count to 1

func startService(sess *session.Session, subscription APISubscription, tasks []*string) error {

	svc := ecs.New(sess)

	var varClusterName = os.Getenv("CLUSTER_NAME")

	for _, task := range tasks {
		vTask := &ecs.StopTaskInput{
			Cluster: aws.String(varClusterName),
			Task:    task,
		}
		waiting := &ecs.DescribeTasksInput{
			Cluster: varClusterName,
			Tasks:   tasks,
		}

	}

	updateECS := &ecs.UpdateServiceInput{
		Cluster:            aws.String(varClusterName),
		DesiredCount:       aws.Int64(1),
		ForceNewDeployment: aws.Bool(true),
		Service:            aws.String(subscription.ID),
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

// LambdaHandler : AWS Lambda function handler
func LambdaHandler(subscription gettask.APISubscription) (ResponseData, error) {

	var res ResponseData
	var t []*string

	// Connection information
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	if t, e := gettask.GetTask(sess, subscription); e != nil {
		fmt.Println(e.Error())
		return res, e
	}

	if e := stoptask.StopTasks(sess, t); e != nil {
		fmt.Println(e.Error())
		return res, e
	}

	res.ResponseMessage = fmt.Sprintln("Service updated", subscription.ID)

	return res, nil

}
