package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if os.Args[2] == "mfa" {
		var mfa string
		fmt.Print("Enter token for arn:aws:iam::1234564789:mfa/USER: ")
		_, err := fmt.Scanln(&mfa)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("ENV=debug\nAWS_SUCCESS=true")
}
