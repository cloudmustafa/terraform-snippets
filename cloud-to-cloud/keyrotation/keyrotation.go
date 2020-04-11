package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
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

func main() {

	lambda.Start(LambdaHandler)

}

// LambdaHandler : AWS Lambda function handler
func LambdaHandler(parameters InputParameters) (ResponseData, error) {

	var res ResponseData
	iamUser := parameters.SubscriptionID

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Get the old Access Key
	// and call the deleteOldAccessKey function
	e := getOldAccessKey(sess, iamUser)
	if e != nil {
		fmt.Println(e.Error())
		return res, e
	}

	// Creates new Access Key
	newAccessKey, secretKey, e := createNewAccessKey(sess, iamUser)
	if e != nil {
		fmt.Println(e.Error())
		return res, e
	}
	// Update Secrets Manager with new SecretString
	updateSecretsManager(sess, newAccessKey, secretKey, iamUser)
	if e != nil {
		fmt.Println(e.Error())
		return res, e
	}

	sss := map[string]string{"Access": newAccessKey, "Secret": secretKey}
	s, _ := json.Marshal(sss)
	fmt.Println(string(s))

	res.ResponseMessage = fmt.Sprintln(string(s))

	return res, nil
}

// Retrieves the old Access Key for the IAM user
func getOldAccessKey(sess *session.Session, iamUser string) error {

	svc := iam.New(sess)

	input := &iam.ListAccessKeysInput{
		UserName: aws.String(iamUser),
	}

	result, err := svc.ListAccessKeys(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeNoSuchEntityException:
				fmt.Println(iam.ErrCodeNoSuchEntityException, aerr.Error())
			case iam.ErrCodeServiceFailureException:
				fmt.Println(iam.ErrCodeServiceFailureException, aerr.Error())
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

	var k *string
	for _, r := range result.AccessKeyMetadata {
		fmt.Println(r.AccessKeyId)
		k = r.AccessKeyId
	}
	fmt.Println(result)
	fmt.Println(aws.StringValue(k))
	keyResult := *k

	if e := deleteOldAccessKey(sess, keyResult, iamUser); e != nil {
		fmt.Println(e.Error())
		return e
	}

	return err

}

// Deletes the old Access Key for the IAM user
func deleteOldAccessKey(sess *session.Session, accessKey, iamUser string) error {

	svc := iam.New(sess)

	input := &iam.DeleteAccessKeyInput{
		AccessKeyId: aws.String(accessKey),
		UserName:    aws.String(iamUser),
	}

	result, err := svc.DeleteAccessKey(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeNoSuchEntityException:
				fmt.Println(iam.ErrCodeNoSuchEntityException, aerr.Error())
			case iam.ErrCodeLimitExceededException:
				fmt.Println(iam.ErrCodeLimitExceededException, aerr.Error())
			case iam.ErrCodeServiceFailureException:
				fmt.Println(iam.ErrCodeServiceFailureException, aerr.Error())
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

	fmt.Println(result)

	return nil
}

// Returns the access and secret key that is created for the IAM user
func createNewAccessKey(sess *session.Session, iamUser string) (string, string, error) {

	svc := iam.New(sess)
	input := &iam.CreateAccessKeyInput{
		UserName: aws.String(iamUser),
	}

	result, err := svc.CreateAccessKey(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeNoSuchEntityException:
				fmt.Println(iam.ErrCodeNoSuchEntityException, aerr.Error())
			case iam.ErrCodeLimitExceededException:
				fmt.Println(iam.ErrCodeLimitExceededException, aerr.Error())
			case iam.ErrCodeServiceFailureException:
				fmt.Println(iam.ErrCodeServiceFailureException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}

		return "", "", err
	}

	var akey = *result.AccessKey.AccessKeyId
	var skey = *result.AccessKey.SecretAccessKey

	fmt.Println(result)

	return akey, skey, err
}

// Updates Secrets Manager with the new Secret String including the Access and Secret Key
func updateSecretsManager(sess *session.Session, newAccessKey, secretKey, iamUser string) error {

	svc := secretsmanager.New(sess)

	//Environment
	var varEnvironment = os.Getenv("ENVIRONMENT")

	askey := fmt.Sprintf("{\"ACCESS\":\"%s\",\"SECRET\":\"%s\",\"Environment\":\"%s\"}", newAccessKey, secretKey, varEnvironment)

	input := &secretsmanager.UpdateSecretInput{
		SecretId:     aws.String(iamUser),
		SecretString: aws.String(askey),
	}

	result, err := svc.UpdateSecret(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case secretsmanager.ErrCodeInvalidParameterException:
				fmt.Println(secretsmanager.ErrCodeInvalidParameterException, aerr.Error())
			case secretsmanager.ErrCodeInvalidRequestException:
				fmt.Println(secretsmanager.ErrCodeInvalidRequestException, aerr.Error())
			case secretsmanager.ErrCodeLimitExceededException:
				fmt.Println(secretsmanager.ErrCodeLimitExceededException, aerr.Error())
			case secretsmanager.ErrCodeEncryptionFailure:
				fmt.Println(secretsmanager.ErrCodeEncryptionFailure, aerr.Error())
			case secretsmanager.ErrCodeResourceExistsException:
				fmt.Println(secretsmanager.ErrCodeResourceExistsException, aerr.Error())
			case secretsmanager.ErrCodeResourceNotFoundException:
				fmt.Println(secretsmanager.ErrCodeResourceNotFoundException, aerr.Error())
			case secretsmanager.ErrCodeMalformedPolicyDocumentException:
				fmt.Println(secretsmanager.ErrCodeMalformedPolicyDocumentException, aerr.Error())
			case secretsmanager.ErrCodeInternalServiceError:
				fmt.Println(secretsmanager.ErrCodeInternalServiceError, aerr.Error())
			case secretsmanager.ErrCodePreconditionNotMetException:
				fmt.Println(secretsmanager.ErrCodePreconditionNotMetException, aerr.Error())
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

	fmt.Println(result)

	return nil
}
