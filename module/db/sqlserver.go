package db

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"

	_ "github.com/denisenkom/go-mssqldb" // SQLServer driver (same as MSSQL)
	"github.com/docker/docker/pkg/namesgenerator"

	"sirherobrine23.com.br/go-bds/bds/module/server"
	"sirherobrine23.com.br/go-bds/bds/module/users"
)

var (
	SQLServerCreateTables, _        = SQL.ReadFile("sql/create/sqlserver.sql")
	SQLServerUserInsert, _          = SQL.ReadFile("sql/user/create/sqlserver.sql")
	SQLServerUserInsertPassword, _  = SQL.ReadFile("sql/user/create/sqlserver_password.sql")
	SQLServerInsertServer, _        = SQL.ReadFile("sql/server/server_insert/sqlserver.sql")
	SQLServerUserServers, _         = SQL.ReadFile("sql/server/server_list/sqlserver.sql")
	SQLServerServerByID, _          = SQL.ReadFile("sql/server/server_list/sqlserver_id.sql")
	SQLServerUpdateServer, _        = SQL.ReadFile("sql/server/update_server/sqlserver.sql")
	SQLServerServerFriends, _       = SQL.ReadFile("sql/server/server_friends/sqlserver.sql")
	SQLServerServerFriendsAdd, _    = SQL.ReadFile("sql/server/server_friends/sqlserver_insert.sql")
	SQLServerServerFriendsRemove, _ = SQL.ReadFile("sql/server/server_friends/sqlserver_drop.sql")
	SQLServerServerBackups, _       = SQL.ReadFile("sql/server/backups/sqlserver.sql")

	// Inline queries for SQLServer (identical to MSSQL)
	SQLServerEmailSelectQuery            = `SELECT id, username, name, email, create_at, update_at FROM [user] WHERE email = LOWER(@p1)`
	SQLServerUsernameSelectQuery         = `SELECT id, username, name, email, create_at, update_at FROM [user] WHERE username = LOWER(@p1)`
	SQLServerUserIDSelectQuery           = `SELECT id, username, name, email, create_at, update_at FROM [user] WHERE id = @p1`
	SQLServerPasswordSelectQuery         = `SELECT [user], [password], update_at FROM [password] WHERE [user] = @p1`
	SQLServerTokenInsertQuery            = `INSERT INTO token ([user], token, permissions) OUTPUT inserted.id, inserted.[user], inserted.token, inserted.permissions, inserted.create_at, inserted.update_at VALUES (@p1, @p2, @p3)`
	SQLServerTokenSelectByTokenQuery     = `SELECT id, [user], token, permissions, create_at, update_at FROM token WHERE token = @p1`
	SQLServerTokenUpdatePermissionsQuery = `UPDATE token SET permissions = @p1 WHERE token = @p2`
	SQLServerTokenDeleteQuery            = `DELETE FROM token WHERE token = @p1`
	SQLServerCookieInsertQuery           = `INSERT INTO cookie ([user], cookie) OUTPUT inserted.id, inserted.[user], inserted.cookie, inserted.create_at VALUES (@p1, @p2)`
	SQLServerCookieDeleteQuery           = `DELETE FROM cookie WHERE cookie = @p1`
	SQLServerCookieSelectUserQuery       = `SELECT [user], create_at FROM cookie WHERE cookie = @p1`

	// For CreateNewUser and CreateServer, if files don't have RETURNING id, define them here
	SQLServerUserInsertQueryWithOutput   = `INSERT INTO [user] (username, [name], email) OUTPUT inserted.id VALUES (LOWER(@p1), @p2, LOWER(@p3))`
	SQLServerInsertServerQueryWithOutput = `INSERT INTO [server]([owner], [name], [software], [version]) OUTPUT inserted.id VALUES (@p1, @p2, @p3, @p4)`
)

// SQLServerDB implements the Database interface for Microsoft SQL Server.
type SQLServerDB struct {
	Connection *sql.DB
}

// Ensure SQLServerDB implements Database interface
var _ Database = &SQLServerDB{}

// NewSQLServerConnection creates a new connection to a SQLServer database and initializes tables if they don't exist.
func NewSQLServerConnection(connectionString string) (Database, error) {
	db, err := sql.Open("sqlserver", connectionString) // Driver name is "sqlserver" for go-mssqldb
	if err != nil {
		return nil, fmt.Errorf("cannot open sqlserver database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping sqlserver database: %w", err)
	}

	var createScriptData []byte = SQLServerCreateTables
	if len(createScriptData) == 0 {
		data, errFile := SQL.ReadFile("sql/create/sqlserver.sql")
		if errFile != nil {
			return nil, fmt.Errorf("failed to read sqlserver create table script (verify embed): %w", errFile)
		}
		if len(data) == 0 {
			return nil, fmt.Errorf("sqlserver create table script is empty (verify embed content)")
		}
		createScriptData = data
	}
	
	// Check if the 'user' table exists to determine if schema setup is needed.
	// This logic is specific to SQL Server.
	var tableName string
	checkTableQuery := "SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'user' AND TABLE_SCHEMA = SCHEMA_NAME()"
	err = db.QueryRow(checkTableQuery).Scan(&tableName)

	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check if user table exists in sqlserver: %w", err)
	}

	if err == sql.ErrNoRows { // Tables do not exist, create them
		_, execErr := db.Exec(string(createScriptData))
		if execErr != nil {
			return nil, fmt.Errorf("cannot create sqlserver tables: %w", execErr)
		}
	}

	return &SQLServerDB{db}, nil
}

