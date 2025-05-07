package permission

import (
	"encoding/json"
	"slices"
	"strings"
)

// User permission
type Permission int

const (
	Root Permission = iota << 4 // Have all permission on bds-dashboard

	CreateServer // Create server
	DeleteServer // Delete server
	ListServer   // List other server in bds-dashboard
	EditServer   // Edit server after creation

	ServerOwner  // Server owner
	ServerView   // Server view
	ServerEdit   // Server edit
	ServerUpdate // Server update

	Unknown = -1 // No valid permission
)

var permisionString = map[Permission]string{
	CreateServer: "create_server",
	DeleteServer: "delete_server",
	ListServer:   "list_server",
	EditServer:   "edit_server",

	ServerOwner:  "server_owner",
	ServerView:   "server_view",
	ServerEdit:   "server_edit",
	ServerUpdate: "server_update",
}

func (p Permission) IsRoot() bool               { return p == Root }
func (p Permission) Compare(p2 Permission) bool { return p&p2 > 0 }

func (p Permission) String() string {
	if str, ok := permisionString[p]; ok {
		return str
	}

	ps := []string{}
	for _, pe := range p.Parsions() {
		ps = append(ps, pe.String())
	}
	return strings.Join(ps, ",")
}

func (p Permission) MarshalText() ([]byte, error) { return []byte(p.String()), nil }

func (p Permission) MarshalJSON() ([]byte, error) {
	ps := []string{}
	for _, pe := range p.Parsions() {
		ps = append(ps, pe.String())
	}
	return json.Marshal(ps)
}

func (p *Permission) UnmarshalJSON(data []byte) error {
	var ps []string
	if err := json.Unmarshal(data, &ps); err != nil {
		return err
	}

	*p = 0
	for pe, name := range permisionString {
		if slices.Contains(ps, name) {
			*p |= pe
		}
	}

	return nil
}

func (p Permission) Parsions() []Permission {
	ps := []Permission{}
	for per := range permisionString {
		if p&per > 0 {
			ps = append(ps, per)
			p &= per
		}
	}
	return ps
}
