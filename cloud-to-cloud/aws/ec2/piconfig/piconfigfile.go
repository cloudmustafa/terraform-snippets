package piconfigfile

import (
	"fmt"
	"math/rand"
	"maxedge/cloud-to-cloud/aws/dynamodb"
	"os"
	"strconv"
)

// ResponseData : Data structure for returned data from Lambda function
type ResponseData struct {
	ResponseMessage string `json:"responseMesssage"`
}

// Random number generator for location1 of the file
func randomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

// CreatePiConfig : Creates the Pi file for the user to use to
// configure thier host instance
func CreatePiConfig(stackID string, subscription dynamodb.APISubscription) (ResponseData, error) {

	var res ResponseData

	// Location1 in the file must be an int and the code requires a string
	// Creating random number, and converting to string
	var location1 = strconv.Itoa(randomInt(100, 999999999))

	//writes to local file system
	f, _ := os.Create("C:/Users/allenw/Downloads/pitest1.txt")

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
			fmt.Fprintln(f, value+",,,"+stackID+",Float64,1,"+location1+",0,7,1,4")
		}
	}

	fmt.Fprintln(f, endOfFile)

	return res, nil
}
