package permission

import (
	"encoding/json"
	"slices"
	"strings"
)

// User permission
type Permission uint

const (
	Unknown Permission = iota << 4

	CreateServer // Create server
	DeleteServer // Delete server
	ListServer   // List other server in bds-dashboard
	EditServer   // Edit server after creation

	// Have all permission on bds-dashboard
	Root = CreateServer |
		DeleteServer |
		ListServer |
		EditServer
)

var permisionString = map[Permission]string{
	CreateServer: "create_server",
	DeleteServer: "delete_server",
	ListServer:   "list_server",
	EditServer:   "edit_server",
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
