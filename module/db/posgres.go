package db

import (
	"database/sql"
	"fmt"
	"net/http"

	"sirherobrine23.com.br/go-bds/bds/module/server"
	"sirherobrine23.com.br/go-bds/bds/module/users"
)

var (
	PostgresCreateTables, _       = SQL.ReadFile("sql/create/Postgres.sql")
	PostgresInsertUserPassword, _ = SQL.ReadFile("sql/user_insert/Postgres.sql")

	_ Database = &Postgres{}
)

type Postgres struct {
	Connection *sql.DB
}

func NewPostgresConnection(connection string) (Database, error) {
	db, err := sql.Open("Postgres", connection)
	if err != nil {
		return nil, fmt.Errorf("cannot open posgres connection: %s", err)
	}

	// Create table is not exists
	if _, err = db.Exec(string(PostgresCreateTables)); err != nil {
		return nil, fmt.Errorf("cannot create tables: %s", err)
	}

	return &Postgres{db}, nil
}

func (ps *Postgres) AddNewFriend(server *server.Server, perm []server.ServerPermission, friends ...users.User) error
func (ps *Postgres) Cookie(cookie *http.Cookie) (*users.User, error)
func (ps *Postgres) CreateCookie(user *users.User) (*users.Cookie, *http.Cookie, error)
func (ps *Postgres) CreateNewUser(user *users.User) (*users.User, error)
func (ps *Postgres) CreateServer(user *users.User, server *server.Server) (*server.Server, error)
func (ps *Postgres) CreateToken(user *users.User, perm ...users.TokenPermission) (*users.Token, error)
func (ps *Postgres) DeleteCookie(cookie *users.Cookie) error
func (ps *Postgres) DeleteToken(token *users.Token) error
func (ps *Postgres) ID(id int64) (*users.User, error)
func (ps *Postgres) Password(UserID int64) (*users.Password, error)
func (ps *Postgres) RemoveFriend(server *server.Server, friends ...users.User) error
func (ps *Postgres) Server(ID int64) (*server.Server, error)
func (ps *Postgres) Token(token string) (*users.Token, *users.User, error)
func (ps *Postgres) UpdateServer(server *server.Server) error
func (ps *Postgres) UpdateToken(token *users.Token, newPerms ...users.TokenPermission) error
func (ps *Postgres) UserServers(user *users.User) ([]*server.Server, error)
func (ps *Postgres) Username(username string) (*users.User, error)
