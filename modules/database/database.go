package db

import (
	"sirherobrine23.com.br/go-bds/bds/modules/config"
	"xorm.io/xorm"
	"xorm.io/xorm/log"
	"xorm.io/xorm/names"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

var DatabaseConnection *xorm.Engine

func ConnectDB() error {
	section, err := config.ConfigProvider.GetSection("database")
	if err != nil {
		if section, err = config.ConfigProvider.NewSection("database"); err != nil {
			return err
		}
		section.NewKey("DRIVER", "sqlite3")
		section.NewKey("CONNECTION", "./bds.db")
	}

	section.Key("DRIVER").MustString("sqlite3")
	section.Key("CONNECTION").MustString("./bds.db")
	if DatabaseConnection, err = xorm.NewEngine(section.Key("DRIVER").String(), section.Key("CONNECTION").String()); err != nil {
		return err
	}
	DatabaseConnection.SetMapper(names.SameMapper{})
	DatabaseConnection.ShowSQL(true)
	DatabaseConnection.Logger().SetLevel(log.LOG_DEBUG)

	// Create tables
	if err := DatabaseConnection.CreateTables(&User{}, &Token{}, &Cookie{}); err != nil {
		return err
	}
	return nil
}
