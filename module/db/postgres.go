package db

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/docker/docker/pkg/namesgenerator"

	"sirherobrine23.com.br/go-bds/bds/module/server"
	"sirherobrine23.com.br/go-bds/bds/module/users"
)

var (
	PostgresCreateTables, _        = SQL.ReadFile("sql/create/postgresql.sql")
	PostgresUserInsertFile, _      = SQL.ReadFile("sql/user/create/postgresql.sql") // INSERT ... (no RETURNING specified in file, needs adjustment or local override)
	PostgresUserInsertPassword, _  = SQL.ReadFile("sql/user/create/postgresql_password.sql")
	PostgresInsertServerFile, _    = SQL.ReadFile("sql/server/server_insert/postgresql.sql") // INSERT ... (no RETURNING specified in file, needs adjustment or local override)
	PostgresUserServers, _         = SQL.ReadFile("sql/server/server_list/postgresql.sql")
	PostgresServerByID, _          = SQL.ReadFile("sql/server/server_list/postgresql_id.sql")
	PostgresUpdateServer, _        = SQL.ReadFile("sql/server/update_server/postgresql.sql")
	PostgresServerFriends, _       = SQL.ReadFile("sql/server/server_friends/postgresql.sql")
	PostgresServerFriendsAdd, _    = SQL.ReadFile("sql/server/server_friends/postgresql_insert.sql")
	PostgresServerFriendsRemove, _ = SQL.ReadFile("sql/server/server_friends/postgresql_drop.sql")
	PostgresServerBackups, _       = SQL.ReadFile("sql/server/backups/postgresql.sql")

	// Inline queries for direct use (PostgreSQL syntax)
	PostgresEmailSelectQuery            = `SELECT id, username, "name", email, create_at, update_at FROM "user" WHERE email = LOWER($1)`
	PostgresUsernameSelectQuery         = `SELECT id, username, "name", email, create_at, update_at FROM "user" WHERE username = LOWER($1)`
	PostgresUserIDSelectQuery           = `SELECT id, username, "name", email, create_at, update_at FROM "user" WHERE id = $1`
	PostgresPasswordSelectQuery         = `SELECT "user", "password", update_at FROM "password" WHERE "user" = $1`
	PostgresTokenInsertQuery            = `INSERT INTO "token" ("user", token, permissions) VALUES ($1, $2, $3) RETURNING id, "user", token, permissions, create_at, update_at`
	PostgresTokenSelectByTokenQuery     = `SELECT id, "user", token, permissions, create_at, update_at FROM "token" WHERE token = $1`
	PostgresTokenUpdatePermissionsQuery = `UPDATE "token" SET permissions = $1 WHERE token = $2`
	PostgresTokenDeleteQuery            = `DELETE FROM "token" WHERE token = $1`
	PostgresCookieInsertQuery           = `INSERT INTO "cookie" ("user", cookie) VALUES ($1, $2) RETURNING id, "user", cookie, create_at`
	PostgresCookieDeleteQuery           = `DELETE FROM "cookie" WHERE cookie = $1`
	PostgresCookieSelectUserQuery       = `SELECT "user", create_at FROM "cookie" WHERE cookie = $1`

	// For CreateNewUser and CreateServer, if files don't have RETURNING id, define them here
	PostgresUserInsertQueryWithReturnID   = `INSERT INTO "user" (username, "name", email) VALUES (LOWER($1), $2, LOWER($3)) RETURNING id`
	PostgresServerInsertQueryWithReturnID = `INSERT INTO "server" ("owner", "name", software, "version") VALUES ($1, $2, $3, $4) RETURNING id`
)

// PostgresDB implements the Database interface for PostgreSQL.
type PostgresDB struct {
	Connection *sql.DB
}

// Ensure PostgresDB implements Database interface
var _ Database = &PostgresDB{}

// NewPostgresConnection creates a new connection to a PostgreSQL database and initializes tables.
func NewPostgresConnection(connectionString string) (Database, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("cannot open postgres database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres database: %w", err)
	}

	var createScriptData []byte = PostgresCreateTables
	if len(createScriptData) == 0 {
		data, errFile := SQL.ReadFile("sql/create/postgresql.sql")
		if errFile != nil {
			return nil, fmt.Errorf("failed to read postgres create table script (verify embed): %w", errFile)
		}
		if len(data) == 0 {
			return nil, fmt.Errorf("postgres create table script is empty (verify embed content)")
		}
		createScriptData = data
	}

	_, execErr := db.Exec(string(createScriptData))
	if execErr != nil {
		return nil, fmt.Errorf("cannot create postgres tables: %w", execErr)
	}

	return &PostgresDB{db}, nil
}

