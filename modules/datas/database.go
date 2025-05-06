package datas

import (
	"database/sql"
	"fmt"

	zfg "github.com/chaindead/zerocfg"

	"sirherobrine23.com.br/go-bds/bds/modules/datas/token"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/user"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/user/cookie"

	_ "sirherobrine23.com.br/go-bds/bds/modules/datas/internal/clients" // Load database clients
)

var (
	Driver *string = zfg.Str("db.type", "sqlite3", "Database drive, example sqlite") // Database driver
	Host   *string = zfg.Str("db.host", "./bds.db", "Database connection")           // Host connection
	Name   *string = zfg.Str("db.name", "bds", "Database name")                      // Database name
)

type DatabaseSchemas struct {
	Database *sql.DB // Drive connection

	User   user.UserSearch // User database
	Token  token.Token     // Token database
	Cookie cookie.Cookie   // Web cookies
}

func Connect() (*DatabaseSchemas, error) {
	driveSql, err := sql.Open(*Driver, *Host)
	if err != nil {
		return nil, err
	}

	switch *Driver {
	case "sqlite3":
		if err := user.SqliteStartTable(driveSql); err != nil {
			return nil, err
		}

		if err := cookie.SqliteStartTable(driveSql); err != nil {
			return nil, err
		}

		if err := token.SqliteStartTable(driveSql); err != nil {
			return nil, err
		}

		return &DatabaseSchemas{
			Database: driveSql,

			User:   user.SqliteSearch(driveSql),
			Token:  token.SqliteToken(driveSql),
			Cookie: cookie.SqliteCookieConn(driveSql),
		}, nil
	default:
		driveSql.Close()
		return nil, fmt.Errorf("database not supported by bds dashboard")
	}
}
