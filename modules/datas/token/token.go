package token

import "sirherobrine23.com.br/go-bds/bds/modules/datas/permission"

const TokenSize int = 32

// Token
type Token interface {
	// Check token exists and have permission required,
	// if set [permission.Unknown] check only exists
	Check(token string) (exist bool, userID int, perm permission.Permission, err error)

	// Create token
	Create(userID int, perm permission.Permission) (token string, err error)
}
