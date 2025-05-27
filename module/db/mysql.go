package db

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
	"github.com/docker/docker/pkg/namesgenerator"

	"sirherobrine23.com.br/go-bds/bds/module/server"
	"sirherobrine23.com.br/go-bds/bds/module/users"
)

var (
	MysqlCreateTables, _        = SQL.ReadFile("sql/create/mysql.sql")
	MysqlUserInsertFile, _      = SQL.ReadFile("sql/user/create/mysql.sql") 
	MysqlUserInsertPassword, _  = SQL.ReadFile("sql/user/create/mysql_password.sql")
	MysqlInsertServerFile, _    = SQL.ReadFile("sql/server/server_insert/mysql.sql") 
	MysqlUserServers, _         = SQL.ReadFile("sql/server/server_list/mysql.sql")
	MysqlServerByID, _          = SQL.ReadFile("sql/server/server_list/mysql_id.sql")
	MysqlUpdateServer, _        = SQL.ReadFile("sql/server/update_server/mysql.sql")
	MysqlServerFriends, _       = SQL.ReadFile("sql/server/server_friends/mysql.sql")
	MysqlServerFriendsAdd, _    = SQL.ReadFile("sql/server/server_friends/mysql_insert.sql")
	MysqlServerFriendsRemove, _ = SQL.ReadFile("sql/server/server_friends/mysql_drop.sql")
	MysqlServerBackups, _       = SQL.ReadFile("sql/server/backups/mysql.sql")

	// Inline queries for direct use (as per task spec)
	MysqlEmailSelectQuery            = "SELECT id, username, `name`, email, create_at, update_at FROM `user` WHERE email = LOWER(?)"
	MysqlUsernameSelectQuery         = "SELECT id, username, `name`, email, create_at, update_at FROM `user` WHERE username = LOWER(?)"
	MysqlUserIDSelectQuery           = "SELECT id, username, `name`, email, create_at, update_at FROM `user` WHERE id = ?"
	MysqlPasswordSelectQuery         = "SELECT `user`, `password`, update_at FROM `password` WHERE `user` = ?"
	MysqlTokenInsertQuery            = "INSERT INTO `token` (`user`, token, permissions) VALUES (?, ?, ?)" 
	MysqlTokenSelectByTokenQuery     = "SELECT id, `user`, token, permissions, create_at, update_at FROM `token` WHERE token = ?" 
	MysqlTokenUpdatePermissionsQuery = "UPDATE `token` SET permissions = ? WHERE token = ?" 
	MysqlTokenDeleteQuery            = "DELETE FROM `token` WHERE token = ?" 
	MysqlCookieInsertQuery           = "INSERT INTO `cookie` (`user`, cookie) VALUES (?, ?)" 
	MysqlCookieDeleteQuery           = "DELETE FROM `cookie` WHERE cookie = ?" 
	MysqlCookieSelectUserQuery       = "SELECT `user`, create_at FROM `cookie` WHERE cookie = ?" 

	// Additional query needed for CreateCookie to fetch all fields
	MysqlCookieSelectFullQuery = "SELECT id, `user`, cookie, create_at FROM `cookie` WHERE cookie = ?"
)

// MysqlDB implements the Database interface for MySQL.
type MysqlDB struct {
	Connection *sql.DB
}

// Ensure MysqlDB implements Database interface
var _ Database = &MysqlDB{}

// NewMysqlConnection creates a new connection to a MySQL database and initializes tables if they don't exist.
func NewMysqlConnection(connectionString string) (Database, error) {
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, fmt.Errorf("cannot open mysql database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping mysql database: %w", err)
	}
	
	// Ensure create script is loaded, handling potential silent failures from embed
	var createScriptData []byte = MysqlCreateTables
	if len(createScriptData) == 0 {
		data, errFile := SQL.ReadFile("sql/create/mysql.sql")
		if errFile != nil {
			return nil, fmt.Errorf("failed to read mysql create table script (verify embed): %w", errFile)
		}
		if len(data) == 0 {
			return nil, fmt.Errorf("mysql create table script is empty (verify embed content)")
		}
		createScriptData = data
	}

	_, execErr := db.Exec(string(createScriptData))
	if execErr != nil {
		return nil, fmt.Errorf("cannot create mysql tables: %w", execErr)
	}

	return &MysqlDB{db}, nil
}

