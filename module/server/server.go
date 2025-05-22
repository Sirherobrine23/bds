// Server maneger
package server

import "time"

// Server info
type Server struct {
	ID    int64  `json:"id"`    // Server ID
	Owner int64  `json:"owner"` // Server owner, forekin key
	Name  string `json:"name"`  // Server name

	Software string `json:"software"` // Server software
	Version  string `json:"version"`  // Server version

	CreateAt time.Time `json:"create_at"` // Date of creation
	UpdateAt time.Time `json:"update_at"` // Date to update any row in database
}

// Servers external users
type ServerFriends struct {
	ID         int64              `json:"id"`          // ID
	ServerID   int64              `json:"server_id"`   // Server ID, foregin key
	UserID     int64              `json:"user_id"`     // user ID, foregin key
	Permission []ServerPermission `json:"permissions"` // Permission
}

// Server backup
type ServerBackup struct {
	ID       int64  `json:"id"`        // Backup ID
	ServerID int64  `json:"server_id"` // Server reference, foregin key
	UUID     string `json:"uuid"`      // Backup UUID
}

// Runner info
type ServerRunner struct {
	ID     int64 `json:"id"`      // Runner ID
	Global bool  `json:"global"`  // Runner is global, to all users in instance
	Local  bool  `json:"local"`   // Runner is to the specifiq user
	UserID int64 `json:"user_id"` // user id if is local runner
}
