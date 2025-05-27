package db

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"

	"sirherobrine23.com.br/go-bds/bds/module/server"
	"sirherobrine23.com.br/go-bds/bds/module/users"

	"github.com/docker/docker/pkg/namesgenerator"
)

var (
	SqliteCreateTables, _        = SQL.ReadFile("sql/create/sqlite.sql")
	SqliteInsertUserPassword, _  = SQL.ReadFile("sql/server/user_insert/sqlite.sql")
	SqliteInsertServer, _        = SQL.ReadFile("sql/server/server_insert/sqlite.sql")
	SqliteUserServers, _         = SQL.ReadFile("sql/server/server_list/sqlite.sql")
	SqliteServer, _              = SQL.ReadFile("sql/server/server_list/sqlite_id.sql")
	SqliteUpdateServer, _        = SQL.ReadFile("sql/server/update_server/sqlite.sql")
	SqliteServerFriends, _       = SQL.ReadFile("sql/server/server_friends/sqlite.sql")
	SqliteServerFriendsAdd, _    = SQL.ReadFile("sql/server/server_friends/sqlite_insert.sql")
	SqliteServerFriendsRemove, _ = SQL.ReadFile("sql/server/server_friends/sqlite_drop.sql")
	SqliteServerBackups, _       = SQL.ReadFile("sql/server/backup/sqlite.sql")

	SqliteUserInsert, _         = SQL.ReadFile("sql/user/create/sqlite.sql")
	SqliteUserInsertPassword, _ = SQL.ReadFile("sql/user/create/sqlite_password.sql")

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

func (slite *Sqlite) CreateNewUser(user *users.User, password *users.Password) (*users.User, error) {
	if err := password.HashPassword(*passwordToEncrypt); err != nil {
		return nil, err
	}

	// Insert user to database
	result, err := slite.Connection.Exec(string(SqliteUserInsert), user.Username, user.Name, user.Email)
	if err != nil {
		return nil, fmt.Errorf("cannot insert user: %s", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("cannot get new user ID: %s", err)
	}

	// Insert password to database
	if _, err = slite.Connection.Exec(string(SqliteUserInsertPassword), userID, password.Password); err != nil {
		return nil, fmt.Errorf("cannot insert password: %s", err)
	}

	return slite.UserID(userID)
}

func (slite *Sqlite) returnUser(row *sql.Row) (*users.User, error) {
	if err := row.Err(); err != nil {
		if err == sql.ErrNoRows {
			err = ErrUserNotExists
		}
		return nil, err
	}

	user := new(users.User)
	if err := row.Scan(&user.UserID, &user.Username, &user.Name, &user.Email, &user.CreateAt, &user.UpdateAt); err != nil {
		return nil, err
	}
	return user, nil
}

func (slite *Sqlite) Email(email string) (*users.User, error) {
	return slite.returnUser(slite.Connection.QueryRow("SELECT id, username, name, email, create_at, update_at FROM user WHERE email = LOWER($1)", email))
}
func (slite *Sqlite) Username(username string) (*users.User, error) {
	return slite.returnUser(slite.Connection.QueryRow("SELECT id, username, name, email, create_at, update_at FROM user WHERE username = LOWER($1)", username))
}
func (slite *Sqlite) UserID(id int64) (*users.User, error) {
	return slite.returnUser(slite.Connection.QueryRow("SELECT id, username, name, email, create_at, update_at FROM user WHERE id = $1", id))
}

func (slite *Sqlite) Password(UserID int64) (*users.Password, error) {
	row := slite.Connection.QueryRow("SELECT user, password, update_at FROM password WHERE user = $1", UserID)
	if err := row.Err(); err != nil {
		if err == sql.ErrNoRows {
			err = ErrUserNotExists
		}
		return nil, err
	}

	password := new(users.Password)
	if err := row.Scan(&password.UserID, &password.Password, &password.UpdateAt); err != nil {
		return nil, err
	}
	return password, nil
}

func (slite *Sqlite) CreateToken(user *users.User, perm ...users.TokenPermission) (*users.Token, error) {
	newRandData := make([]byte, 256)
	if _, err := rand.Read(newRandData); err != nil {
		return nil, err
	}
	tokenValue := hex.EncodeToString(newRandData)

	row := slite.Connection.QueryRow("INSERT INTO token (user, token, permissions) VALUES ($1, $2, $3); RETURNING id, user, token, permissions, create_at, update_at", user.UserID, tokenValue)
	if err := row.Err(); err != nil {
		if err == sql.ErrNoRows {
			err = ErrUserNotExists
		}
		return nil, err
	}

	token := new(users.Token)
	if err := row.Scan(&token.ID, &token.User, &token.Token, &token.Permissions, &token.CreateAt, &token.UpdateAt); err != nil {
		return nil, err
	}
	return token, nil
}

func (slite *Sqlite) Token(token string) (*users.Token, *users.User, error) {
	row := slite.Connection.QueryRow("SELECT id, user, token, permissions, create_at, update_at FROM token WHERE token = $1", token)
	if err := row.Err(); err != nil {
		if err == sql.ErrNoRows {
			err = io.EOF
		}
		return nil, nil, err
	}

	tokenStruct := new(users.Token)
	if err := row.Scan(&tokenStruct.ID, &tokenStruct.User, &tokenStruct.Token, &tokenStruct.Permissions, &tokenStruct.CreateAt, &tokenStruct.UpdateAt); err != nil {
		return nil, nil, err
	}

	user, err := slite.UserID(tokenStruct.User)
	if err != nil {
		return nil, nil, err
	}
	return tokenStruct, user, nil
}

func (slite *Sqlite) UpdateToken(token *users.Token, newPerms ...users.TokenPermission) error {
	_, err := slite.Connection.Exec("UPDATE token SET permissions = $1 WHERE token = $2", users.TokenPermissions(newPerms), token.Token)
	if err != nil {
		if err == sql.ErrNoRows {
			err = io.EOF
		}
		return err
	}
	token.Permissions = users.TokenPermissions(newPerms)
	return nil
}

func (slite *Sqlite) DeleteToken(token *users.Token) error {
	_, err := slite.Connection.Exec("DELETE FROM token WHERE token = $1", token.Token)
	if err == sql.ErrNoRows {
		err = io.EOF
	}
	return err
}

func (slite *Sqlite) CreateCookie(user *users.User) (*users.Cookie, *http.Cookie, error) {
	newRandData := make([]byte, 128)
	if _, err := rand.Read(newRandData); err != nil {
		return nil, nil, err
	}

	cookieValue := hex.EncodeToString(newRandData)
	row := slite.Connection.QueryRow("INSERT INTO cookie (user, cookie) VALUES ($1, $2); RETURNING id, user, cookie, create_at", user.UserID, cookieValue)
	if err := row.Err(); err != nil {
		return nil, nil, err
	}

	cookie := new(users.Cookie)
	if err := row.Scan(&cookie.ID, &cookie.User, &cookie.Cookie, &cookie.CreateAt); err != nil {
		return nil, nil, err
	}

	httpCookie := &http.Cookie{
		Name:     "bds",
		Path:     "/",
		Value:    cookie.Cookie,
		Expires:  time.Now().Add(DefaultCookieTime),
		SameSite: http.SameSiteStrictMode,
	}

	return cookie, httpCookie, nil
}

func (slite *Sqlite) DeleteCookie(cookie *users.Cookie) error {
	_, err := slite.Connection.Exec("DELETE FROM cookie WHERE cookie = $1", cookie.Cookie)
	if err == sql.ErrNoRows {
		err = io.EOF
	}
	return err
}

func (slite *Sqlite) Cookie(cookie *http.Cookie) (*users.User, error) {
	if cookie.Value == "" || len(cookie.Value) != 256 {
		return nil, fmt.Errorf("invalid cookie")
	}

	row := slite.Connection.QueryRow("SELECT user, create_at FROM cookie WHERE cookie = $1", cookie.Value)
	if err := row.Err(); err != nil {
		if err == sql.ErrNoRows {
			err = io.EOF
		}
		return nil, err
	}

	var userID int64
	var createAt time.Time
	if err := row.Scan(&userID, &createAt); err != nil {
		return nil, err
	}

	if createAt.Add(DefaultCookieTime).Compare(time.Now()) < 0 {
		return nil, fmt.Errorf("cookie expired")
	}

	return slite.UserID(userID)
}

func (slite *Sqlite) CreateServer(user *users.User, Server *server.Server) (*server.Server, error) {
	if user == nil {
		return nil, fmt.Errorf("set valid user to create server")
	}

	if Server == nil {
		Server = &server.Server{
			Owner:    user.UserID,
			Software: "bedrock",
			Version:  "latest",
		}
	}

	if Server.Name == "" {
		Server.Name = namesgenerator.GetRandomName(0)
	}

	// Insert server to database
	// owner, name, software, version
	result, err := slite.Connection.Exec(string(SqliteInsertServer),
		Server.Owner,
		Server.Name,
		Server.Software,
		Server.Version,
	)

	if err != nil {
		return nil, err
	}

	// Get server ID
	serverID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("cannot get new server ID: %s", err)
	}

	return slite.Server(serverID)
}

func (slite *Sqlite) UserServers(user *users.User) ([]*server.Server, error) {
	rows, err := slite.Connection.Query(string(SqliteUserServers), user.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrServerNotExists
		}
		return nil, err
	}

	var serversList []*server.Server
	for rows.Next() {
		server := new(server.Server)
		// id, name, owner, software, version, create_at, update_at
		if err := rows.Scan(&server.ID, &server.Name, &server.Owner, &server.Software, &server.Version, &server.CreateAt, &server.UpdateAt); err != nil {
			return nil, err
		}
		serversList = append(serversList, server)
	}

	return serversList, rows.Err()
}

