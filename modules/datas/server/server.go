package server

import (
	"database/sql"
	"fmt"

	"sirherobrine23.com.br/go-bds/bds/modules/datas/permission"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/user"
)

type ServerType uint

// Server types
const (
	Bedrock ServerType = iota
	Java
	Pocketmine
	AllayMC
	SpigotMC
	PurpurMC
	PaperMC
	FoliaMC
	VelocityMC
)

var serverNames = []string{
	Bedrock:    "Bedrock",
	Java:       "Java",
	Pocketmine: "PocketminineMP",
	AllayMC:    "AllayMC",
	SpigotMC:   "Spigot",
	PurpurMC:   "Purpur",
	PaperMC:    "Paper",
	FoliaMC:    "Folia",
	VelocityMC: "Velocity",
}

func (s ServerType) String() string { return serverNames[s] }

type ServerOwner struct {
	Permission permission.Permission // User permission in server
	User       *user.User            // User
}

type ServerOwners []*ServerOwner

func (s ServerOwners) UserID(id int64) (*ServerOwner, bool) {
	for _, user := range s {
		if user.User.ID == id {
			return user, true
		}
	}
	return nil, false
}

// Server
type Server struct {
	ID            int64          // Server id
	Name          string         // Server name
	ServerType    ServerType     // Server type
	ServerVersion string         // Server version
	Owners        ServerOwners // Server owners
}

type ServerList struct {
	Driver string  // Driver name
	DB     *sql.DB // Database connection
}

// Get server by owner ID
func (serverDB *ServerList) ByOwner(id int64) ([]*Server, error) {
	switch serverDB.Driver {
	case "sqlite3":
		return serverDB.sqliteByOwner(id)
	default:
		return nil, fmt.Errorf("%s drive not supported in Server list", serverDB.Driver)
	}
}

// Get server by id
func (serverDB *ServerList) ByID(id int64) (*Server, error) {
	switch serverDB.Driver {
	case "sqlite3":
		return serverDB.sqliteByID(id)
	default:
		return nil, fmt.Errorf("%s drive not supported in Server list", serverDB.Driver)
	}
}

// Create server
func (serverDB *ServerList) CreateServer(name, serverVersion string, serverType ServerType, owner *user.User) (*Server, error) {
	if serverVersion == "latest" {
		switch serverType {
		case Bedrock:
		case Java:
		case Pocketmine:
		case AllayMC:
		case SpigotMC:
		case PurpurMC:
		case PaperMC:
		case FoliaMC:
		case VelocityMC:
		}
	}

	switch serverDB.Driver {
	case "sqlite3":
		return serverDB.sqliteCreateServer(name, serverVersion, serverType, owner)
	default:
		return nil, fmt.Errorf("%s drive not supported in Server list", serverDB.Driver)
	}
}
