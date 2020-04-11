package main

import (
	e64 "encoding/base64"
	"fmt"
	"os"

	"github.com/awslabs/goformation/cloudformation"
	"github.com/awslabs/goformation/cloudformation/ec2"
	uuid "github.com/satori/go.uuid"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/awslabs/goformation/cloudformation/tags"

	"maxedge/cloud-to-cloud/aws/dynamodb"
	p "maxedge/cloud-to-cloud/aws/ec2/piconfig"

	cfn "github.com/aws/aws-sdk-go/service/cloudformation"
)

// InputParameters : The stack ID is the random generated ID that is used as the ID and Name for subscription resources
// and is passed in from the UI. Other than the StackID, are the values needed for the Pi interface server
type InputParameters struct {
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

	// Test code
	// p := InputParameters{
	// 	StackID:         "MaxEdge-ddc65b40-5d03-4827-8bf9-2957903600ca-test-2m",
	// 	SourceIP:        "EC2AMAZ-5EEVEUI",
	// 	SourcePort:      "5450",
	// 	DestinationIP:   "10.0.4.53",
	// 	DestinationPort: "5450",
	// 	PiList:          []string{"tag1", "tag2", "tag3", "tag4", "tag5", "tag6", "endTag117"},
	// }
	//LambdaHandler(p)

	lambda.Start(LambdaHandler)
}

// LambdaHandler : AWS Lambda function handler
func LambdaHandler(parameters InputParameters) (ResponseData, error) {

	var res ResponseData
	var t = parameters.Subscription.Assets

	// Used to identify the EC2 instance as part of the subscription
	id := GenerateStackID()

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Creates CloudFormation stack
	if e := CreateStackResources(id, parameters.Subscription); e != nil {
		fmt.Println(e.Error())
		return res, e
	}

	// Creates the Pi output file
	rp, e := p.CreatePiConfig(id, t)
	if e != nil {
		fmt.Println(e.Error())
		return res, e
	}

	// Inserts data into the 'subscriptions' table
	if e := dynamodb.InsertIntoDynamoDB(sess, id, parameters.Subscription); e != nil {
		fmt.Printf("Got error inserting data: %s", e.Error())
		return res, e
	}

	//Need to return the data to user in a response
	// res.ResponseMessage = fmt.Sprintln("Config File: ", rp)

	return res, nil
}

// CLOUDFORMATION CODE

// CreateStackResources : Creates the stack resources using the parameters passed in
func CreateStackResources(stackID string, subscription dynamodb.APISubscription) error {

	// Lamba environment variables

	var amiID = os.Getenv("AMI_ID")
	var securityGroup = os.Getenv("SECURITY_GROUP")
	var subNet = os.Getenv("SUBNET")
	var keyName = os.Getenv("KEY_NAME")
	var instanceType = os.Getenv("INSTANCE_TYPE")
	var iamRole = os.Getenv("IAM_ROLE")

	// Create a new CloudFormation template
	template := cloudformation.NewTemplate()

	// StackID with 'PI' as a prefix to avoid
	// duplicate stack names in CloudFormation
	id := stackID

	tags := []tags.Tag{
		tags.Tag{
			Key:   "Product",
			Value: "MaxEdge",
		},
		// This tag is used to display the "Name" in the console
		tags.Tag{
			Key:   "Name",
			Value: id,
		},
	}

	// Declaring the PI .exe for teh service to deal with the qoutes
	// being stripped off and not workign in PowerShell
	piToPi := `C:\Program Files (x86)\PIPC\Interfaces\PItoPI\PItoPI.exe`

	// Pi .bat file needed for the Pi interface service
	// The source, destination and ports are added to this string and passed in as parameters
	// from the user.
	// The PowerShell commands that updates the .bat file that is on the AMI
	// with the client passed in parameters.
	// PowerShell command and the "<powershell></powershell>" is required when
	// using UserDate in EC2 and using PowerShell commands
	// Then changes the directory to the location of the nssm program to start the PItoPIInt service.
	// Services will restart on reboot or instance restart
	userData := fmt.Sprintf(`<powershell> 
	Set-Content -Path "C:\Program Files (x86)\PIPC\Interfaces\PItoPI\PItoPIInt.bat" -Value '"`+piToPi+`" Int /src_host=%s:%s /PS=%s /ID=%s /host=%s:%s /pisdk=0 /maxstoptime=120 /PercentUp=100 /sio /perf=8 /f=00:00:01'
	cd "C:\Program Files (x86)\PIPC\Interfaces\PItoPI\"; .\nssm start PItoPIInt 
	</powershell>
	<persist>true</persist>`, subscription.SourceIP, subscription.SourcePort, id, id, subscription.DestinationIP, subscription.DestinationPort)

	// Required base64 encoding
	// If you are using a command line tool, base64-encoding is performed
	// for you, and you can load the text from a file. Otherwise, you must provide
	// base64-encoded text. User data is limited to 16 KB.
	encodeUserDate := e64.StdEncoding.EncodeToString([]byte(userData))

	// CloudFormation E2 properties
	template.Resources["EC2Pi"] = &ec2.Instance{
		ImageId:            (amiID),
		KeyName:            (keyName),
		InstanceType:       (instanceType),
		SecurityGroupIds:   []string{securityGroup},
		SubnetId:           (subNet),
		IamInstanceProfile: (iamRole),
		UserData:           encodeUserDate,
		Tags:               (tags),
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

// Create a CloudFormation stack given a body template
// Method has a maximum template size of 51,200 bytes
// https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/cloudformation-limits.html
func createStackFromBody(sess client.ConfigProvider, templateBody []byte, stackName string) error {

	// Tags for the CloudFormation stack
	tags := []*cfn.Tag{
		{
			Key:   aws.String("Product"),
			Value: aws.String("MaxEdge"),
		},
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

	// Creates the stack
	_, err := svc.CreateStack(input)
	if err != nil {
		fmt.Println("Got error creating stack:")
		fmt.Println(err.Error())
		return err
	}
	return nil
}