func (slite *Sqlite) Server(ID int64) (*server.Server, error) {
	row := slite.Connection.QueryRow(string(SqliteUserServers), ID)
	if err := row.Err(); err != nil {
		if err == sql.ErrNoRows {
			err = ErrServerNotExists
		}
		return nil, err
	}

	server := new(server.Server)
	// id, name, owner, software, version, create_at, update_at
	if err := row.Scan(&server.ID, &server.Name, &server.Owner, &server.Software, &server.Version, &server.CreateAt, &server.UpdateAt); err != nil {
		return nil, err
	}

	return server, nil
}

func (slite *Sqlite) UpdateServer(server *server.Server) error {
	_, err := slite.Connection.Exec(string(SqliteUpdateServer), server.ID, server.Name, server.Software, server.Version)
	if err == sql.ErrNoRows {
		err = ErrServerNotExists
	}
	return err
}

func (slite *Sqlite) ServerFriends(serverID int64) ([]*server.ServerFriends, error) {
	// id, server_id, user_id, permissions
	rows, err := slite.Connection.Query(string(SqliteServerFriends), serverID)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrServerNotExists
		}
		return nil, err
	}

	var friendsList []*server.ServerFriends
	for rows.Next() {
		friend := new(server.ServerFriends)
		if err := rows.Scan(&friend.ID, &friend.ServerID, &friend.UserID, &friend.Permission); err != nil {
			return nil, err
		}
		friendsList = append(friendsList, friend)
	}
	return friendsList, rows.Err()
}

func (slite *Sqlite) AddNewFriend(server *server.Server, perm server.ServerPermissions, friends ...users.User) error {
	for _, friend := range friends {
		_, err := slite.Connection.Exec(string(SqliteServerFriendsAdd), server.ID, friend.UserID, perm)
		if err != nil {
			return err
		}
	}
	return nil
}

func (slite *Sqlite) RemoveFriend(server *server.Server, friends ...users.User) error {
	for _, friend := range friends {
		_, err := slite.Connection.Exec(string(SqliteServerFriendsRemove), server.ID, friend.UserID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (slite *Sqlite) ServerBackups(serverID int64) ([]*server.ServerBackup, error) {
	rows, err := slite.Connection.Query(string(SqliteServerBackups), serverID)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrServerNotExists
		}
		return nil, err
	}

	var backupsList []*server.ServerBackup
	for rows.Next() {
		backup := new(server.ServerBackup)
		// id, server_id, uuid, software, version, create_at
		if err := rows.Scan(&backup.ID, &backup.ServerID, &backup.UUID, &backup.Software, &backup.Version, &backup.CreateAt); err != nil {
			return nil, err
		}
		backupsList = append(backupsList, backup)
	}

	return backupsList, rows.Err()
}