// User interface implementations

func (mdb *MysqlDB) returnUser(row *sql.Row) (*users.User, error) {
	if err := row.Err(); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotExists
		}
		return nil, err
	}
	user := new(users.User)
	if err := row.Scan(&user.UserID, &user.Username, &user.Name, &user.Email, &user.CreateAt, &user.UpdateAt); err != nil {
		return nil, fmt.Errorf("failed to scan user data: %w", err)
	}
	return user, nil
}

func (mdb *MysqlDB) Email(email string) (*users.User, error) {
	return mdb.returnUser(mdb.Connection.QueryRow(MysqlEmailSelectQuery, email))
}

func (mdb *MysqlDB) Username(username string) (*users.User, error) {
	return mdb.returnUser(mdb.Connection.QueryRow(MysqlUsernameSelectQuery, username))
}

func (mdb *MysqlDB) UserID(id int64) (*users.User, error) {
	return mdb.returnUser(mdb.Connection.QueryRow(MysqlUserIDSelectQuery, id))
}

func (mdb *MysqlDB) Password(UserID int64) (*users.Password, error) {
	row := mdb.Connection.QueryRow(MysqlPasswordSelectQuery, UserID)
	if err := row.Err(); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotExists
		}
		return nil, err
	}
	password := new(users.Password)
	if err := row.Scan(&password.UserID, &password.Password, &password.UpdateAt); err != nil {
		return nil, fmt.Errorf("failed to scan password data: %w", err)
	}
	return password, nil
}

func (mdb *MysqlDB) CreateNewUser(user *users.User, password *users.Password) (*users.User, error) {
	if err := password.HashPassword(); err != nil {
		return nil, err
	}

	result, err := mdb.Connection.Exec(string(MysqlUserInsertFile), user.Username, user.Name, user.Email)
	if err != nil {
		return nil, fmt.Errorf("cannot insert user: %w", err)
	}
	userID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("cannot get new user ID: %w", err)
	}

	_, err = mdb.Connection.Exec(string(MysqlUserInsertPassword), userID, password.Password)
	if err != nil {
		return nil, fmt.Errorf("cannot insert password: %w", err)
	}

	return mdb.UserID(userID)
}

func (mdb *MysqlDB) CreateToken(user *users.User, perm ...users.TokenPermission) (*users.Token, error) {
	newRandData := make([]byte, 256)
	if _, err := rand.Read(newRandData); err != nil {
		return nil, err
	}
	tokenValue := hex.EncodeToString(newRandData)
	permissions := users.TokenPermissions(perm)

	_, err := mdb.Connection.Exec(MysqlTokenInsertQuery, user.UserID, tokenValue, permissions)
	if err != nil {
		return nil, fmt.Errorf("cannot insert token: %w", err)
	}

	token := new(users.Token)
	err = mdb.Connection.QueryRow(MysqlTokenSelectByTokenQuery, tokenValue).Scan(
		&token.ID, &token.User, &token.Token, &token.Permissions, &token.CreateAt, &token.UpdateAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("failed to retrieve token after insert: %w", io.EOF) 
		}
		return nil, fmt.Errorf("failed to scan token after insert: %w", err)
	}
	return token, nil
}

func (mdb *MysqlDB) Token(tokenStr string) (*users.Token, *users.User, error) {
	row := mdb.Connection.QueryRow(MysqlTokenSelectByTokenQuery, tokenStr)
	if err := row.Err(); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, io.EOF
		}
		return nil, nil, fmt.Errorf("querying token failed: %w", err)
	}
	token := new(users.Token)
	if err := row.Scan(&token.ID, &token.User, &token.Token, &token.Permissions, &token.CreateAt, &token.UpdateAt); err != nil {
		return nil, nil, fmt.Errorf("failed to scan token data: %w", err)
	}
	user, err := mdb.UserID(token.User)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user for token: %w", err)
	}
	return token, user, nil
}

