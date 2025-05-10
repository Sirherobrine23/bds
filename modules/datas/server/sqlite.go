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
)

func CreateSqliteTable(conn *sql.DB) error {
	_, err := conn.Exec(createTableSqlite)
	return err
}

func (server *ServerList) sqliteByID(id int64) (*Server, error) {
	row := server.DB.QueryRow("SELECT id, name, server_type, server_version FROM servers WHERE id = $1", id)
	if err := row.Err(); err != nil {
		return nil, err
	}

	userServer := &Server{Owners: []*ServerOwner{}}
	if err := row.Scan(&userServer.ID, &userServer.Name, &userServer.ServerType, &userServer.ServerVersion); err != nil {
		return nil, err
	}

	// p.user_id, p.permission, u.name, u.username, u.permission
	ownerRows, err := server.DB.Query(listOwnerSqlite, id)
	if err != nil {
		return nil, err
	}
	defer ownerRows.Close()

	for ownerRows.Next() {
		var userID int64
		var name, username string
		var perm, userPerm permission.Permission

		if err := ownerRows.Scan(&userID, &perm, &name, &username, &userPerm); err != nil {
			return nil, err
		}

		user, err := (&user.UserSearch{Driver: server.Driver, DB: server.DB}).ByID(userID)
		if err != nil {
			return nil, err
		}
		userServer.Owners = append(userServer.Owners, &ServerOwner{Permission: perm, User: user})
	}

	if err := ownerRows.Err(); err != nil {
		return nil, err
	}

	return userServer, nil
}

func (server *ServerList) sqliteByOwner(id int64) ([]*Server, error) {
	var serverIDs []int64
	serverIDsRows, err := server.DB.Query("SELECT server_id FROM servers_permission WHERE user_id = $1 AND permission = $2", id, permission.ServerOwner)
	if err != nil {
		return nil, err
	}

	for serverIDsRows.Next() {
		var serverID int64
		if err := serverIDsRows.Scan(&serverID); err != nil {
			return nil, err
		}
		serverIDs = append(serverIDs, serverID)
	}

	if err := serverIDsRows.Err(); err != nil {
		return nil, err
	}

	servers := []*Server{}
	for _, serverID := range serverIDs {
		server, err := server.ByID(serverID)
		if err != nil {
			return nil, err
		}
		servers = append(servers, server)
	}

	return servers, nil
}

func (server *ServerList) sqliteCreateServer(name, serverVersion string, serverType ServerType, owner *user.User) (*Server, error) {
	res, err := server.DB.Exec("INSERT INTO servers(\"name\", server_type, server_version) VALUES ($1, $2, $3)", name, serverType, serverVersion)
	if err != nil {
		return nil, err
	}

	serverID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	_, err = server.DB.Exec("INSERT INTO servers_permission(user_id, server_id, permission) VALUES ($1, $2, $3)", owner.ID, serverID, permission.ServerOwner)
	if err != nil {
		return nil, err
	}

	return server.ByID(serverID)
}
