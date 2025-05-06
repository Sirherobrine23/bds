package token

import (
	"crypto/rand"
	"encoding/hex"

	"sirherobrine23.com.br/go-bds/bds/modules/datas/permission"
)

const TokenSize int = 32

// Token
type Token interface {
	// Check token exists and have permission required,
	// if set [permission.Unknown] check only exists
	Check(token string) (exist bool, userID int, perm permission.Permission, err error)

	// Create token
	Create(userID int, perm permission.Permission) (token string, err error)
}

func makeNewTokenValue() string {
	tokenBuff := make([]byte, TokenSize)
	rand.Read(tokenBuff)
	tokenBuff[0] = 'b'
	tokenBuff[1] = 's'
	tokenBuff[2] = 'd'
	return hex.EncodeToString(tokenBuff)
}
