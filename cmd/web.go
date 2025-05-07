package cmd

import (
	"github.com/chaindead/zerocfg"
	"github.com/chaindead/zerocfg/env"
	"github.com/chaindead/zerocfg/yaml"
	"github.com/urfave/cli/v2"
	"sirherobrine23.com.br/go-bds/bds/modules/datas"
	httpserver "sirherobrine23.com.br/go-bds/bds/modules/http_server"
	"sirherobrine23.com.br/go-bds/bds/modules/web"
)

// Web subcommand
var Web = &cli.Command{
	Name:        "web",
	Description: "start interface dashboard",
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

		// Web interface
		httpWeb, err := web.MountRouter(&web.WebConfig{DatabaseSchemas: databaseConnection})
		if err != nil {
			return err
		}

		return httpserver.ListenAndServe(ctx.String("listen"), httpWeb)
	},
}
