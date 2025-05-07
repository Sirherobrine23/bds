package datas

import (
	"database/sql"
	"fmt"

	zfg "github.com/chaindead/zerocfg"

	"sirherobrine23.com.br/go-bds/bds/modules/datas/server"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/token"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/user"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/user/cookie"

	_ "sirherobrine23.com.br/go-bds/bds/modules/datas/internal/sqlclients" // Load database clients
)

var (
	Driver *string = zfg.Str("db.type", "sqlite3", "Database drive, example sqlite") // Database driver
	Host   *string = zfg.Str("db.host", "./bds.db", "Database connection")           // Host connection
	Name   *string = zfg.Str("db.name", "bds", "Database name")                      // Database name
)

type DatabaseSchemas struct {
	Database *sql.DB // Drive connection

	User    user.UserSearch   // User database
	Token   token.Token       // Token database
	Cookie  cookie.Cookie     // Web cookies
	Servers server.ServerList // Servers
}

func Connect() (dbs *DatabaseSchemas, err error) {
	dbs = &DatabaseSchemas{}
	if dbs.Database, err = sql.Open(*Driver, *Host); err != nil {
		return nil, err
	}

	switch *Driver {
	case "sqlite3":
		if err = user.SqliteStartTable(dbs.Database); err != nil {
			return nil, err
		} else if err = cookie.SqliteStartTable(dbs.Database); err != nil {
			return nil, err
		} else if err = token.SqliteStartTable(dbs.Database); err != nil {
			return nil, err
		} else if err = server.SqliteStartTable(dbs.Database); err != nil {
			return nil, err
		}
		dbs.User, dbs.Token, dbs.Cookie = user.SqliteSearch(dbs.Database), token.SqliteToken(dbs.Database), cookie.SqliteCookie(dbs.Database)
		return
	case "postgres":
	case "mysql":
	case "mssql":
	case "sqlserver":
	}

	dbs.Database.Close()
	return nil, fmt.Errorf("database not supported by bds dashboard")
}