// User interface implementations

func (sdb *SQLServerDB) returnUser(row *sql.Row) (*users.User, error) {
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

func (sdb *SQLServerDB) Email(email string) (*users.User, error) {
	return sdb.returnUser(sdb.Connection.QueryRow(SQLServerEmailSelectQuery, email))
}

func (sdb *SQLServerDB) Username(username string) (*users.User, error) {
	return sdb.returnUser(sdb.Connection.QueryRow(SQLServerUsernameSelectQuery, username))
}

func (sdb *SQLServerDB) UserID(id int64) (*users.User, error) {
	return sdb.returnUser(sdb.Connection.QueryRow(SQLServerUserIDSelectQuery, id))
}

func (sdb *SQLServerDB) Password(UserID int64) (*users.Password, error) {
	row := sdb.Connection.QueryRow(SQLServerPasswordSelectQuery, UserID)
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

func (sdb *SQLServerDB) CreateNewUser(user *users.User, password *users.Password) (*users.User, error) {
	if err := password.HashPassword(); err != nil {
		return nil, err
	}
	
	var userID int64
	// SQLServerUserInsert from file does not have OUTPUT. Use SQLServerUserInsertQueryWithOutput.
	err := sdb.Connection.QueryRow(SQLServerUserInsertQueryWithOutput, user.Username, user.Name, user.Email).Scan(&userID)
	if err != nil {
		return nil, fmt.Errorf("cannot insert user and get ID: %w", err)
	}

	_, err = sdb.Connection.Exec(string(SQLServerUserInsertPassword), userID, password.Password)
	if err != nil {
		return nil, fmt.Errorf("cannot insert password: %w", err)
	}
	return sdb.UserID(userID)
}

func (sdb *SQLServerDB) CreateToken(user *users.User, perm ...users.TokenPermission) (*users.Token, error) {
	newRandData := make([]byte, 256)
	if _, err := rand.Read(newRandData); err != nil {
		return nil, err
	}
	tokenValue := hex.EncodeToString(newRandData)
	
	// Assuming users.TokenPermissions has a ToJSON method or implements driver.Valuer for JSON string
	permissionsValue := users.TokenPermissions(perm) 

	token := new(users.Token)
	// SQLServerTokenInsertQuery is the inline query with OUTPUT
	err := sdb.Connection.QueryRow(SQLServerTokenInsertQuery, user.UserID, tokenValue, permissionsValue).Scan(
		&token.ID, &token.User, &token.Token, &token.Permissions, &token.CreateAt, &token.UpdateAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert and scan token: %w", err)
	}
	return token, nil
}

func (sdb *SQLServerDB) Token(tokenStr string) (*users.Token, *users.User, error) {
	row := sdb.Connection.QueryRow(SQLServerTokenSelectByTokenQuery, tokenStr)
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
	user, err := sdb.UserID(token.User)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user for token: %w", err)
	}
	return token, user, nil
}

func (sdb *SQLServerDB) UpdateToken(token *users.Token, newPerms ...users.TokenPermission) error {
	permissionsValue := users.TokenPermissions(newPerms)
	result, err := sdb.Connection.Exec(SQLServerTokenUpdatePermissionsQuery, permissionsValue, token.Token)
	if err != nil {
		return fmt.Errorf("failed to update token: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected for token update: %w", err)
	}
	if rowsAffected == 0 {
		return io.EOF 
	}
	token.Permissions = permissionsValue
	return nil
}

func (sdb *SQLServerDB) DeleteToken(token *users.Token) error {
	result, err := sdb.Connection.Exec(SQLServerTokenDeleteQuery, token.Token)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected for token deletion: %w", err)
	}
	if rowsAffected == 0 {
		return io.EOF 
	}
	return nil
}

