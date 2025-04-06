package db

import (
	"errors"
	"fmt"

	"sirherobrine23.com.br/go-bds/bds/modules/config"
	"sirherobrine23.com.br/go-bds/bds/modules/pass"
)

var (
	ErrInvalidPassword error = errors.New("invalid password")
)

type Token struct {
	TokenID  int64  `json:"id" xorm:"'id' pk autoincr"`    // Token ID
	User     *User  `json:"user" xorm:"'user_id' notnull"` // User
	Token    string `json:"-" xorm:"token"`                // Token secret
	Password string `json:"-" xorm:"password"`             // Password
}

func (token Token) Compare(password string) error {
	if token.Password == "" {
		return ErrInvalidPassword
	}
	section, err := config.ConfigProvider.GetSection("ENCRYPTS")
	if err != nil {
		return err
	}

	decodedPass, err := pass.Decrypt(section.Key("AUTH_TOKEN").String(), token.Password)
	if err != nil {
		return fmt.Errorf("cannot check password: %s", err)
	} else if decodedPass == password {
		return nil
	}
	return ErrInvalidPassword
}

func CreatePassword(password string, user *User) (*Token, error) {
	section, err := config.ConfigProvider.GetSection("ENCRYPTS")
	if err != nil {
		return nil, err
	} else if user == nil {
		return nil, ErrUserExist
	}

	// Create struct
	newToken := &Token{User: user, TokenID: 0, Token: ""}

	// Encrypt password
	if newToken.Password, err = pass.Encrypt(section.Key("AUTH_TOKEN").String(), password); err != nil {
		return nil, fmt.Errorf("cannot encrypt password: %s", err)
	} else if _, err = DatabaseConnection.InsertOne(newToken); err != nil {
		return nil, fmt.Errorf("cannot insert to database: %s", err)
	}
	
	return newToken, nil
}
