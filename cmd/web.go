package cmd

import (
	"github.com/urfave/cli/v2"
	db "sirherobrine23.com.br/go-bds/bds/modules/database"
	"sirherobrine23.com.br/go-bds/bds/routers"
)

var Web = &cli.Command{
	Name: "web",
	Action: func(ctx *cli.Context) error {
		if err := db.ConnectDB(); err != nil {
			return err
		}
		defer db.DatabaseConnection.Close() // Close database connection
		return routers.Listen()
	},
}
