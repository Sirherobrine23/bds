package datas

import (
	"database/sql"
	"fmt"

	zfg "github.com/chaindead/zerocfg"

	"sirherobrine23.com.br/go-bds/bds/modules/datas/token"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/user"

	_ "sirherobrine23.com.br/go-bds/bds/modules/datas/internal/clients" // Load database clients
)

var (
	Driver *string = zfg.Str("db.type", "sqlite3", "Database drive, example sqlite") // Database driver
	Host   *string = zfg.Str("db.host", "./bds.db", "Database connection")           // Host connection
	Name   *string = zfg.Str("db.name", "bds", "Database name")                      // Database name

	EncryptKey *string = zfg.Str("encrypt.password", "", "Password to encrypt many secret values")
)

type DatabaseSchemas struct {
	Database *sql.DB

	User  user.UserSearch // User database
	Token token.Token     // Token database
}

func Connect() (*DatabaseSchemas, error) {
	sqlOpen, err := sql.Open(*Driver, *Host)
	if err != nil {
		return nil, err
	}

	switch *Driver {
	case "sqlite3":
		if err := sqliteStartTables(sqlOpen); err != nil {
			return nil, err
		}

		// return mount
		return sqliteMount(sqlOpen)
	default:
		sqlOpen.Close()
		return nil, fmt.Errorf("database not supported")
	}
}