// User interface implementations

func (pg *PostgresDB) returnUser(row *sql.Row) (*users.User, error) {
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

func (pg *PostgresDB) Email(email string) (*users.User, error) {
	return pg.returnUser(pg.Connection.QueryRow(PostgresEmailSelectQuery, email))
}

func (pg *PostgresDB) Username(username string) (*users.User, error) {
	return pg.returnUser(pg.Connection.QueryRow(PostgresUsernameSelectQuery, username))
}

func (pg *PostgresDB) UserID(id int64) (*users.User, error) {
	return pg.returnUser(pg.Connection.QueryRow(PostgresUserIDSelectQuery, id))
}

func (pg *PostgresDB) Password(UserID int64) (*users.Password, error) {
	row := pg.Connection.QueryRow(PostgresPasswordSelectQuery, UserID)
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

func (pg *PostgresDB) CreateNewUser(user *users.User, password *users.Password) (*users.User, error) {
	if err := password.HashPassword(); err != nil {
		return nil, err
	}

	var userID int64
	// PostgresUserInsertFile from sql/user/create/postgresql.sql is: INSERT INTO "user" (username, "name", email) VALUES (LOWER($1), $2, LOWER($3));
	// We need RETURNING id. Use the locally defined PostgresUserInsertQueryWithReturnID.
	err := pg.Connection.QueryRow(PostgresUserInsertQueryWithReturnID, user.Username, user.Name, user.Email).Scan(&userID)
	if err != nil {
		return nil, fmt.Errorf("cannot insert user and get ID: %w", err)
	}

	// PostgresUserInsertPassword from file: INSERT INTO "password"("user", "password") VALUES ($1, $2);
	_, err = pg.Connection.Exec(string(PostgresUserInsertPassword), userID, password.Password)
	if err != nil {
		return nil, fmt.Errorf("cannot insert password: %w", err)
	}

	return pg.UserID(userID)
}

func (pg *PostgresDB) CreateToken(user *users.User, perm ...users.TokenPermission) (*users.Token, error) {
	newRandData := make([]byte, 256)
	if _, err := rand.Read(newRandData); err != nil {
		return nil, err
	}
	tokenValue := hex.EncodeToString(newRandData)
	permissions := users.TokenPermissions(perm) // Assumes this type implements driver.Valuer for JSONB

	token := new(users.Token)
	err := pg.Connection.QueryRow(PostgresTokenInsertQuery, user.UserID, tokenValue, permissions).Scan(
		&token.ID, &token.User, &token.Token, &token.Permissions, &token.CreateAt, &token.UpdateAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert and scan token: %w", err)
	}
	return token, nil
}

func (pg *PostgresDB) Token(tokenStr string) (*users.Token, *users.User, error) {
	row := pg.Connection.QueryRow(PostgresTokenSelectByTokenQuery, tokenStr)
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
	user, err := pg.UserID(token.User)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user for token: %w", err)
	}
	return token, user, nil
}

func (pg *PostgresDB) UpdateToken(token *users.Token, newPerms ...users.TokenPermission) error {
	permissions := users.TokenPermissions(newPerms)
	result, err := pg.Connection.Exec(PostgresTokenUpdatePermissionsQuery, permissions, token.Token)
	if err != nil {
		return fmt.Errorf("failed to update token: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected for token update: %w", err)
	}
	if rowsAffected == 0 {
		return io.EOF // Token not found to update
	}
	token.Permissions = permissions
	return nil
}

func (pg *PostgresDB) DeleteToken(token *users.Token) error {
	result, err := pg.Connection.Exec(PostgresTokenDeleteQuery, token.Token)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected for token deletion: %w", err)
	}
	if rowsAffected == 0 {
		return io.EOF // Token not found to delete
	}
	return nil
}

