package server

import (
	"sirherobrine23.com.br/go-bds/bds/modules/datas/permission"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/user"
)

type ServerType uint

type ServerOwner struct {
	Permission permission.Permission // User permission in server
	User       user.User             // User
}

type ServerList interface {
	ByID(id int) (Server, error)      // Get server by id
	ByOwner(id int) ([]Server, error) // Get server by owner
}

type Server interface {
	ID() int                         // Server id
	Name() string                    // Server name
	ServerType() ServerType          // Server type
	ServerVersion() string           // Server version
	Owners() ([]*ServerOwner, error) // Server owners
}
