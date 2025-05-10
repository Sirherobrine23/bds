package user

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/chaindead/zerocfg"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/encrypt"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/permission"
)

var (
	minPasswordLength = 8
	hasUpperRegex     = regexp.MustCompile(`[A-Z]`)
	hasSpecialRegex   = regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>/?~]`)
	usernameCheck     = regexp.MustCompile(`^[a-zA-Z0-9_.-]{3,20}$`)
	emailCheck        = regexp.MustCompile(`(?m)^(((((((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?([A-Za-z0-9!#-'*+\/=?^_\x60{|}~-])+((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?)|(((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?"((\s? +)?(([!#-[\]-~])|(\\([ -~]|\s))))*(\s? +)?"))?)?(((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?<(((((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?(([A-Za-z0-9!#-'*+\/=?^_\x60{|}~-])+(\.([A-Za-z0-9!#-'*+\/=?^_\x60{|}~-])+)*)((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?)|(((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?"((\s? +)?(([!#-[\]-~])|(\\([ -~]|\s))))*(\s? +)?"))@((((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?(([A-Za-z0-9!#-'*+\/=?^_\x60{|}~-])+(\.([A-Za-z0-9!#-'*+\/=?^_\x60{|}~-])+)*)((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?)|(((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?\[((\s? +)?([!-Z^-~]))*(\s? +)?\]((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?)))>((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?))|(((((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?(([A-Za-z0-9!#-'*+\/=?^_\x60{|}~-])+(\.([A-Za-z0-9!#-'*+\/=?^_\x60{|}~-])+)*)((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?)|(((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?"((\s? +)?(([!#-[\]-~])|(\\([ -~]|\s))))*(\s? +)?"))@((((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?(([A-Za-z0-9!#-'*+\/=?^_\x60{|}~-])+(\.([A-Za-z0-9!#-'*+\/=?^_\x60{|}~-])+)*)((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?)|(((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?\[((\s? +)?([!-Z^-~]))*(\s? +)?\]((((\s? +)?(\(((\s? +)?(([!-'*-[\]-~]*)|(\\([ -~]|\s))))*(\s? +)?\)))(\s? +)?)|(\s? +))?))))$`)

	EncryptKey *string = zerocfg.Str("encrypt.password", "", "Password to encrypt many secret values")
)

func ValidPasswordCombinedCheck(password string) bool {
	if len(password) < minPasswordLength {
		return false
	} else if !hasUpperRegex.MatchString(password) {
		return false
	} else if !hasSpecialRegex.MatchString(password) {
		return false
	}
	return true
}

func ValidUsername(username string) bool {
	return usernameCheck.MatchString(username)
}
func ValidEmail(username string) bool {
	return emailCheck.MatchString(username)
}

// User info
type User struct {
	ID         int64                 // Return unique user ID to reference in all points
	Name       string                // User name
	Username   string                // Username/nick name
	Permission permission.Permission // User permissions
	Password   *Password
}

// Password check and storage
type Password struct {
	ID           int64  // user ID
	PasswordHash string // Password hash

	DB     *sql.DB
	Driver string
}

// Check password is valid
func (pass *Password) Check(password string) (bool, error) {
	originalPassword, err := encrypt.Decrypt(*EncryptKey, pass.PasswordHash)
	if err != nil {
		return false, fmt.Errorf("cannot descrypt password: %s", err)
	}
	return password == originalPassword, nil
}

// Storage password
func (pass *Password) Storage(password string) error {
	password, err := encrypt.Encrypt(*EncryptKey, password)
	if err != nil {
		return fmt.Errorf("cannot encrypt password: %s", err)
	}

	switch pass.Driver {
	case "sqlite3":
		return pass.sqliteStorage(password)
	default:
		return fmt.Errorf("%s not supported yet", pass.Driver)
	}
}

type UserSearch struct {
	Driver string  // Driver name
	DB     *sql.DB // Database connection
}

func (users *UserSearch) Username(username string) (*User, error) {
	switch users.Driver {
	case "sqlite3":
		return users.sqliteUsername(username)
	default:
		return nil, fmt.Errorf("%s not supported yet", users.Driver)
	}
}

func (users *UserSearch) ByID(ID int64) (*User, error) {
	switch users.Driver {
	case "sqlite3":
		return users.sqliteByID(ID)
	default:
		return nil, fmt.Errorf("%s not supported yet", users.Driver)
	}
}

func (users *UserSearch) Create(name, username, email, password string) (*User, error) {
	name, username, email = strings.ToLower(name), strings.ToLower(username), strings.ToLower(email)
	if len(name) < 3 {
		return nil, fmt.Errorf("invalid name length")
	} else if !ValidUsername(username) {
		return nil, fmt.Errorf("invalid username")
	} else if !ValidEmail(email) {
		return nil, fmt.Errorf("invalid email")
	}

	var err error
	if password, err = encrypt.Encrypt(*EncryptKey, password); err != nil {
		return nil, err
	}

	switch users.Driver {
	case "sqlite3":
		return users.sqliteCreate(name, username, email, password)
	default:
		return nil, fmt.Errorf("%s not supported yet", users.Driver)
	}
}
