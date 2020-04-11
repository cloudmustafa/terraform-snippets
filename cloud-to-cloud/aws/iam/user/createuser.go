package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
)

// InputParameters : Struct here to map to DynamoDB table attributes (struct in a struct)
type iamParameters struct {
	UserName string
}

type secretParameters struct {
	AccessKey string
	SecretKey string
}

type s3Parameters struct {
}

// ResponseData : Data structure for returned data from Lambda function
type ResponseData struct {
	ResponseMessage string `json:"responseMesssage"`
}

func main() {

	//lambda.Start(LambdaHandler)

	// Test code
	var uu = "test_subscriber"

	LambdaHandler(uu)
}

// LambdaHandler : AWS Lambda function handler
func LambdaHandler(user string) (ResponseData, error) {

	var res ResponseData
	var accessSecret secretParameters

	// Connection information
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	vuser, e := createUser(sess, user)

	if e != nil {
		fmt.Println(e.Error())
		return res, e
	}

	res.ResponseMessage = fmt.Sprintln("User created:", user)

	if _, e := createAccessKey(sess, vuser); e != nil {
		fmt.Println(e.Error())
		return res, e
	}

	res.ResponseMessage = fmt.Sprintln("Access Key created:", accessSecret.AccessKey)

	return res, nil

}
