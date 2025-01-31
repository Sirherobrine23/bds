package db

import (
	"crypto/rand"
	"errors"
	"slices"

	"crypto/aes"
	"crypto/cipher"

	"golang.org/x/crypto/scrypt"

	"sirherobrine23.com.br/go-bds/bds/modules/config"
)

var (
	ErrTokenInvalid = errors.New("token invalid")
	ErrSaltPassword = errors.New("token not salted")
)

const (
	defaultSaltSize  = 24
	defaultTokenSize = 32
)

type Token struct {
	TokenID    int64  `json:"id" xorm:"'id' pk"`                    // Token ID
	User       User   `json:"user" xorm:"'user' extends"`           // User
	IsPassword bool   `json:"isPassword" xorm:"'ispass' default 0"` // Is Password not token
	Secret     []byte `json:"-" xorm:"'secret' notnull"`            // Password or token secret key
	IV         []byte `json:"-" xorm:"'iv' notnull"`                // Initialization Vector
	Cypther    []byte `json:"-" xorm:"'cypther' notnull"`           // Cypther
}

func createTokensDB() error {
	ok, err := DatabaseConnection.IsTableExist(&Token{})
	if err == nil && ok {
		err = DatabaseConnection.Sync(&Token{})
	} else if err == nil && !ok {
		err = DatabaseConnection.CreateTables(&Token{})
	}
	return err
}

// Save token in database
func (token Token) Save(user User) error {
	if token.IsPassword {
		table := DatabaseConnection.Table(&Token{})
		ok, err := table.Exist(&Token{User: user, IsPassword: true})
		if err == nil && ok {
			_, err = table.Where("user = ? AND ispass = ?", user, true).Update(&token) // Update password
		}
		return err
	}
	token.User = user
	_, err := DatabaseConnection.Insert(token)
	return err
}

func (token *Token) Salt() error {
	// Check if token is already salted
	if len(token.Secret) > 0 && len(token.IV) > 0 {
		return nil
	}

	section, err := config.ConfigProvider.GetSection("ENCRYPTS")
	if err != nil {
		return err
	}
	tokenToAuth := []byte(section.Key("AUTH_TOKEN").String())

	// Make secret and iv
	token.Secret, token.IV = make([]byte, defaultSaltSize), make([]byte, aes.BlockSize)
	if _, err = rand.Read(token.Secret); err != nil {
		return err
	} else if _, err = rand.Read(token.IV); err != nil {
		return err
	}

	// Encrypt token
	key, err := scrypt.Key(tokenToAuth, token.Secret, 24, 8, 1, 32)
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	// Encrypt secret
	encodedToken := make([]byte, len(token.Secret))
	stream := cipher.NewCBCEncrypter(block, token.IV)
	stream.CryptBlocks(encodedToken, token.Secret)
	token.Secret = encodedToken

	return nil
}

func (token Token) Compare(secret string) error {
	// Check if token is already salted
	if !(len(token.Secret) > 0 && len(token.IV) > 0) {
		return ErrSaltPassword
	}

	section, err := config.ConfigProvider.GetSection("ENCRYPTS")
	if err != nil {
		return err
	}
	tokenToAuth := []byte(section.Key("AUTH_TOKEN").String())

	// Decrypt token
	key, err := scrypt.Key(tokenToAuth, token.Secret, 24, 8, 1, 32)
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	// Decrypt secret
	decodedToken := make([]byte, len(token.Secret))
	stream := cipher.NewCBCDecrypter(block, token.IV)
	stream.CryptBlocks(decodedToken, token.Secret)

	// Compare secrets
	if string(decodedToken) != secret {
		return ErrTokenInvalid
	}

	return nil
}

// Create token from password string
func CreateTokenPassword(pass string) (*Token, error) {
	token := &Token{IsPassword: true, Secret: []byte(pass)}
	err := token.Salt()
	if err == nil {
		if err = token.Compare(pass); err != nil {
			token = nil
		}
	}
	return token, err
}

func CreateToken() (*Token, error) {
	token := &Token{
		Secret:     make([]byte, defaultTokenSize),
		IsPassword: false,
	}
	if _, err := rand.Read(token.Secret); err != nil {
		return nil, err
	}

	err := token.Salt()
	return token, err
}

func (token Token) String() string {
	return string(slices.Concat(token.IV, token.Secret, token.Cypther))
}
