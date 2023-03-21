package main

import (
	"log"

	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/callback/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		log.Fatalln(err)
	}
}
