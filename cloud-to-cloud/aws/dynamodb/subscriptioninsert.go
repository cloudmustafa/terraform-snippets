package dynamodb

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// MetaData : Struct here to map to DynamoDB table attributes (struct in a struct)
type MetaData struct {
	HostName             string `json:"hostname"`
	DeviceID             string `json:"deviceid"`
	TopicName            string `json:"topicname"`
	ClientCertificate    string `json:"clientcertificate"`
	ClientKey            string `json:"clientkey"`
	CertificateAuthority string `json:"certificateauthority"`
	AccessKey            string `json:"accesskey"`
	Secret               string `json:"secret"`
	Location             string `json:"location"`
	Frequency            int    `json:"frequency"`
	PrivateKey           string `json:"privatekey"`
	ProjectID            string `json:"projectid"`
	Region               string `json:"region"`
	RegistryID           string `json:"registryid"`
	SasToken             string `json:"sastoken"`
	PolicyName           string `json:"policyname"`
}

// Asset : Struct here to map to DynamoDB table attributes (struct in a struct)
type Asset struct {
	AssetID int      `json:"assetid"`
	Name    string   `json:"name"`
	Tags    []string `json:"tags"`
}

// APISubscription : Struct here to map to DynamoDB table attributes (main struct)
type APISubscription struct {
	SubscriptionID     string   `json:"subscriptionID"`
	Name               string   `json:"name"`
	Assets             []Asset  `json:"assets"`
	Target             string   `json:"target"`
	MetaData           MetaData `json:"metadata"`
	StartTIme          int      `json:"starttime"`
	EndTime            int      `json:"endtime"`
	SubscriptionStatus string   `json:"subscriptionstatus"`
	SourceIP           string   `json:"sourceIP"`
	SourcePort         string   `json:"sourcePort"`
	DestinationIP      string   `json:"destinationIP"`
	DestinationPort    string   `json:"destinationPort"`
}

// InsertIntoDynamoDB : Function to load data into DynamoDB
func InsertIntoDynamoDB(sess *session.Session, stackID string, subscription APISubscription) error {

	// The table name in DynamoDB
	var tableName = os.Getenv("SUBSCRIPTION_METADATA_TABLE")

	subscription.SubscriptionID = stackID
	subscription.SubscriptionStatus = "Start"

	sub := subscription

	svc := dynamodb.New(sess)

	av, err := dynamodbattribute.MarshalMap(sub)

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	result, err := svc.PutItem(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				fmt.Println(dynamodb.ErrCodeConditionalCheckFailedException, aerr.Error())
			case dynamodb.ErrCodeProvisionedThroughputExceededException:
				fmt.Println(dynamodb.ErrCodeProvisionedThroughputExceededException, aerr.Error())
			case dynamodb.ErrCodeResourceNotFoundException:
				fmt.Println(dynamodb.ErrCodeResourceNotFoundException, aerr.Error())
			case dynamodb.ErrCodeItemCollectionSizeLimitExceededException:
				fmt.Println(dynamodb.ErrCodeItemCollectionSizeLimitExceededException, aerr.Error())
			case dynamodb.ErrCodeTransactionConflictException:
				fmt.Println(dynamodb.ErrCodeTransactionConflictException, aerr.Error())
			case dynamodb.ErrCodeRequestLimitExceeded:
				fmt.Println(dynamodb.ErrCodeRequestLimitExceeded, aerr.Error())
			case dynamodb.ErrCodeInternalServerError:
				fmt.Println(dynamodb.ErrCodeInternalServerError, aerr.Error())
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
