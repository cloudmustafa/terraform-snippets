package main

import (
	"encoding/json"
	"fmt"
	"os"

	uuid "github.com/satori/go.uuid"

	"maxedge/cloud-to-cloud/aws/dynamodb"
	accesskeysecretkey "maxedge/cloud-to-cloud/aws/iam/accesskey"
	s3prefix "maxedge/cloud-to-cloud/aws/s3"
	sm "maxedge/cloud-to-cloud/aws/secretsmanager"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/awslabs/goformation/cloudformation"
	"github.com/awslabs/goformation/cloudformation/ecs"
	"github.com/awslabs/goformation/cloudformation/iam"
	"github.com/awslabs/goformation/cloudformation/secretsmanager"
	"github.com/awslabs/goformation/cloudformation/tags"

	// Name conflict, requires an explicit declaration
	// https://golang.org/ref/spec#Import_declarations
	cfn "github.com/aws/aws-sdk-go/service/cloudformation"
)

// InputParameters : Data structure for input paramters into Lambda function
type InputParameters struct {

	//JSON object passed in from /start API
	Subscription dynamodb.APISubscription `json:"Subscription"`
}

// GenerateStackID : Generates ID that is used as unique ID for different services
// Prefixed 'MaxEdge-' so there will always be a letter as the first character,
// due to CloudFormation StackName constraint
func GenerateStackID() string {
	return `MaxEdge-` + uuid.Must(uuid.NewV4()).String()
}

// ResponseData : Data structure for returned data from Lambda function
type ResponseData struct {
	ResponseMessage string `json:"responseMesssage"`
}

// Main function reqired by Lambda
func main() {

	lambda.Start(LambdaHandler)
}

// LambdaHandler : AWS Lambda function handler
func LambdaHandler(parameters InputParameters) (ResponseData, error) {

	var res ResponseData

	id := GenerateStackID()

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Creates S3 bucket
	if e := s3prefix.SubscriptionS3Folder(sess, id); e != nil {
		fmt.Println(e.Error())
		return res, e
	}

	// Creates CloudFormation stack
	if e := CreateStackResources(id, parameters.Subscription); e != nil {
		fmt.Println(e.Error())
		return res, e
	}

	// Inserts data into the 'subscriptions' table
	if e := dynamodb.InsertIntoDynamoDB(sess, id, parameters.Subscription); e != nil {
		fmt.Printf("Got error inserting data: %s", e.Error())
		return res, e
	}

	// Create user & access key
	akey, skey, e := accesskeysecretkey.CreateAccessSecretKey(sess, id)
	if e != nil {
		fmt.Println(e.Error())
		return res, e
	}

	// Inserts data into Secrets Manager
	if e := sm.SaveAccessKeySecretKey(sess, id, akey, skey); e != nil {
		fmt.Printf("Got error inserting Access and Secret Keys: %s", e.Error())
		return res, e
	}

	sss := map[string]string{"Access": akey, "Secret": skey}
	s, _ := json.Marshal(sss)
	fmt.Println(string(s))
	//k := string(s)

	res.ResponseMessage = fmt.Sprintln(string(s))
	// res.ResponseMessage = fmt.Sprintln("Stack created.", id)
	// res.ResponseMessage = fmt.Sprintln("Stack created.", akey, skey)

	return res, nil
}

// CLOUDFORMATION CODE

