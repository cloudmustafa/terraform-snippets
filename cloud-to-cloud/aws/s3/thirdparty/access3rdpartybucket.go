package thirdparty

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// InputParameters : Struct here to map to DynamoDB table attributes (struct in a struct)
type InputParameters struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
}

// ResponseData : Data structure for returned data from Lambda function
type ResponseData struct {
	ResponseMessage string `json:"responseMesssage"`
}

func main() {

	var i = InputParameters{}
	//lambda.Start(LambdaHandler)
	LambdaHandler(i)
}

// LambdaHandler : AWS Lambda function handler
func LambdaHandler(parameter InputParameters) (ResponseData, error) {

	var res ResponseData

	// Connection information
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	if e := insertData(sess, parameter); e != nil {
		fmt.Println(e.Error())
		return res, nil
	}

	res.ResponseMessage = fmt.Sprintln("Bucket:", parameter.Bucket)

	return res, nil

}

// Stoping the task by setting the ECS service desired count to 0
func insertData(sess *session.Session, parameter InputParameters) error {

	svc := s3.New(sess)

	input := &s3.PutObjectInput{
		Bucket: aws.String(parameter.Bucket),
		Key:    aws.String(parameter.Key),
	}

	_, err := svc.PutObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				fmt.Println(s3.ErrCodeNoSuchKey, aerr.Error())
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
