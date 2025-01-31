package db

import "errors"

var ErrUserExist error = errors.New("user not exists")

// Users slice
type Users []*User

// Base struct to User
type User struct {
	ID       int64  `json:"id" xorm:"pk"`                               // User id
	Username string `json:"username" xorm:"varcha(125) notnull unique"` // Username
	Email    string `json:"email" xorm:"text notnull unique"`           // User email
	Name     string `json:"name" xorm:"text notnull"`                   // User name
	Active   bool   `json:"actived" xorm:"bool default 0"`              // If account is actived
	Banned   bool   `json:"banned" xorm:"bool default 0"`               // If account not regular to access
}

func createUserDB() error {
	ok, err := DatabaseConnection.IsTableExist(&User{})
	if err == nil && ok {
		err = DatabaseConnection.Sync(&User{})
	} else if err == nil && !ok {
		err = DatabaseConnection.CreateTables(&User{})
	}
	return err
}

// Load all users to Struct
func (users *Users) ListAll() error {
	// Get table from user struct
	table := DatabaseConnection.Table(&User{})
	defer table.Close()
	return table.Find(*users)
}

// Insert user to table
func (user *User) CreateUser() error {
	// Get table from user struct
	table := DatabaseConnection.Table(&User{})
	defer table.Close()

	// Check if exists username or email
	if ok, err := table.Exist(&User{Username: user.Username}, &User{Email: user.Email}); err != nil {
		return err
	} else if ok {
		return ErrUserExist
	}

	// Clean ID
	user.ID = 0
	// Insert to table
	_, err := table.Insert(user)
	return err
}