// CreateStackResources : Creates the stack resources using the parameters passed in
func CreateStackResources(stackID string, subscription dynamodb.APISubscription) error {

	// Create a new CloudFormation template
	template := cloudformation.NewTemplate()

	//Generate GUID
	id := stackID

	tags := []tags.Tag{
		tags.Tag{
			Key:   "Product",
			Value: "MaxEdge",
		},
		// tags.Tag{
		// 	Key:   "Subscription_ID",
		// 	Value: subscription.Name,
		// },
	}

	// AWS ECS Prpoperties
	varRequiresCompatibilities := []string{"FARGATE"}

	// Lambda Environment Variables //

	// AWS Account ID
	var varAwsAccountID = os.Getenv("AWS_ACCOUNT_ID")

	// Task Definition
	var varExecutionRoleArn = os.Getenv("EXECUTION_ROLE")
	var varTaskRoleArn = os.Getenv("TASK_ROLE")
	var varEcsTaskDefinitionRef = os.Getenv("CLOUDFORMATION_TASK_DEFINITION_REF")

	// Container Definition
	var varImage = os.Getenv("CONTAINER_IMAGE")

	//Network Definition
	var varSecurityGroup = os.Getenv("SECURITY_GROUP")
	var varSubnet1 = os.Getenv("SUBNET_1")
	var varSubnet2 = os.Getenv("SUBNET_2")
	var varSubnet3 = os.Getenv("SUBNET_3")

	//ECS Service
	var varClusterName = os.Getenv("CLUSTER_NAME")
	var varEcsRef = os.Getenv("CLOUDFORMATION_ECS_SERVICE_REF")

	//Secret
	var varSecretRef = os.Getenv("CLOUDFORMATION_SECRET_REF")

	// Create IAM User
	template.Resources["IAMUSER"] = &iam.User{
		UserName: string(id),
		Tags:     tags,
	}

	// Assigning the subscribers ID to a string so it can be added to the policy
	bucket := fmt.Sprintf("\"arn:aws:s3:::maxedgecloudtocloudpoc-sandbox/%s/*\"", id)
	var roleName string = "ROLE_" + id
	var policyName string = "Policy" + id

	// S3 GetObject policy for the created IAM user for a subscription
	// User will assume the role
	// Action is to Assume the sts role
	template.Resources["IAMPolicy"] = &iam.Policy{
		PolicyName:                 policyName,
		PolicyDocument:             ("{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Action\":\"sts:AssumeRole\",\"Resource\":\"arn:aws:iam::" + varAwsAccountID + ":role/" + roleName + "\"}]}"),
		Users:                      []string{id},
		AWSCloudFormationDependsOn: []string{"IAMUSER"},
	}

	p := iam.Role_Policy{
		PolicyDocument: ("{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Action\":\"s3:GetObject\",\"Resource\":" + bucket + "}]}"),
		PolicyName:     id,
	}

	// Assume Role or Trust Policy
	template.Resources["IAMROLE"] = &iam.Role{
		AssumeRolePolicyDocument:   ("{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"AWS\": \"arn:aws:iam::" + varAwsAccountID + ":user/" + id + "\"},\"Action\":[\"sts:AssumeRole\"]}]}"),
		RoleName:                   roleName,
		Policies:                   []iam.Role_Policy{p},
		AWSCloudFormationDependsOn: []string{"IAMUSER"},
		Tags:                       tags,
	}

	// Task Definition Secret - sets the value for the container based on what was
	// created in the template.
	TaskDefSecret := []ecs.TaskDefinition_Secret{}
	varTaskDefSecret := ecs.TaskDefinition_Secret{
		//Referance to the ARNs created when the secret is created
		Name:      cloudformation.Ref(varSecretRef),
		ValueFrom: cloudformation.Ref(varSecretRef),
	}
	TaskDefSecret = append(TaskDefSecret, varTaskDefSecret)

	// TargetName come from the DynamoDB table and stores the cloud platform
	// endpoint for the data destination
	group := "/MaxEdge/CloudToCloud/" + subscription.Target

	cwLog := &ecs.TaskDefinition_LogConfiguration{
		LogDriver: "awslogs",
		Options:   map[string]string{"awslogs-create-group": "true", "awslogs-group": group, "awslogs-region": "us-east-1", "awslogs-stream-prefix": id},
	}

	//Task Definition Container Definition - Setting properties
	TaskDefConDef := []ecs.TaskDefinition_ContainerDefinition{}
	varTaskDef := ecs.TaskDefinition_ContainerDefinition{
		Image:            varImage,
		Name:             id,
		Secrets:          TaskDefSecret,
		LogConfiguration: cwLog,
	}
	TaskDefConDef = append(TaskDefConDef, varTaskDef)

	// Create an Amazon ECS Task Definition - Setting properties
	template.Resources[varEcsTaskDefinitionRef] = &ecs.TaskDefinition{
		ExecutionRoleArn:        varExecutionRoleArn,
		TaskRoleArn:             varTaskRoleArn,
		Memory:                  "1024",
		NetworkMode:             "awsvpc",
		RequiresCompatibilities: varRequiresCompatibilities,
		Cpu:                     "512",
		ContainerDefinitions:    TaskDefConDef,
		Tags:                    tags,
	}

	// ECS Service Network configuration
	snc := &ecs.Service_NetworkConfiguration{
		AwsvpcConfiguration: &ecs.Service_AwsVpcConfiguration{

			// Required to access ECR
			AssignPublicIp: "ENABLED",
			// The Security Group needs to allow traffic via port :443
			SecurityGroups: []string{varSecurityGroup},
			Subnets:        []string{varSubnet1, varSubnet2, varSubnet3},
		},
	}

	// Create Amazon ECS Service
	template.Resources[varEcsRef] = &ecs.Service{
		LaunchType:           "FARGATE",
		Cluster:              varClusterName,
		TaskDefinition:       cloudformation.Ref(varEcsTaskDefinitionRef),
		ServiceName:          id,
		DesiredCount:         1,
		NetworkConfiguration: snc,
		SchedulingStrategy:   "REPLICA",
		Tags:                 tags,
		PropagateTags:        "TASK_DEFINITION",
	}

	// Create an Amazon Secret
	template.Resources[varSecretRef] = &secretsmanager.Secret{
		Name:        id,
		Description: "Metadata for companies",
		Tags:        tags,
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

	return nil
}

func createStackFromBody(sess client.ConfigProvider, templateBody []byte, stackName string) error {
	// Create a CloudFormation stack given a body template
	// Method has a maximum template size of 51,200 bytes
	// https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/cloudformation-limits.html

	// TAGS
	// New Code 1/28/2020
	// These are test tags
	tags := []*cfn.Tag{
		{
			Key:   aws.String("Product"),
			Value: aws.String("MaxEdge"),
		},
		// 	{
		// 		Key:   aws.String("Other"),
		// 		Value: aws.String("Test"),
		// 	},
	}

	//Creates CloudFormation stack
	svc := cfn.New(sess)
	input := &cfn.CreateStackInput{
		TemplateBody: aws.String(string(templateBody)),
		StackName:    aws.String(stackName),
		Tags:         tags,
		Capabilities: []*string{aws.String("CAPABILITY_NAMED_IAM")}, // Required because of creating a stack that is creating IAM resources
	}

	fmt.Println("Stack creation initiated...")

	_, err := svc.CreateStack(input)
	if err != nil {
		fmt.Println("Got error creating stack:")
		fmt.Println(err.Error())
		return err
	}
	return nil
}
