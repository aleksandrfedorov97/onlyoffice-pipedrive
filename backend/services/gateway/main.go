package main

import (
	"log"

	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/gateway/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		log.Fatalln(err)
	}
}
