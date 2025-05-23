package server

import (
	"database/sql/driver"
	"encoding/json"
)

type ServerPermission int

type ServerPermissions []ServerPermission

const (
	Unknown ServerPermission = iota
	View
	Edit
)

func (ns ServerPermission) String() string {
	switch ns {
	case View:
		return "view"
	case Edit:
		return "edit"
	default:
		return "unknown"
	}
}

func (ns ServerPermission) TextMarshall() ([]byte, error) { return []byte(ns.String()), nil }
func (ns *ServerPermission) TextUnmarshal(data []byte) error {
	switch string(data) {
	case "view":
		*ns = View
	case "edit":
		*ns = Edit
	default:
		*ns = Unknown
	}
	return nil
}

// Scan implements the [Scanner] interface.
func (ns *ServerPermissions) Scan(value any) error {
	if value == nil {
		*ns = ServerPermissions{Unknown}
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
func (ns ServerPermissions) Value() (driver.Value, error) {
	d, err := json.Marshal(ns)
	if err != nil {
		return nil, err
	}
	return string(d), nil
}
