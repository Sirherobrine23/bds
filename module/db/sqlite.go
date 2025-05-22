package db

import (
	"database/sql"
	"fmt"
	"net/http"

	"sirherobrine23.com.br/go-bds/bds/module/server"
	"sirherobrine23.com.br/go-bds/bds/module/users"
)

var (
	SqliteCreateTables, _       = SQL.ReadFile("sql/create/sqlite.sql")
	SqliteInsertUserPassword, _ = SQL.ReadFile("sql/user_insert/sqlite.sql")

	_ Database = &Sqlite{}
)

type Sqlite struct {
	Connection *sql.DB
}

func NewSqliteConnection(connection string) (Database, error) {
	db, err := sql.Open("sqlite3", connection)
	if err != nil {
		return nil, fmt.Errorf("cannot open sqlite database: %s", err)
	}

	// Create table is not exists
	if _, err = db.Exec(string(SqliteCreateTables)); err != nil {
		return nil, fmt.Errorf("cannot create tables: %s", err)
	}

	return &Sqlite{db}, nil
}

func (sqlite *Sqlite) AddNewFriend(server *server.Server, perm []server.ServerPermission, friends ...users.User) error
func (sqlite *Sqlite) Cookie(cookie *http.Cookie) (*users.User, error)
func (sqlite *Sqlite) CreateCookie(user *users.User) (*users.Cookie, *http.Cookie, error)
func (sqlite *Sqlite) CreateNewUser(user *users.User) (*users.User, error)
func (sqlite *Sqlite) CreateServer(user *users.User, server *server.Server) (*server.Server, error)
func (sqlite *Sqlite) CreateToken(user *users.User, perm ...users.TokenPermission) (*users.Token, error)
func (sqlite *Sqlite) DeleteCookie(cookie *users.Cookie) error
func (sqlite *Sqlite) DeleteToken(token *users.Token) error
func (sqlite *Sqlite) ID(id int64) (*users.User, error)
func (sqlite *Sqlite) Password(UserID int64) (*users.Password, error)
func (sqlite *Sqlite) RemoveFriend(server *server.Server, friends ...users.User) error
func (sqlite *Sqlite) Server(ID int64) (*server.Server, error)
func (sqlite *Sqlite) Token(token string) (*users.Token, *users.User, error)
func (sqlite *Sqlite) UpdateServer(server *server.Server) error
func (sqlite *Sqlite) UpdateToken(token *users.Token, newPerms ...users.TokenPermission) error
func (sqlite *Sqlite) UserServers(user *users.User) ([]*server.Server, error)
func (sqlite *Sqlite) Username(username string) (*users.User, error)
