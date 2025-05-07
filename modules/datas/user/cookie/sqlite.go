package cookie

import (
	"database/sql"
	_ "embed"
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

func SqliteCookie(conn *sql.DB) Cookie { return &Sqlite{conn} }

type Sqlite struct {
	Connection *sql.DB
}

func (sqlite *Sqlite) Cookie(cookieValue string) (exist bool, userID int, err error) {
	var cookies int
	if err = sqlite.Connection.QueryRow("SELECT count(*) FROM cookies WHERE cookie = $1", cookieValue).Scan(&cookies); err != nil {
		return
	}

	if cookies > 0 {
		exist = true
		err = sqlite.Connection.QueryRow("SELECT user_id FROM cookies WHERE cookie = $1 LIMIT 1", cookieValue).Scan(&userID)
	}

	return
}

func (sqlite *Sqlite) CreateCookie(userID int) (cookieValue string, err error) {
	cookieValue = newCookieValue()
	_, err = sqlite.Connection.Exec("INSERT INTO cookies(user_id, cookie) VALUES ($1, $2)", userID, cookieValue)
	return
}
