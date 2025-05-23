package users

import (
	"database/sql/driver"
	"encoding/json"
)

// Token permission
type TokenPermission int

type TokenPermissions []TokenPermission

// Token permissions
const (
	Unknown      TokenPermission = iota // Disabled token
	CreateServer                        // Create server
	DeleteServer                        // Delete server
	UpdateServer                        // Update Server
)

func (ns TokenPermission) String() string {
	switch ns {
	case CreateServer:
		return "create_server"
	case DeleteServer:
		return "delete_server"
	case UpdateServer:
		return "update_server"
	default:
		return "unknown"
	}
}

func (ns TokenPermission) TextMarshall() ([]byte, error) { return []byte(ns.String()), nil }
func (ns *TokenPermission) TextUnmarshal(data []byte) error {
	switch string(data) {
	case "create_server":
		*ns = CreateServer
	case "delete_server":
		*ns = DeleteServer
	case "update_server":
		*ns = UpdateServer
	default:
		*ns = Unknown
	}
	return nil
}

// Scan implements the [Scanner] interface.
func (ns *TokenPermissions) Scan(value any) error {
	if value == nil {
		*ns = TokenPermissions{Unknown}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, ns)
	case string:
		return json.Unmarshal([]byte(v), ns)
	}

	return nil
}

// Value implements the [driver.Valuer] interface.
func (ns TokenPermissions) Value() (driver.Value, error) {
	d, err := json.Marshal(ns)
	if err != nil {
		return nil, err
	}
	return string(d), nil
}
