package cookie

import (
	"database/sql"
	_ "embed"
)

//go:embed sql/sqlite_create.sql
var sqliteTableCreate string

// Create table on start
func CreateSqliteTable(connection *sql.DB) error {
	_, err := connection.Exec(sqliteTableCreate)
	return err
}

func (sqlite *Cookie) sqliteInsertCookie(userID int64, cookieValue string) error {
	_, err := sqlite.DB.Exec("INSERT INTO cookies(user_id, cookie) VALUES ($1, $2)", userID, cookieValue)
	return err
}

func (sqlite *Cookie) sqliteCookie(cookieValue string) (exist bool, userID int64, err error) {
	var cookies int
	if err = sqlite.DB.QueryRow("SELECT count(*) FROM cookies WHERE cookie = $1", cookieValue).Scan(&cookies); err != nil {
		return
	}

	if cookies > 0 {
		exist = true
		err = sqlite.DB.QueryRow("SELECT user_id FROM cookies WHERE cookie = $1 LIMIT 1", cookieValue).Scan(&userID)
	}

	return
}
