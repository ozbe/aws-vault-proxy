package main

import (
	"fmt"
	"log"
)

func main() {
	var mfa string
	fmt.Println("Enter token for arn:aws:iam::ACCOUNTID:mfa/USER:")
	_, err := fmt.Scanf("%s", &mfa)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("AWS_SUCCESS=true")
}
