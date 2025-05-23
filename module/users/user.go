// User information and authenticantion
package users

import (
	"fmt"
	"time"

	"sirherobrine23.com.br/go-bds/bds/module/encrypt"
)

// User representation
type User struct {
	UserID   int64     `json:"id"`        // User ID to make linkers
	Username string    `json:"username"`  // Username
	Name     string    `json:"name"`      // Name to show
	Email    string    `json:"email"`     // user email to check unique value
	CreateAt time.Time `json:"create_at"` // Date of user creation
	UpdateAt time.Time `json:"update_at"` // Date to update any row in database
}

// Password storage
type Password struct {
	UserID   int64     `json:"id"`        // User ID, foregin key
	UpdateAt time.Time `json:"update_at"` // Data password update
	Password string    `json:"password"`  // Password hash
}

// Cookie storage to web
type Cookie struct {
	ID       int64     `json:"id"`        // Cookie id
	User     int64     `json:"user_id"`   // User ID
	Cookie   string    `json:"cookie"`    // cookie value
	CreateAt time.Time `json:"create_at"` // time creation
}

// Token to auth API router
type Token struct {
	ID          int64            `json:"id"`          // Cookie id
	User        int64            `json:"user_id"`     // User ID
	Token       string           `json:"token"`       // Token value in hex code
	Permissions TokenPermissions `json:"permissions"` // Token permission
	CreateAt    time.Time        `json:"create_at"`   // time creation
	UpdateAt    time.Time        `json:"update_at"`   // Date to update any row in database
}

// Convert plain key to hash encrypted key
func (pass *Password) HashPassword(encryptKey string) error {
	newKey, err := encrypt.Encrypt(pass.Password, encryptKey)
	if err != nil {
		return fmt.Errorf("cannot hash plain password: %s", err)
	}
	pass.Password = newKey
	return nil
}

// Descrypt password and check if is same
func (pass Password) Check(password, encryptKey string) (bool, error) {
	plainPassword, err := encrypt.Decrypt(encryptKey, pass.Password)
	if err != nil {
		return false, err
	}
	return plainPassword == password, nil
}
