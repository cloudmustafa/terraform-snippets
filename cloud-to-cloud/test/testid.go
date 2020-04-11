package main

import (
	"fmt"

	uuid "github.com/satori/go.uuid"
)

func main() {

	fmt.Println(generateStackID())
}
func generateStackID() string {
	return `MaxEdge-` + uuid.Must(uuid.NewV4()).String()
}
