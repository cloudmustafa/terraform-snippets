package runtask

import (
	"context"
	"fmt"

	//Need v2 for ecs
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	//Import to get task definition
)

// RunEcsTask : Creates and run a new task using the passed in Task Definition
func RunEcsTask(taskdefinition string) error {

	varTaskdefinitionName := taskdefinition
	//varTaskdefinitionName := "cloudtocloud-insertdata"
	//varCapacityProviderStrat := "FARGATE"
	varCluster := "cloudtocloud"
	//varTaskdefinitionName := "cloudtocloud-insertdata:1"
	varSubnets := []string{
		"subnet-00534747bcc25a82b",
		"subnet-0ac4e0869f688225c",
		"subnet-06f003ebf6501c183",
	}

	config, err := external.LoadDefaultAWSConfig()
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}

	// Set the AWS Region that the service clients should use
	config.Region = endpoints.UsEast1RegionID

	vpcConfig := &ecs.NetworkConfiguration{
		AwsvpcConfiguration: &ecs.AwsVpcConfiguration{

			SecurityGroups: []string{"sg-0a2b2e01fd4db7101"},
			Subnets:        varSubnets,
		},
	}

	svc := ecs.New(config)

	input := &ecs.RunTaskInput{
		LaunchType:           "FARGATE",
		Cluster:              aws.String(varCluster),
		Count:                aws.Int64(1),
		TaskDefinition:       aws.String(varTaskdefinitionName),
		NetworkConfiguration: vpcConfig,
		PlatformVersion:      aws.String("1.3.0"),
	}

	req := svc.RunTaskRequest(input)
	result, err := req.Send(context.Background())
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ecs.ErrCodeServerException:
				fmt.Println(ecs.ErrCodeServerException, aerr.Error())
			case ecs.ErrCodeException:
				fmt.Println(ecs.ErrCodeException, aerr.Error())
			case ecs.ErrCodeInvalidParameterException:
				fmt.Println(ecs.ErrCodeInvalidParameterException, aerr.Error())
			case ecs.ErrCodeClusterNotFoundException:
				fmt.Println(ecs.ErrCodeClusterNotFoundException, aerr.Error())
			case ecs.ErrCodeUnsupportedFeatureException:
				fmt.Println(ecs.ErrCodeUnsupportedFeatureException, aerr.Error())
			case ecs.ErrCodePlatformUnknownException:
				fmt.Println(ecs.ErrCodePlatformUnknownException, aerr.Error())
			case ecs.ErrCodePlatformTaskDefinitionIncompatibilityException:
				fmt.Println(ecs.ErrCodePlatformTaskDefinitionIncompatibilityException, aerr.Error())
			case ecs.ErrCodeAccessDeniedException:
				fmt.Println(ecs.ErrCodeAccessDeniedException, aerr.Error())
			case ecs.ErrCodeBlockedException:
				fmt.Println(ecs.ErrCodeBlockedException, aerr.Error())
			default:
				fmt.Println(aerr.Error())

				return aerr
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
	}

	fmt.Println(result)
	return nil
}
