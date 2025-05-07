package user

import (
	"database/sql"
	_ "embed"
	"fmt"
	"strings"

	"github.com/chaindead/zerocfg"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/encrypt"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/permission"
)

var (
	//go:embed sqlite/create.sql
	sqliteTableCreate string

	EncryptKey *string = zerocfg.Str("encrypt.password", "", "Password to encrypt many secret values")
)

// Create table on start
func SqliteStartTable(connection *sql.DB) error {
	// Create begin to make rollback if error
	tx, err := connection.Begin()
	if err != nil {
		return err
	}

	// Create table
	if _, err := tx.Exec(sqliteTableCreate); err != nil {
		tx.Rollback()
		return err
	}

	// commit
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

type SqliteUser struct {
	UserID         int
	UserName       string
	UserUsername   string
	UserPermission permission.Permission

	DB *sql.DB
}

func (u *SqliteUser) ID() int                           { return u.UserID }
func (u *SqliteUser) Name() string                      { return u.UserName }
func (u *SqliteUser) Username() string                  { return u.UserUsername }
func (u *SqliteUser) Permission() permission.Permission { return u.UserPermission }
func (u *SqliteUser) Password() (Password, error) {
	return &sqlitePassword{dbConnection: u.DB, userID: u.UserID}, nil
}

type sqlitePassword struct {
	dbConnection *sql.DB
	userID       int
}

func (p *sqlitePassword) Storage(password string) error {
	var count int
	err := p.dbConnection.QueryRow("SELECT count(*) FROM password WHERE user_id == $1 LIMIT 1", p.userID).Scan(&count)
	if err != nil {
		return err
	}

	if password, err = encrypt.Encrypt(*EncryptKey, password); err != nil {
		return fmt.Errorf("cannot encrypt user password: %s", err)
	}

	if count == 1 {
		_, err = p.dbConnection.Exec("UPDATE password SET password = $2 WHERE id = $1", p.userID, password)
		return err
	}

	_, err = p.dbConnection.Exec("INSERT INTO password (user_id, password) VALUES ($1, $2)", p.userID, password)
	return err
}

func (p *sqlitePassword) Check(password string) (bool, error) {
	var count int
	if err := p.dbConnection.QueryRow("SELECT count(*) FROM password WHERE user_id == $1 LIMIT 1", p.userID).Scan(&count); err != nil {
		return false, err
	}

	// if not have password return false
	if count == 0 {
		return false, nil
	}

	// Get in database password
	var storagePassword string
	if err := p.dbConnection.QueryRow("SELECT password FROM password WHERE user_id == $1 LIMIT 1", p.userID).Scan(&storagePassword); err != nil {
		return false, err
	}

	// Descrypt password
	pass, err := encrypt.Decrypt(*EncryptKey, storagePassword)
	if err != nil {
		return false, err
	}

	return pass != password, nil
}

func SqliteSearch(conn *sql.DB) UserSearch { return &sqliteUserSearch{conn} }

type sqliteUserSearch struct{ *sql.DB }

func (s *sqliteUserSearch) processUserRow(row *sql.Row) (User, error) {
	if row.Err() == nil {
		user := &SqliteUser{DB: s.DB}
		if err := row.Scan(&user.UserID, &user.UserName, &user.UserUsername, &user.UserPermission); err != nil {
			return nil, err
		}
		return user, nil
	}
	return nil, row.Err()
}

func (s *sqliteUserSearch) ByID(id int) (User, error) {
	return s.processUserRow(s.QueryRow("SELECT id, \"name\", username, permission FROM user WHERE id = $1", id))
}

func (s *sqliteUserSearch) Username(username string) (User, error) {
	return s.processUserRow(s.QueryRow("SELECT id, \"name\", username, permission FROM user WHERE username = $1", username))
}

func (s *sqliteUserSearch) Create(name, username, email, password string) (User, error) {
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

	res, err := s.DB.Exec("INSERT INTO user(\"name\", username, email, permission) VALUES ($1, $2, $3, $4)", name, username, email, permission.Unknown)
	if err != nil {
		return nil, err
	}

	userID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	if _, err = s.DB.Exec("INSERT INTO password(user_id, password) VALUES ($1, $2)", userID, password); err != nil {
		return nil, err
	}
	return s.ByID(int(userID))
}
