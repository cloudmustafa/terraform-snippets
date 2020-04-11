package main

import (
	"encoding/json"
	"fmt"
)

func main() {

	// cpu := 512
	// mem := 1024

	testJSON := []byte{"tags":[
		"tag",
		"tag",
		"tag",
		"tag", 
		"tag",
		"tag", 
		"tag",
		"tag",
		"tag",
		]}

	type tags struct {
		Tags []string
	}

	t := tags{}

	json.Unmarshal([]byte(testJSON), &t)

	fmt.Println(t)

}
