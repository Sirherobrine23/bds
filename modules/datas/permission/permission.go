package permission

import (
	"encoding/json"
	"slices"
	"strings"
)

// User permission
type Permission int

const (
	Unknown Permission = -1 // No valid permission
	Root    Permission = 0  // Have all permission on bds-dashboard

	WebCreateServer Permission = 8  // Create server
	WebDeleteServer Permission = 16 // Delete server
	WebListServer   Permission = 24 // List other server in bds-dashboard
	WebEditServer   Permission = 32 // Edit server after creation

	ServerView   Permission = 64  // Server view
	ServerEdit   Permission = 128 // Server edit
	ServerUpdate Permission = 256 // Server update

	// Server owner
	ServerOwner Permission = ServerView | ServerEdit | ServerUpdate
)

var permisionString = map[Permission]string{
	WebCreateServer: "create_server",
	WebDeleteServer: "delete_server",
	WebListServer:   "list_server",
	WebEditServer:   "edit_server",

	ServerView:   "server_view",
	ServerEdit:   "server_edit",
	ServerUpdate: "server_update",
	ServerOwner:  "server_owner",
}

func (p Permission) IsRoot() bool { return p == Root }

// Get all permissions from p2 and check if have in p
func (p Permission) Contains(p2 Permission) bool {
	if p.IsRoot() {
		return true
	}
	pd := p.Permisions()
	ps := p2.Permisions()
	return slices.ContainsFunc(pd, func(pd Permission) bool { return slices.Contains(ps, pd) })
}

func (p Permission) String() string {
	if str, ok := permisionString[p]; ok {
		return str
	}

	ps := []string{}
	for _, pe := range p.Permisions() {
		ps = append(ps, pe.String())
	}
	return strings.Join(ps, ",")
}

func (p Permission) MarshalText() ([]byte, error) { return []byte(p.String()), nil }

func (p Permission) MarshalJSON() ([]byte, error) {
	ps := []string{}
	for _, pe := range p.Permisions() {
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

func (p Permission) Permisions() []Permission {
	ps := []Permission{}
	for per := range permisionString {
		if p&per > 0 {
			ps = append(ps, per)
			p &= per
		}
	}
	return ps
}
