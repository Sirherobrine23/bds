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
	// Create begin to make rollback if error
	tx, err := connection.Begin()
	if err != nil {
		return err
	}

	// Create table
	if _, err := tx.Exec(sqliteTableCreate); err != nil {
		tx.Rollback()
		return err
	}

	// commit
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func SqliteToken(conn *sql.DB) Token { return &sqliteToken{conn} }

type sqliteToken struct{ *sql.DB }

func (t *sqliteToken) Check(token string) (exist bool, userID int, perm permission.Permission, err error) {
	row := t.QueryRow("SELECT (user_id, permission) FROM token WHERE token = $1", token)
	if err = row.Err(); err == nil {
		err = row.Scan(&userID, &perm)
		exist = true
	}
	return
}

func (t *sqliteToken) Create(userID int, perm permission.Permission) (string, error) {
	tokenString := makeNewTokenValue()
	_, err := t.Exec("INSERT INTO tokens (user_id, token, permission) VALUE ($1, $2, $3)", userID, tokenString, perm)
	return tokenString, err
}
