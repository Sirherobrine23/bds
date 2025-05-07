package server

import (
	"database/sql"

	_ "embed"

	"sirherobrine23.com.br/go-bds/bds/modules/datas/permission"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/user"
)

var (
	//go:embed sql/sqlite.sql
	createTableSqlite string

	//go:embed sql/sqlite_owners.sql
	listOwnerSqlite string

	//go:embed sql/sqlite_servers.sql
	listServersSqlite string
)

type Sqlite struct {
	ServerID            int
	ServerName          string
	ServerServerType    ServerType
	ServerServerVersion string

	db *sql.DB
}

// Create table on start
func SqliteStartTable(connection *sql.DB) error {
	// Create begin to make rollback if error
	tx, err := connection.Begin()
	if err != nil {
		return err
	}

	// Create table
	if _, err := tx.Exec(createTableSqlite); err != nil {
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

type SqliteSearch struct{ *sql.DB }

func ServerSqlite(conn *sql.DB) ServerList { return &SqliteSearch{conn} }

func (sql *Sqlite) ID() int                { return sql.ServerID }
func (sql *Sqlite) Name() string           { return sql.ServerName }
func (sql *Sqlite) ServerType() ServerType { return sql.ServerServerType }
func (sql *Sqlite) ServerVersion() string  { return sql.ServerServerVersion }

func (sql *Sqlite) Owners() ([]*ServerOwner, error) {
	// p.user_id, p.permission, u.name, u.username, u.permission
	rows, err := sql.db.Query(listOwnerSqlite, sql.ServerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []*ServerOwner{}
	for rows.Next() {
		var perm, userPerm permission.Permission
		var userID int
		var name, username string

		if err := rows.Scan(&userID, &perm, &name, &username, &userPerm); err != nil {
			return users, err
		}

		users = append(users, &ServerOwner{
			Permission: perm,
			User: &user.SqliteUser{
				DB: sql.db,

				UserID:         userID,
				UserPermission: userPerm,
				UserName:       name,
				UserUsername:   username,
			},
		})
	}

	return users, rows.Err()
}

func (server *SqliteSearch) ByID(id int) (Server, error) {
	row := server.QueryRow("SELECT id, name, server_type, server_version FROM servers WHERE id = $1", id)
	if err := row.Err(); err != nil {
		return nil, err
	}
	userServer := &Sqlite{db: server.DB}
	return userServer, row.Scan(&userServer.ServerID, &userServer.ServerName, &userServer.ServerServerType, &userServer.ServerServerVersion)
}

func (server *SqliteSearch) ByOwner(id int) ([]Server, error) {
	rows, err := server.Query(listServersSqlite, id, permission.ServerOwner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	servers := []Server{}
	for rows.Next() {
		userServer := &Sqlite{db: server.DB}
		if err := rows.Scan(
			&userServer.ServerID,
			&userServer.ServerName,
			&userServer.ServerServerType,
			&userServer.ServerServerVersion,
		); err != nil {
			return servers, err
		}

		servers = append(servers, userServer)
	}

	return servers, rows.Err()
}
