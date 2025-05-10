package token

import (
	"database/sql"
	_ "embed"

	"sirherobrine23.com.br/go-bds/bds/modules/datas/permission"
)

var (
	//go:embed sqlite/create.sql
	sqliteTableCreate string
)

// Create table on start
func SqliteStartTable(connection *sql.DB) error {
	_, err := connection.Exec(sqliteTableCreate)
	return err
}

func (t *Token) sqliteCheck(token string) (exist bool, userID int64, perm permission.Permission, err error) {
	row := t.DB.QueryRow("SELECT (user_id, permission) FROM token WHERE token = $1", token)
	if err = row.Err(); err == nil {
		err = row.Scan(&userID, &perm)
		exist = true
	}
	return
}

func (t *Token) sqliteCreate(userID int, token string, perm permission.Permission) error {
	_, err := t.DB.Exec("INSERT INTO tokens (user_id, token, permission) VALUE ($1, $2, $3)", userID, token, perm)
	return err
}
