package main

import (
	"log"

	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/auth/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		log.Fatalln(err)
	}
}
