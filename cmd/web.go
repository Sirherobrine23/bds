package cmd

import "github.com/urfave/cli/v2"

// Web subcommand
var Web = &cli.Command{
	Name:        "web",
	Description: "start interface dashboard",
	Action: func(ctx *cli.Context) error {
		return nil
	},
}
