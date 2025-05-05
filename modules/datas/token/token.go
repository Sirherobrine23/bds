package token

import "sirherobrine23.com.br/go-bds/bds/modules/datas/permission"

// Token
type Token interface {
	// Check token exists and have permission required,
	// if set [permission.Unknown] check only exists
	Check(token string, perm permission.Permission) (exist bool, userID int, err error)

	// Create token
	Create(userID int) (token string, err error)
}