func (mdb *MysqlDB) UpdateToken(token *users.Token, newPerms ...users.TokenPermission) error {
	permissions := users.TokenPermissions(newPerms)
	result, err := mdb.Connection.Exec(MysqlTokenUpdatePermissionsQuery, permissions, token.Token)
	if err != nil {
		return fmt.Errorf("failed to update token: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return io.EOF
	}
	token.Permissions = permissions
	return nil
}

func (mdb *MysqlDB) DeleteToken(token *users.Token) error {
	result, err := mdb.Connection.Exec(MysqlTokenDeleteQuery, token.Token)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return io.EOF
	}
	return nil
}

func (mdb *MysqlDB) CreateCookie(user *users.User) (*users.Cookie, *http.Cookie, error) {
	newRandData := make([]byte, 128)
	if _, err := rand.Read(newRandData); err != nil {
		return nil, nil, err
	}
	cookieValue := hex.EncodeToString(newRandData)

	_, err := mdb.Connection.Exec(MysqlCookieInsertQuery, user.UserID, cookieValue)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot insert cookie: %w", err)
	}

	cookie := new(users.Cookie)
	err = mdb.Connection.QueryRow(MysqlCookieSelectFullQuery, cookieValue).Scan(
		&cookie.ID, &cookie.User, &cookie.Cookie, &cookie.CreateAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, fmt.Errorf("failed to retrieve cookie after insert: %w", io.EOF)
		}
		return nil, nil, fmt.Errorf("failed to scan cookie after insert: %w", err)
	}

	httpCookie := &http.Cookie{
		Name:     "bds",
		Path:     "/",
		Value:    cookie.Cookie,
		Expires:  time.Now().Add(DefaultCookieTime),
		SameSite: http.SameSiteStrictMode,
		HttpOnly: true,
		Secure:   true,
	}
	return cookie, httpCookie, nil
}

func (mdb *MysqlDB) DeleteCookie(cookie *users.Cookie) error {
	result, err := mdb.Connection.Exec(MysqlCookieDeleteQuery, cookie.Cookie)
	if err != nil {
		return fmt.Errorf("failed to delete cookie: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return io.EOF
	}
	return nil
}

func (mdb *MysqlDB) Cookie(cookie *http.Cookie) (*users.User, error) {
	if cookie.Value == "" || len(cookie.Value) != 256 {
		return nil, fmt.Errorf("invalid cookie format or length")
	}
	var userID int64
	var createAt time.Time
	err := mdb.Connection.QueryRow(MysqlCookieSelectUserQuery, cookie.Value).Scan(&userID, &createAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, io.EOF
		}
		return nil, fmt.Errorf("failed to scan cookie data: %w", err)
	}
	if createAt.Add(DefaultCookieTime).Before(time.Now()) {
		return nil, fmt.Errorf("cookie expired")
	}
	return mdb.UserID(userID)
}

// Server interface implementations

func (mdb *MysqlDB) CreateServer(user *users.User, svr *server.Server) (*server.Server, error) {
	if user == nil {
		return nil, fmt.Errorf("valid user required to create server")
	}
	newSvr := *svr
	if newSvr.Owner == 0 { newSvr.Owner = user.UserID }
	if newSvr.Software == "" { newSvr.Software = "bedrock" }
	if newSvr.Version == "" { newSvr.Version = "latest" }
	if newSvr.Name == "" { newSvr.Name = namesgenerator.GetRandomName(0) }

	result, err := mdb.Connection.Exec(string(MysqlInsertServerFile),
		newSvr.Owner, newSvr.Name, newSvr.Software, newSvr.Version)
	if err != nil {
		return nil, fmt.Errorf("cannot insert server: %w", err)
	}
	serverID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("cannot get new server ID: %w", err)
	}
	return mdb.Server(serverID)
}

