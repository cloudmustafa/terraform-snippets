package main

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

// InputParameters : Data structure for input paramters into Lambda function
type InputParameters struct {

	//Subscription Information
	SubscriptionID string `json:"subscriptionID"`
}

// ResponseData : Data structure for returned data from Lambda function
type ResponseData struct {
	ResponseMessage string `json:"responseMesssage"`
}

// SecretString : Data structure to populate the access/secret key
type SecretString struct {
	AccessKey string `json:"ACCESS"`
	SecretKey string `json:"SECRET"`
}

func main() {

	lambda.Start(LambdaHandler)

}

// LambdaHandler : AWS Lambda function handler
func LambdaHandler(parameters InputParameters) (ResponseData, error) {

	var res ResponseData
	subID := parameters.SubscriptionID

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	r, e := getKeyFromSecret(sess, subID)
	if e != nil {
		fmt.Println(e.Error())
		return res, e
	}

	res.ResponseMessage = fmt.Sprintln(r)

	return res, nil
}

func getKeyFromSecret(sess *session.Session, subID string) (ResponseData, error) {

	var res ResponseData

	svc := secretsmanager.New(sess)

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(subID),
	}

	result, err := svc.GetSecretValue(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case secretsmanager.ErrCodeResourceNotFoundException:
				fmt.Println(secretsmanager.ErrCodeResourceNotFoundException, aerr.Error())
			case secretsmanager.ErrCodeInvalidParameterException:
				fmt.Println(secretsmanager.ErrCodeInvalidParameterException, aerr.Error())
			case secretsmanager.ErrCodeInvalidRequestException:
				fmt.Println(secretsmanager.ErrCodeInvalidRequestException, aerr.Error())
			case secretsmanager.ErrCodeDecryptionFailure:
				fmt.Println(secretsmanager.ErrCodeDecryptionFailure, aerr.Error())
			case secretsmanager.ErrCodeInternalServiceError:
				fmt.Println(secretsmanager.ErrCodeInternalServiceError, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}

		return res, err
	}

	askey := result.SecretString

	// Declare ss as struct
	ss := SecretString{}

	//UnMarshal to get Access and Secret out of object
	json.Unmarshal([]byte(*askey), &ss)

	// Marshaling to json to pass back to API
	sss := map[string]string{"Access": ss.AccessKey, "Secret": ss.SecretKey}
	s, _ := json.Marshal(sss)

	res.ResponseMessage = fmt.Sprintln(string(s))

	return res, err

}
