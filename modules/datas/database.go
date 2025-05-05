package datas

import zfg "github.com/chaindead/zerocfg"

var (
	Driver *string = zfg.Str("db.type", "sqlite", "Database drive, example sqlite") // Database driver
	Host   *string = zfg.Str("db.host", "./bds.db", "Database connection")          // Host connection
	Name   *string = zfg.Str("db.name", "bds", "Database name")                     // Database name
)
