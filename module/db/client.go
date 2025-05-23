package db

import (
	"embed"
	"errors"
	"net/http"
	"time"

	_ "sirherobrine23.com.br/go-bds/bds/module/db/internal/sqlclients"
	"sirherobrine23.com.br/go-bds/bds/module/server"

	"sirherobrine23.com.br/go-bds/bds/module/users"
)

// SQL files to open and interactive with database
//
//go:embed sql
var SQL embed.FS

var (
	ErrServerNotExists error = errors.New("server not exists")
	ErrUserNotExists   error = errors.New("user not exists")

	DefaultCookieTime = time.Hour * 24 * 7 * 30 * 15
)

// Database interface
type Database interface {
	User
	Server
}

// Return database to user
type User interface {
	Username(username string) (*users.User, error) // Get by username user
	UserID(id int64) (*users.User, error)          // get by ID user

	Password(UserID int64) (*users.Password, error)        // Get from database password storage
	Cookie(cookie *http.Cookie) (*users.User, error)       // Get users from cookie
	Token(token string) (*users.Token, *users.User, error) // Get token
	Email(email string) (*users.User, error)               // Get by email user

	CreateNewUser(user *users.User, password *users.Password) (*users.User, error)     // Create new user
	CreateCookie(user *users.User) (*users.Cookie, *http.Cookie, error)                // Create cookie
	CreateToken(user *users.User, perm ...users.TokenPermission) (*users.Token, error) // Create token

	DeleteCookie(cookie *users.Cookie) error // Remove cookie
	DeleteToken(token *users.Token) error    // Delete token

	UpdateToken(token *users.Token, newPerms ...users.TokenPermission) error // Update permissions to token
}

// Server maneger
type Server interface {
	Server(ID int64) (*server.Server, error)                       // Get server by ID
	UserServers(user *users.User) ([]*server.Server, error)        // get all server to user
	ServerFriends(serverID int64) ([]*server.ServerFriends, error) // Get server friends by server ID
	ServerBackups(serverID int64) ([]*server.ServerBackup, error)  // Get server backups by server ID

	CreateServer(user *users.User, Server *server.Server) (*server.Server, error) // Create new server
	UpdateServer(Server *server.Server) error                                     // Update server

	AddNewFriend(Server *server.Server, perm server.ServerPermissions, friends ...users.User) error // Add new users to server friends list
	RemoveFriend(Server *server.Server, friends ...users.User) error                                // Remove friends from server
}
