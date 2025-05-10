package datas

import (
	"database/sql"
	"fmt"

	zfg "github.com/chaindead/zerocfg"

	"sirherobrine23.com.br/go-bds/bds/modules/datas/server"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/user"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/user/cookie"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/user/token"

	_ "sirherobrine23.com.br/go-bds/bds/modules/datas/internal/sqlclients" // Load database clients
)

var (
	Driver *string = zfg.Str("db.type", "sqlite3", "Database drive, example sqlite") // Database driver
	Host   *string = zfg.Str("db.host", "./bds.db", "Database connection")           // Host connection
	Name   *string = zfg.Str("db.name", "bds", "Database name")                      // Database name
)

type DatabaseSchemas struct {
	Database *sql.DB // Drive connection

	User    *user.UserSearch   // User database
	Token   *token.Token       // Token database
	Cookie  *cookie.Cookie     // Web cookies
	Servers *server.ServerList // Servers
}

func Connect() (*DatabaseSchemas, error) {
	db, err := sql.Open(*Driver, *Host)
	if err != nil {
		return nil, fmt.Errorf("cannot open database: %s", err)
	}

	rootDB := &DatabaseSchemas{
		Database: db,

		User:    &user.UserSearch{Driver: *Driver, DB: db},
		Token:   &token.Token{Driver: *Driver, DB: db},
		Cookie:  &cookie.Cookie{Driver: *Driver, DB: db},
		Servers: &server.ServerList{Driver: *Driver, DB: db},
	}

	switch *Driver {
	case "sqlite3":
		if err = user.CreateSqliteTable(db); err != nil {
			return nil, err
		} else if err = token.SqliteStartTable(db); err != nil {
			return nil, err
		} else if err = cookie.CreateSqliteTable(db); err != nil {
			return nil, err
		} else if err = server.CreateSqliteTable(db); err != nil {
			return nil, err
		}
	case "postgres":
		fallthrough
	case "mysql":
		fallthrough
	case "mssql":
		fallthrough
	case "sqlserver":
		fallthrough
	default:
		db.Close()
		return nil, fmt.Errorf("database not supported by bds dashboard")
	}

	return rootDB, nil
}
