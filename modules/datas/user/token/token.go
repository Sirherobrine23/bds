package token

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"

	"sirherobrine23.com.br/go-bds/bds/modules/datas/permission"
)

const TokenSize int = 32

// Token
type Token struct {
	Driver string
	DB     *sql.DB
}

// Check token exists and have permission required,
// if set [permission.Unknown] check only exists
func (tk *Token) Check(token string) (exist bool, userID int64, perm permission.Permission, err error) {
	switch tk.Driver {
	case "sqlite3":
		return tk.sqliteCheck(token)
	default:
		return false, -1, -1, fmt.Errorf("%s not supported yet", tk.Driver)
	}
}

// Create token
func (tk *Token) Create(userID int, perm permission.Permission) (token string, err error) {
	tokenBuff := make([]byte, TokenSize)
	rand.Read(tokenBuff)
	tokenBuff[0] = 'b'
	tokenBuff[1] = 's'
	tokenBuff[2] = 'd'
	token = hex.EncodeToString(tokenBuff)

	switch tk.Driver {
	case "sqlite3":
		err = tk.sqliteCreate(userID, token, perm)
	default:
		return "", fmt.Errorf("%s not supported yet", tk.Driver)
	}
	return
}
