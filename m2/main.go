package main

import (
	"context"
	"manage2/internal/commands"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name:      "m2",
		Usage:     "Manage game servers",
		UsageText: "m2 -g GAME COMMAND [OPTION...]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "log-level",
				Usage:   "Set log `LEVEL`: error, warn, info, debug",
				Aliases: []string{"l"},
				Value:   "info",
			},
			&cli.StringFlag{
				Name:     "game",
				Usage:    "Run command for `GAME`",
				Category: "Required options:",
				Required: true,
				Aliases:  []string{"g"},
			},
		},
		Commands: []*cli.Command{
			{
				Name:   "start",
				Usage:  "Start a server",
				Action: commands.Start,
			},
			{
				Name:   "daemon",
				Usage:  "Start a server daemon",
				Hidden: true,
				Action: commands.Daemon,
			},
		},
	}
	if err := app.Run(context.Background(), os.Args); err != nil {
		panic("Failed to run application: " + err.Error())
	}
}
