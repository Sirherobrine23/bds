package cookie

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
)

type Cookie struct {
	Driver string
	DB     *sql.DB
}

func (cookie *Cookie) Cookie(cookieValue string) (exist bool, userID int64, err error) {
	switch cookie.Driver {
	case "sqlite3":
		return cookie.sqliteCookie(cookieValue)
	default:
		return false, -1, fmt.Errorf("%s not supported yet", cookie.Driver)
	}
}

func (cookie *Cookie) CreateCookie(userID int64) (cookieValue string, err error) {
	cookieBytes := make([]byte, 12)
	rand.Read(cookieBytes)
	cookieBytes[0] = 'm'
	cookieBytes[1] = 'a'
	cookieBytes[2] = 'y'
	cookieBytes[3] = '1'
	cookieBytes[4] = '4'
	cookieValue = hex.EncodeToString(cookieBytes)

	switch cookie.Driver {
	case "sqlite3":
		err = cookie.sqliteInsertCookie(userID, cookieValue)
	default:
		return "", fmt.Errorf("%s not supported yet", cookie.Driver)
	}
	return
}
