package datas

import (
	"context"
	"crypto/rand"
	"database/sql"
	"embed"
	"encoding/hex"
	"fmt"

	"sirherobrine23.com.br/go-bds/bds/modules/datas/encrypt"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/permission"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/token"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/user"
)

//go:embed sql/sqlite/*
var sqliteSqls embed.FS

func sqliteStartTables(connection *sql.DB) error {
	createTableSql, err := sqliteSqls.ReadFile("sql/sqlite/create.sql")
	if err != nil {
		return fmt.Errorf("cannot open sql to create table: %s", err)
	}

	ctx := context.Background()
	tx, err := connection.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("cannot get begin of table creations: %s", err)
	}

	// Create tables in database
	if _, err := tx.Exec(string(createTableSql)); err != nil {
		tx.Rollback()
		return fmt.Errorf("cannot make tables in database: %s", err)
	}

	// Commit data
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return fmt.Errorf("cannot commit table creation: %s", err)
	}

	return nil
}

func sqliteMount(connection *sql.DB) (*DatabaseSchemas, error) {
	dbData := &DatabaseSchemas{Database: connection}

	dbData.User = &sqliteUserSearch{connection}
	dbData.Token = &sqliteToken{connection}

	return dbData, nil
}

type sqliteUserSearch struct{ *sql.DB }

func (s *sqliteUserSearch) processUserRow(row *sql.Row) (user.User, error) {
	if row.Err() == nil {
		user := &sqliteUser{dbConnection: s.DB}
		if err := row.Scan(&user.userID, &user.userName, &user.userUsername, &user.userPermission); err != nil {
			return nil, err
		}
		return user, nil
	}
	return nil, row.Err()
}

func (s *sqliteUserSearch) ByID(id int) (user.User, error) {
	return s.processUserRow(s.QueryRow("SELECT (id, \"name\", username, permission) FROM user WHERE id = $1", id))
}

func (s *sqliteUserSearch) Username(username string) (user.User, error) {
	return s.processUserRow(s.QueryRow("SELECT (id, \"name\", username, permission) FROM user WHERE username = $1", username))
}

type sqliteToken struct{ *sql.DB }

func (t *sqliteToken) Check(token string) (exist bool, userID int, perm permission.Permission, err error) {
	row := t.QueryRow("SELECT (user_id, permission) FROM token WHERE token = $1", token)
	if err = row.Err(); err == nil {
		err = row.Scan(&userID, &perm)
		exist = true
	}
	return
}

func (t *sqliteToken) Create(userID int, perm permission.Permission) (string, error) {
	tokenBuff := make([]byte, token.TokenSize)
	rand.Read(tokenBuff)
	tokenBuff[0] = 'b'
	tokenBuff[1] = 's'
	tokenBuff[2] = 'd'
	tokenString := hex.EncodeToString(tokenBuff)
	_, err := t.Exec("INSERT INTO tokens (user_id, token, permission) VALUE ($1, $2, $3)", userID, tokenString, perm)
	return tokenString, err
}

var _ user.User = &sqliteUser{}

type sqliteUser struct {
	userID         int
	userName       string
	userUsername   string
	userPermission permission.Permission

	dbConnection *sql.DB
}

func (u *sqliteUser) ID() int                           { return u.userID }
func (u *sqliteUser) Name() string                      { return u.userName }
func (u *sqliteUser) Username() string                  { return u.userUsername }
func (u *sqliteUser) Permission() permission.Permission { return u.userPermission }
func (u *sqliteUser) Password() (user.Password, error) {
	return &sqlitePassword{dbConnection: u.dbConnection, userID: u.userID}, nil
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
