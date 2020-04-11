package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

// MetaData : The stack ID is the random generated ID that is used as the ID and Name for subscription resources
// and is passed in from the UI. Other than the StackID, are the values needed for the Pi interface server
// type MetaData struct {
// 	HostName             string `json:"hostname"`
// 	DeviceID             string `json:"deviceid"`
// 	TopicName            string `json:"topicname"`
// 	ClientCertificate    string `json:"clientcertificate"`
// 	ClientKey            string `json:"clientkey"`
// 	CertificateAuthority string `json:"certificateauthority"`
// 	AccessKey            string `json:"accesskey"`
// 	Secret               string `json:"secret"`
// 	Location             string `json:"location"`
// 	Frequency            int    `json:"frequency"`
// 	PrivateKey           string `json:"privatekey"`
// 	ProjectID            string `json:"projectid"`
// 	Region               string `json:"region"`
// 	RegistryID           string `json:"registryid"`
// 	SasToken             string `json:"sastoken"`
// 	PolicyName           string `json:"policyname"`
// }

// Asset : Struct here to map to DynamoDB table attributes (struct in a struct)
type Asset struct {
	AssetID int      `json:"assetid"`
	Name    string   `json:"name"`
	Tags    []string `json:"tags"`
}

// APISubscription : Struct here to map to DynamoDB table attributes (main struct)
type APISubscription struct {
	SubscriptionID string `json:"subscriptionID"`
	//	Name               string   `json:"name"`
	Assets []Asset `json:"assets"`
	//	Target             string   `json:"target"`
	//	MetaData           MetaData `json:"metadata"`
	//	StartTIme          int      `json:"starttime"`
	//	EndTime            int      `json:"endtime"`
	// SubscriptionStatus string `json:"subscriptionstatus"`
	// SourceIP           string `json:"sourceIP"`
	// SourcePort         string `json:"sourcePort"`
	// DestinationIP      string `json:"destinationIP"`
	// DestinationPort    string `json:"destinationPort"`
	Location1 string `json:"location1"`
}

// ResponseData : Data structure for returned data from Lambda function
type ResponseData struct {
	ResponseMessage string `json:"responseMesssage"`
}

var tableName = os.Getenv("SUBSCRIPTION_METADATA_TABLE")

// Main function reqired by Lambda
func main() {

	// Test code
	// t := Asset{
	// 	Tags: []string{"tag1", "tag2", "tag3", "tag444", "tag5555", "tag6666", "endTag119"},
	// }

	// p := APISubscription{
	// 	SubscriptionID:  "MaxEdge-e2c0f3f6-d28a-400f-9108-26be5ed342b8",
	// 	SourceIP:        "EC2AMAZ-5EEVEUI",
	// 	SourcePort:      "5450",
	// 	DestinationIP:   "10.0.4.99",
	// 	DestinationPort: "5450",
	// 	Assets:          []Asset{t},
	// }

	// LambdaHandler(p)

	lambda.Start(LambdaHandler)
}

// LambdaHandler : AWS Lambda function handler
func LambdaHandler(subscription APISubscription) (ResponseData, error) {

	var res ResponseData
	// var id = subscription.SubscriptionID

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Inserts data into the 'subscriptions' table
	if e := getLocation1(sess, subscription); e != nil {
		fmt.Printf("Got error inserting data: %s", e.Error())
		return res, e
	}

	return res, nil
}

func getLocation1(sess *session.Session, subscription APISubscription) error {

	svc := dynamodb.New(sess)

	// The table name in DynamoDB
	// var tableName = "subscriptions"

	// Getting the location from the database
	// because its need to add to the new file
	loc1 := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"subscriptionID": {
				S: aws.String(subscription.SubscriptionID),
			},
		},
		TableName: aws.String(tableName),
	}

	locationResult, err := svc.GetItem(loc1)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeProvisionedThroughputExceededException:
				fmt.Println(dynamodb.ErrCodeProvisionedThroughputExceededException, aerr.Error())
			case dynamodb.ErrCodeResourceNotFoundException:
				fmt.Println(dynamodb.ErrCodeResourceNotFoundException, aerr.Error())
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

	item := APISubscription{}

	err = dynamodbattribute.UnmarshalMap(locationResult.Item, &item)
	if err != nil {
		panic(fmt.Sprintf("Error,  %v", err))
	}

	fmt.Println(item.Location1)

	// Inserts data into the 'subscriptions' table
	if e := updateMetaData(sess, subscription); e != nil {
		fmt.Printf("Got error inserting data: %s", e.Error())
		return e
	}

	// Creates the Pi output file
	_, e := updatePiTags(subscription.SubscriptionID, item.Location1, subscription)
	if e != nil {
		fmt.Println(e.Error())
		return e
	}

	return err
}

func updateMetaData(sess *session.Session, subscription APISubscription) error {

	svc := dynamodb.New(sess)

	// The table name in DynamoDB
	//var tableName = os.Getenv("SUBSCRIPTION_METADATA_TABLE")
	// var tableName = "subscriptions"

	// Updating tags
	update := expression.Set(
		expression.Name("assets"),
		expression.Value(subscription.Assets),
	)

	expr, err := expression.NewBuilder().
		WithUpdate(update).
		Build()

	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		Key: map[string]*dynamodb.AttributeValue{
			"subscriptionID": {
				S: aws.String(subscription.SubscriptionID),
			},
		},
		TableName:        aws.String(tableName),
		UpdateExpression: expr.Update(),
	}

	results, err := svc.UpdateItem(input) //Updating the item
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

	fmt.Println(subscription.Location1)
	fmt.Println(results)
	return err
}

// CreatePiConfig : Creates the Pi file for the user to use to
// configure thier host instance
func updatePiTags(stackID, loc1 string, subscription APISubscription) (ResponseData, error) {

	var res ResponseData

	// Location1 in the file must be an int and the code requires a string
	// Creating random number, and converting to string
	//var location1 = strconv.Itoa(randomInt(100, 999999999))

	//writes to local file system
	f, _ := os.Create("C:/Users/allenw/Downloads/updatePiTag.txt")

	//f, _ := os.Create("/tmp/pitest1.txt")

	defer f.Close()

	// Left justified file header information
	// File has a empty line at the top that does not need to be removed
	file := `
@table pipoint
@ptclass classic
@mode create, t
@mode edit, t
@istr tag,instrumenttag,descriptor,pointsource,pointtype,compressing,location1,location2,location3,location4,location5`

	// Left justified end of file information
	endOfFile := `
@ends
@exit`

	//Writing data to file then looping and adding tags and meta data on a new line per tag
	fmt.Fprintln(f, file)
	for idx := range subscription.Assets {
		for _, value := range subscription.Assets[idx].Tags {
			fmt.Fprintln(f, value+",,,"+stackID+",Float64,1,"+loc1+",0,7,1,4")
			//	fmt.Fprintln(f, value+",,,"+stackID+",Float64,1,"+value+",0,7,1,4")
		}
	}

	fmt.Fprintln(f, endOfFile)

	return res, nil
}