func (pg *PostgresDB) CreateCookie(user *users.User) (*users.Cookie, *http.Cookie, error) {
	newRandData := make([]byte, 128)
	if _, err := rand.Read(newRandData); err != nil {
		return nil, nil, err
	}
	cookieValue := hex.EncodeToString(newRandData)

	cookie := new(users.Cookie)
	err := pg.Connection.QueryRow(PostgresCookieInsertQuery, user.UserID, cookieValue).Scan(
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

func (pg *PostgresDB) DeleteCookie(cookie *users.Cookie) error {
	result, err := pg.Connection.Exec(PostgresCookieDeleteQuery, cookie.Cookie)
	if err != nil {
		return fmt.Errorf("failed to delete cookie: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected for cookie deletion: %w", err)
	}
	if rowsAffected == 0 {
		return io.EOF // Cookie not found to delete
	}
	return nil
}

func (pg *PostgresDB) Cookie(cookie *http.Cookie) (*users.User, error) {
	if cookie.Value == "" || len(cookie.Value) != 256 {
		return nil, fmt.Errorf("invalid cookie format or length")
	}
	var userID int64
	var createAt time.Time
	err := pg.Connection.QueryRow(PostgresCookieSelectUserQuery, cookie.Value).Scan(&userID, &createAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, io.EOF // Cookie not found
		}
		return nil, fmt.Errorf("failed to scan cookie data: %w", err)
	}
	if createAt.Add(DefaultCookieTime).Before(time.Now()) {
		return nil, fmt.Errorf("cookie expired")
	}
	return pg.UserID(userID)
}

// Server interface implementations

func (pg *PostgresDB) CreateServer(user *users.User, svr *server.Server) (*server.Server, error) {
	if user == nil {
		return nil, fmt.Errorf("valid user required to create server")
	}
	newSvr := *svr
	if newSvr.Owner == 0 { newSvr.Owner = user.UserID }
	if newSvr.Software == "" { newSvr.Software = "bedrock" }
	if newSvr.Version == "" { newSvr.Version = "latest" }
	if newSvr.Name == "" { newSvr.Name = namesgenerator.GetRandomName(0) }

	var serverID int64
	// PostgresInsertServerFile from sql/server/server_insert/postgresql.sql is: INSERT INTO "server"("owner", "name", software, "version") VALUES ($1, $2, $3, $4);
	// We need RETURNING id. Use the locally defined PostgresServerInsertQueryWithReturnID.
	err := pg.Connection.QueryRow(PostgresServerInsertQueryWithReturnID,
		newSvr.Owner, newSvr.Name, newSvr.Software, newSvr.Version).Scan(&serverID)
	if err != nil {
		return nil, fmt.Errorf("cannot insert server and get ID: %w", err)
	}
	return pg.Server(serverID)
}

func (pg *PostgresDB) UserServers(user *users.User) ([]*server.Server, error) {
	rows, err := pg.Connection.Query(string(PostgresUserServers), user.UserID) // Query from file
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

func (pg *PostgresDB) Server(ID int64) (*server.Server, error) {
	row := pg.Connection.QueryRow(string(PostgresServerByID), ID) // Query from file
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

func (pg *PostgresDB) UpdateServer(svr *server.Server) error {
	// Query from file: UPDATE "server" SET update_at = current_timestamp, "name" = $2, software = $3, "version" = $4 WHERE id = $1;
	// Parameter order from file: name, software, version, id
	result, err := pg.Connection.Exec(string(PostgresUpdateServer), svr.Name, svr.Software, svr.Version, svr.ID)
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

func (pg *PostgresDB) ServerFriends(serverID int64) ([]*server.ServerFriends, error) {
	rows, err := pg.Connection.Query(string(PostgresServerFriends), serverID) // Query from file
	if err != nil {
		return nil, fmt.Errorf("querying server friends failed: %w", err)
	}
	defer rows.Close()
	var friendsList []*server.ServerFriends
	for rows.Next() {
		friend := new(server.ServerFriends)
		// Assumes friend.Permission implements sql.Scanner
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

func (pg *PostgresDB) AddNewFriend(svr *server.Server, perm server.ServerPermissions, friends ...users.User) error {
	stmt, err := pg.Connection.Prepare(string(PostgresServerFriendsAdd)) // Query from file
	if err != nil {
		return fmt.Errorf("failed to prepare add friend statement: %w", err)
	}
	defer stmt.Close()
	for _, friend := range friends {
		// Assumes perm (server.ServerPermissions) implements driver.Valuer
		_, err := stmt.Exec(svr.ID, friend.UserID, perm)
		if err != nil {
			return fmt.Errorf("failed to add friend %s: %w", friend.Username, err)
		}
	}
	return nil
}

func (pg *PostgresDB) RemoveFriend(svr *server.Server, friends ...users.User) error {
	stmt, err := pg.Connection.Prepare(string(PostgresServerFriendsRemove)) // Query from file
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

func (pg *PostgresDB) ServerBackups(serverID int64) ([]*server.ServerBackup, error) {
	rows, err := pg.Connection.Query(string(PostgresServerBackups), serverID) // Query from file
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
