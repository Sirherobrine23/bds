package cmd

import (
	"net/http"

	"sirherobrine23.com.br/go-bds/bds/modules/api"
	"sirherobrine23.com.br/go-bds/bds/modules/datas"

	"github.com/chaindead/zerocfg"
	"github.com/chaindead/zerocfg/env"
	"github.com/chaindead/zerocfg/yaml"
	"github.com/urfave/cli/v2"
)

// Web subcommand
var API = &cli.Command{
	Name:        "api",
	Description: "start only api",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "listen",
			Value: ":3000",
			EnvVars: []string{
				"LISTEN",
				"HTTP_LISTEN",
			},
		},
	},
	Action: func(ctx *cli.Context) error {
		yamlConfig := new(string)
		*yamlConfig = ctx.String("config")
		if err := zerocfg.Parse(env.New(), yaml.New(yamlConfig)); err != nil {
			return err
		}

		// Start database connection
		databaseConnection, err := datas.Connect()
		if err != nil {
			return err
		}

		// Start http server
		httpRouter, err := api.MountRouter(&api.RouteConfig{Token: databaseConnection.Token, User: databaseConnection.User})
		if err != nil {
			return err
		}

		return http.ListenAndServe(ctx.String("listen"), httpRouter)
	},
}
