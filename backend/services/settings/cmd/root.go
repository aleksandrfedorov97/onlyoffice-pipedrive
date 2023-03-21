package cmd

import (
	"os"

	"github.com/urfave/cli/v2"
)

func GetCommands() cli.Commands {
	return []*cli.Command{
		Server(),
	}
}

func Run() error {
	app := &cli.App{
		Name:        "onlyoffice:docserver",
		Description: "Description",
		Authors: []*cli.Author{
			{
				Name:  "Ascensio Systems SIA",
				Email: "support@onlyoffice.com",
			},
		},
		HideVersion: true,
		Commands:    GetCommands(),
	}

	return app.Run(os.Args)
}
