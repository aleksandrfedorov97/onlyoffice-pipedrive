package main

import (
	"log"

	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/builder/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		log.Fatalln(err)
	}
}
