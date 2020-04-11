package accesskeysecretkey

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

// ResponseData : Data structure for returned data from Lambda function
type ResponseData struct {
	ResponseMessage string `json:"responseMesssage"`
}

// CreateAccessSecretKey : Creates the Access Key and Secret key
func CreateAccessSecretKey(sess *session.Session, subID string) (string, string, error) {

	// Connection information
	svc := iam.New(sess)

	u := &iam.GetUserInput{
		UserName: aws.String(subID),
	}

	input := &iam.CreateAccessKeyInput{
		UserName: aws.String(subID),
	}

	// Waits until the IAM user exists before continuing
	if e := svc.WaitUntilUserExists(u); e != nil {
		fmt.Println(e.Error())
		return "", "", e
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

	}

	// Assigning data to variable to be returned to calling function
	var akey = *result.AccessKey.AccessKeyId
	var skey = *result.AccessKey.SecretAccessKey

	fmt.Println(result)
	return akey, skey, err
}
