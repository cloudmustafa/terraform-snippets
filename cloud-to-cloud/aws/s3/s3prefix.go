package s3prefix

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// ResponseData : Data structure for returned data from Lambda function
type ResponseData struct {
	ResponseMessage string `json:"responseMesssage"`
}

// SubscriptionS3Folder : Creating S3 prefix (folder) which is the STACK ID
func SubscriptionS3Folder(sess *session.Session, prefix string) error {

	var s3Bucket = os.Getenv("S3_BUCKET")
	svc := s3.New(sess)

	s3p := fmt.Sprintf("/%s/", prefix)

	input := &s3.PutObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(s3p),
	}

	result, err := svc.PutObject(input)
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
	fmt.Println(result)
	return err

}