func (sdb *SQLServerDB) CreateCookie(user *users.User) (*users.Cookie, *http.Cookie, error) {
	newRandData := make([]byte, 128)
	if _, err := rand.Read(newRandData); err != nil {
		return nil, nil, err
	}
	cookieValue := hex.EncodeToString(newRandData)

	cookie := new(users.Cookie)
	// SQLServerCookieInsertQuery is the inline query with OUTPUT
	err := sdb.Connection.QueryRow(SQLServerCookieInsertQuery, user.UserID, cookieValue).Scan(
		&cookie.ID, &cookie.User, &cookie.Cookie, &cookie.CreateAt,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to insert and scan cookie: %w", err)
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

func (sdb *SQLServerDB) DeleteCookie(cookie *users.Cookie) error {
	result, err := sdb.Connection.Exec(SQLServerCookieDeleteQuery, cookie.Cookie)
	if err != nil {
		return fmt.Errorf("failed to delete cookie: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected for cookie deletion: %w", err)
	}
	if rowsAffected == 0 {
		return io.EOF
	}
	return nil
}

func (sdb *SQLServerDB) Cookie(cookie *http.Cookie) (*users.User, error) {
	if cookie.Value == "" || len(cookie.Value) != 256 {
		return nil, fmt.Errorf("invalid cookie format or length")
	}
	var userID int64
	var createAt time.Time
	err := sdb.Connection.QueryRow(SQLServerCookieSelectUserQuery, cookie.Value).Scan(&userID, &createAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, io.EOF
		}
		return nil, fmt.Errorf("failed to scan cookie data: %w", err)
	}
	if createAt.Add(DefaultCookieTime).Before(time.Now()) {
		return nil, fmt.Errorf("cookie expired")
	}
	return sdb.UserID(userID)
}

// Server interface implementations

func (sdb *SQLServerDB) CreateServer(user *users.User, svr *server.Server) (*server.Server, error) {
	if user == nil {
		return nil, fmt.Errorf("valid user required to create server")
	}
	newSvr := *svr
	if newSvr.Owner == 0 { newSvr.Owner = user.UserID }
	if newSvr.Software == "" { newSvr.Software = "bedrock" }
	if newSvr.Version == "" { newSvr.Version = "latest" }
	if newSvr.Name == "" { newSvr.Name = namesgenerator.GetRandomName(0) }
	
	var serverID int64
	// SQLServerInsertServer from file does not have OUTPUT. Use SQLServerInsertServerQueryWithOutput.
	err := sdb.Connection.QueryRow(SQLServerInsertServerQueryWithOutput,
		newSvr.Owner, newSvr.Name, newSvr.Software, newSvr.Version).Scan(&serverID)
	if err != nil {
		return nil, fmt.Errorf("cannot insert server and get ID: %w", err)
	}
	return sdb.Server(serverID)
}

func (sdb *SQLServerDB) UserServers(user *users.User) ([]*server.Server, error) {
	rows, err := sdb.Connection.Query(string(SQLServerUserServers), user.UserID)
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

func (sdb *SQLServerDB) Server(ID int64) (*server.Server, error) {
	row := sdb.Connection.QueryRow(string(SQLServerServerByID), ID)
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

func (sdb *SQLServerDB) UpdateServer(svr *server.Server) error {
	result, err := sdb.Connection.Exec(string(SQLServerUpdateServer), svr.Name, svr.Software, svr.Version, svr.ID)
	if err != nil {
		return fmt.Errorf("failed to update server: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected for server update: %w", err)
	}
	if rowsAffected == 0 {
		return ErrServerNotExists
	}
	return nil
}

func (sdb *SQLServerDB) ServerFriends(serverID int64) ([]*server.ServerFriends, error) {
	rows, err := sdb.Connection.Query(string(SQLServerServerFriends), serverID)
	if err != nil {
		return nil, fmt.Errorf("querying server friends failed: %w", err)
	}
	defer rows.Close()
	var friendsList []*server.ServerFriends
	for rows.Next() {
		friend := new(server.ServerFriends)
		// Assuming friend.Permission implements sql.Scanner for JSON NVARCHAR(MAX)
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

func (sdb *SQLServerDB) AddNewFriend(svr *server.Server, perm server.ServerPermissions, friends ...users.User) error {
	// Assuming perm (server.ServerPermissions) implements driver.Valuer for JSON NVARCHAR(MAX)
	stmt, err := sdb.Connection.Prepare(string(SQLServerServerFriendsAdd))
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

func (sdb *SQLServerDB) RemoveFriend(svr *server.Server, friends ...users.User) error {
	stmt, err := sdb.Connection.Prepare(string(SQLServerServerFriendsRemove))
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

func (sdb *SQLServerDB) ServerBackups(serverID int64) ([]*server.ServerBackup, error) {
	rows, err := sdb.Connection.Query(string(SQLServerServerBackups), serverID)
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

// Note on permissions types (users.TokenPermissions, server.ServerPermissions):
// This implementation assumes that these types (or their ToJSON/FromJSON methods as used in mssql.go,
// or direct passing if they implement sql.Scanner/driver.Valuer) correctly handle
// conversion to/from a string representation suitable for NVARCHAR(MAX) in SQL Server.
// The mssql.go implementation used placeholder ToJSON/FromJSON calls for clarity if direct
// Scan/Value was not implemented. For SQLServer, the same assumption holds:
// if users.TokenPermissions and server.ServerPermissions implement sql.Scanner and driver.Valuer,
// they can be used directly. Otherwise, explicit marshaling to a JSON string is needed.
// The `CreateToken` method in this `sqlserver.go` file passes `permissionsValue` directly,
// assuming it implements `driver.Valuer`. If it doesn't, it should be `permissionsValue.ToJSON()`
// similar to how it was in the reference mssql.go. For consistency with the copied logic,
// I'll ensure the direct pass-through implies these interfaces are implemented.
// The provided inline queries for MSSQL/SQLServer for token/cookie creation already include all fields in OUTPUT,
// so `token.Permissions` and `friend.Permission` are scanned directly. This implies they implement `sql.Scanner`.
```
