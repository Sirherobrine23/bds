package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"sirherobrine23.com.br/go-bds/bds/cmd"
	"sirherobrine23.com.br/go-bds/bds/modules"
	"sirherobrine23.com.br/go-bds/bds/modules/config"
)

func main() {
	app := cli.NewApp()
	app.HideHelpCommand = true
	app.HideVersion = true
	app.Name = "bds-dashboard"
	app.Usage = "Manager many Minecraft servers with one command or dashboard"
	app.Version = modules.AppVersion

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "config",
			Value: "config.ini",
			Usage: "config file path",
			Aliases: []string{
				"settigs",
				"c",
			},
		},
	}

	// Init process
	app.Before = func(ctx *cli.Context) error {
		err := config.LoadConfig(ctx.String("config"))
		return err
	}

	app.Commands = cmd.Subcomands

	// Start process
	if err := app.Run(os.Args); err != nil {
		switch value := err.(type) {
		case cli.ExitCoder:
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(value.ExitCode())
		default:
			fmt.Fprintln(os.Stderr, value.Error())
			os.Exit(1)
		}
	}
}
