package ec2direect

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// InputParameters : Subscription ID for input paramters into Lambda function
type InputParameters struct {

	//Subscription Information
	SubscriptionID string `json:"subscriptionID"`
}

// ResponseData : Data structure for returned data from Lambda function
type ResponseData struct {
	ResponseMessage string `json:"responseMesssage"`
}

// func main() {

// 	lambda.Start(LambdaHandler)

// }

// LambdaHandler : AWS Lambda function handler
func LambdaHandler(parameters InputParameters) (ResponseData, error) {

	var res ResponseData
	subID := parameters.SubscriptionID

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	_, e := createEc2(sess, subID)
	if e != nil {
		fmt.Println(e.Error())
		return res, e
	}

	res.ResponseMessage = fmt.Sprintln(e)

	return res, nil
}

func createEc2(sess *session.Session, subID string) (ResponseData, error) {

	var res ResponseData
	var amiID = os.Getenv("AMI_ID")
	//var securityGroup = os.Getenv("SECURITY_GROUP")
	//var subNet = os.Getenv("SUB_NET")
	var keyName = os.Getenv("KEY_NAME")
	var instanceType = os.Getenv("INSTANCE_TYPE")

	//var batFile = "<script>\"C:\\Program Files (x86)\\PIPC\\Interfaces\\PItoPI\\PItoPI.exe\" NOVOS_PD584 /src_host=SRVGDYCDSPI01.cds.nov.com:5450 /rh_inc=1 /PTID /DB=2,3,5 /PS=NOVOS_PD584 /ID=584 /host=SRVGDYCDSPIQ02:5450 /pisdk=0 /maxstoptime=120 /PercentUp=100 /sio /perf=8 /f=00:00:02 /f=00:00:05 </script><persist>true</persist>"

	input := &ec2.RunInstancesInput{

		BlockDeviceMappings: []*ec2.BlockDeviceMapping{
			{
				DeviceName: aws.String("/dev/sdh"),
				Ebs: &ec2.EbsBlockDevice{
					VolumeSize: aws.Int64(100),
				},
			},
		},
		ImageId:      aws.String(amiID),
		InstanceType: aws.String(instanceType),
		KeyName:      aws.String(keyName),
		MaxCount:     aws.Int64(1),
		MinCount:     aws.Int64(1),
		// SecurityGroups: []*string{
		// 	aws.String(securityGroup),
		// },
		// SecurityGroupIds: []*string{
		// 	aws.String(securityGroup),
		// },
		// SubnetId: aws.String(subNet),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("instance"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("Product"),
						Value: aws.String("MaxEdge"),
					},
				},
			},
		},
		//	UserData: aws.String(batFile),
		//	DryRun: aws.Bool(true),
	}

	svc := ec2.New(sess)

	result, err := svc.RunInstances(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return res, err
	}
	res.ResponseMessage = fmt.Sprintln("EC2 build:", subID)

	fmt.Println(result)
	return res, err
}