func (mdb *MysqlDB) UserServers(user *users.User) ([]*server.Server, error) {
	rows, err := mdb.Connection.Query(string(MysqlUserServers), user.UserID, user.UserID)
	if err != nil {
		return nil, fmt.Errorf("querying user servers failed: %w", err)
	}
	defer rows.Close()
	var serversList []*server.Server
	for rows.Next() {
		s := new(server.Server)
		if err := rows.Scan(&s.ID, &s.Name, &s.Owner, &s.Software, &s.Version, &s.CreateAt, &s.UpdateAt); err != nil {
			return nil, fmt.Errorf("failed to scan server data: %w", err)
		}
		serversList = append(serversList, s)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iteration error over server rows: %w", err)
	}
	return serversList, nil
}

func (mdb *MysqlDB) Server(ID int64) (*server.Server, error) {
	row := mdb.Connection.QueryRow(string(MysqlServerByID), ID)
	if err := row.Err(); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrServerNotExists
		}
		return nil, fmt.Errorf("querying server by ID failed: %w", err)
	}
	s := new(server.Server)
	if err := row.Scan(&s.ID, &s.Name, &s.Owner, &s.Software, &s.Version, &s.CreateAt, &s.UpdateAt); err != nil {
		return nil, fmt.Errorf("failed to scan server data: %w", err)
	}
	return s, nil
}

func (mdb *MysqlDB) UpdateServer(svr *server.Server) error {
	result, err := mdb.Connection.Exec(string(MysqlUpdateServer), svr.Name, svr.Software, svr.Version, svr.ID)
	if err != nil {
		return fmt.Errorf("failed to update server: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrServerNotExists
	}
	return nil
}

func (mdb *MysqlDB) ServerFriends(serverID int64) ([]*server.ServerFriends, error) {
	rows, err := mdb.Connection.Query(string(MysqlServerFriends), serverID)
	if err != nil {
		return nil, fmt.Errorf("querying server friends failed: %w", err)
	}
	defer rows.Close()
	var friendsList []*server.ServerFriends
	for rows.Next() {
		friend := new(server.ServerFriends)
		if err := rows.Scan(&friend.ID, &friend.ServerID, &friend.UserID, &friend.Permission); err != nil {
			return nil, fmt.Errorf("failed to scan server friend data: %w", err)
		}
		friendsList = append(friendsList, friend)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iteration error over server friend rows: %w", err)
	}
	return friendsList, nil
}

func (mdb *MysqlDB) AddNewFriend(svr *server.Server, perm server.ServerPermissions, friends ...users.User) error {
	stmt, err := mdb.Connection.Prepare(string(MysqlServerFriendsAdd))
	if err != nil {
		return fmt.Errorf("failed to prepare add friend statement: %w", err)
	}
	defer stmt.Close()
	for _, friend := range friends {
		_, err := stmt.Exec(svr.ID, friend.UserID, perm)
		if err != nil {
			return fmt.Errorf("failed to add friend %s: %w", friend.Username, err)
		}
	}
	return nil
}

func (mdb *MysqlDB) RemoveFriend(svr *server.Server, friends ...users.User) error {
	stmt, err := mdb.Connection.Prepare(string(MysqlServerFriendsRemove))
	if err != nil {
		return fmt.Errorf("failed to prepare remove friend statement: %w", err)
	}
	defer stmt.Close()
	for _, friend := range friends {
		result, err := stmt.Exec(svr.ID, friend.UserID)
		if err != nil {
			return fmt.Errorf("failed to remove friend %s: %w", friend.Username, err)
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return fmt.Errorf("friend %s not found on server %d or not removed: %w", friend.Username, svr.ID, ErrUserNotExists)
		}
	}
	return nil
}

func (mdb *MysqlDB) ServerBackups(serverID int64) ([]*server.ServerBackup, error) {
	rows, err := mdb.Connection.Query(string(MysqlServerBackups), serverID)
	if err != nil {
		return nil, fmt.Errorf("querying server backups failed: %w", err)
	}
	defer rows.Close()
	var backupsList []*server.ServerBackup
	for rows.Next() {
		backup := new(server.ServerBackup)
		if err := rows.Scan(&backup.ID, &backup.ServerID, &backup.UUID, &backup.Software, &backup.Version, &backup.CreateAt); err != nil {
			return nil, fmt.Errorf("failed to scan server backup data: %w", err)
		}
		backupsList = append(backupsList, backup)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iteration error over server backup rows: %w", err)
	}
	return backupsList, nil
}
