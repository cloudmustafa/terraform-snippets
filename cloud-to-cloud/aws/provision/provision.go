package provision

import (
	"fmt"

	"maxedge/cloud-to-cloud/aws/dynamodb"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/awslabs/goformation/cloudformation"
	"github.com/awslabs/goformation/cloudformation/ecs"
	"github.com/awslabs/goformation/cloudformation/secretsmanager"

	// Name conflict, requires an explicit declaration
	// https://golang.org/ref/spec#Import_declarations
	cfn "github.com/aws/aws-sdk-go/service/cloudformation"
)

//CLOUDFORMATION CODE

// CreateStackResources : Creates the stack resources using the parameters passed in
func CreateStackResources(stackID, subscriptionID string, subscription dynamodb.APISubscription) error {
	// Create a new CloudFormation template
	template := cloudformation.NewTemplate()

	//Generate GUID
	id := stackID

	//AWS ECS Prpoperties
	varRequiresCompatibilities := []string{"FARGATE"}

	// Task Definition Secret - sets the value for the container based on what was
	// created in the template.
	TaskDefSecret := []ecs.TaskDefinition_Secret{}
	varTaskDefSecret := ecs.TaskDefinition_Secret{
		//Referance to the ARNs created when the secret is created
		Name:      cloudformation.Ref("MaxEdgeCloudToCloud"),
		ValueFrom: cloudformation.Ref("MaxEdgeCloudToCloud"),
	}
	TaskDefSecret = append(TaskDefSecret, varTaskDefSecret)

	//Task Definition Container Definition - Setting properties
	TaskDefConDef := []ecs.TaskDefinition_ContainerDefinition{}
	varTaskDef := ecs.TaskDefinition_ContainerDefinition{
		Image:   "007131566380.dkr.ecr.us-east-1.amazonaws.com/cloudtocloud",
		Name:    id,
		Secrets: TaskDefSecret,
	}
	TaskDefConDef = append(TaskDefConDef, varTaskDef)

	// Create an Amazon ECS Task Definition - Setting properties
	template.Resources["ECSTaskDefinition"] = &ecs.TaskDefinition{
		ExecutionRoleArn:        "arn:aws:iam::007131566380:role/ecsTaskExecutionRole",
		TaskRoleArn:             "arn:aws:iam::007131566380:role/ecsTaskExecutionRole",
		Memory:                  "1024",
		NetworkMode:             "awsvpc",
		RequiresCompatibilities: varRequiresCompatibilities,
		Cpu:                     "512",
		ContainerDefinitions:    TaskDefConDef,
	}

	// ECS Service Network configuration
	snc := &ecs.Service_NetworkConfiguration{
		AwsvpcConfiguration: &ecs.Service_AwsVpcConfiguration{
			AssignPublicIp: "ENABLED",
			SecurityGroups: []string{"sg-0a2b2e01fd4db7101"},
			Subnets:        []string{"subnet-0ac4e0869f688225c"},
		},
	}

	// ECS Service Deployment configuration
	dc := &ecs.Service_DeploymentConfiguration{
		MaximumPercent:        0,
		MinimumHealthyPercent: 100,
	}

	// Create Amazon ECS Service
	template.Resources["ECSService"] = &ecs.Service{
		LaunchType:              "FARGATE",
		Cluster:                 "cloudtocloud",
		TaskDefinition:          cloudformation.Ref("ECSTaskDefinition"),
		ServiceName:             (id),
		DesiredCount:            1,
		NetworkConfiguration:    snc,
		DeploymentConfiguration: dc,
	}

	// Passing in the 'Stack ID and the Environment variable
	// 'sbx', 'dev', 'qa', 'prod'
	secretValue := fmt.Sprintf("{\"ID\":\"%s\",\"ENVIRONMENT\":\"sbx\"}", id)

	// Create an Amazon Secret
	template.Resources["MaxEdgeCloudToCloud"] = &secretsmanager.Secret{
		Name:         string(id),
		Description:  "Metadata for companies",
		SecretString: (secretValue),
	}

	// Get JSON form of AWS CloudFormation template
	j, err := template.JSON()
	if err != nil {
		fmt.Printf("Failed to generate JSON: %s\n", err)
		return err
	}
	fmt.Printf("Template creation for %s Done.\n", id)
	fmt.Println("=====")
	fmt.Println("Generated template:")
	fmt.Printf("%s\n", string(j))
	fmt.Println("=====")

	// Initialize a session that the SDK uses to load
	// credentials from the shared credentials file ~/.aws/credentials
	// and configuration from the shared configuration file ~/.aws/config.
	// https://docs.aws.amazon.com/sdk-for-go/api/aws/session/
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create the stack
	err = createStackFromBody(sess, j, id)
	if err != nil {
		return err
	}

	//Run Task
	// err = runECSRunTask(id)
	// if err != nil {
	// 	return err
	// }
	return nil
}

func createStackFromBody(sess client.ConfigProvider, templateBody []byte, stackName string) error {
	// Create a CloudFormation stack given a body template
	// Method has a maximum template size of 51,200 bytes
	// https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/cloudformation-limits.html
	svc := cfn.New(sess)
	input := &cfn.CreateStackInput{TemplateBody: aws.String(string(templateBody)), StackName: aws.String(stackName)}

	fmt.Println("Stack creation initiated...")

	_, err := svc.CreateStack(input)
	if err != nil {
		fmt.Println("Got error creating stack:")
		fmt.Println(err.Error())
		return err
	}
	return nil
}

// Function that runs the task based on the CloudFormation staci name
// func runECSRunTask(stackName string) error {

// 	ri, e := getphysicalid.GetPhysicalID(stackName)
// 	if e != nil {
// 		return e
// 	}

// 	err := runtask.RunEcsTask(ri)
// 	fmt.Printf(ri)
// 	if err != nil {
// 		fmt.Println("Got error running task:")
// 		fmt.Println(err.Error())
// 		return err
// 	}
// 	return nil
// }
