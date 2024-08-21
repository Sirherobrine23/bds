package cmd

import (
	"github.com/urfave/cli/v2"
	"sirherobrine23.com.br/go-bds/bds/routers"
)

var Web = &cli.Command{
	Name: "web",
	Action: func(ctx *cli.Context) error {
		return routers.Listen()
	},
}
