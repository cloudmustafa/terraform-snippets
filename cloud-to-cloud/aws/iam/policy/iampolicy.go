package iampolicy

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

// PutUserPolicy : Adds policy to user for read/write access to the S3 prefix
// created for the user
func PutUserPolicy(subUser string) error {

	svc := iam.New(session.New())

	input := &iam.PutUserPolicyInput{
		PolicyDocument: aws.String("{\"Version\":\"2012-10-17\",\"Statement\":[{\"Sid\":\"VisualEditor0\",\"Effect\":\"Allow\",\"Action\":[\"s3:GetAccessPoint\",\"s3:PutAccountPublicAccessBlock\",\"s3:GetAccountPublicAccessBlock\",\"s3:ListAllMyBuckets\",\"s3:ListAccessPoints\",\"s3:ListJobs\",\"s3:CreateJob\",\"s3:HeadBucket\"],\"Resource\":\"*\"},{\"Sid\":\"VisualEditor1\",\"Effect\":\"Allow\",\"Action\":\"s3:*\",\"Resource\":\"arn:aws:s3:::maxedgecloudtocloudpoc-sandbox/subfolder1\"}]}"),
		PolicyName:     aws.String("AllAccessPolicy"),
		UserName:       aws.String(subUser),
	}

	result, err := svc.PutUserPolicy(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeLimitExceededException:
				fmt.Println(iam.ErrCodeLimitExceededException, aerr.Error())
			case iam.ErrCodeMalformedPolicyDocumentException:
				fmt.Println(iam.ErrCodeMalformedPolicyDocumentException, aerr.Error())
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

	fmt.Println(result)
	return err
}
