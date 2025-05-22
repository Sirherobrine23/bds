// User information and authenticantion
package users

import "time"

// User representation
type User struct {
	UserID   int64     `json:"id"`        // User ID to make linkers
	Username string    `json:"username"`  // Username
	Name     string    `json:"name"`      // Name to show
	Email    string    `json:"email"`     // user email to check unique value
	CreateAt time.Time `json:"create_at"` // Date of user creation
	UpdateAt time.Time `json:"update_at"` // Date to update any row in database
}

// Password storage
type Password struct {
	UserID   int64     `json:"id"`        // User ID, foregin key
	UpdateAt time.Time `json:"update_at"` // Data password update
	Password string    `json:"password"`  // Password hash
}

// Cookie storage to web
type Cookie struct {
	ID       int64     `json:"id"`        // Cookie id
	User     int64     `json:"user_id"`   // User ID
	Cookie   string    `json:"cookie"`    // cookie value
	CreateAt time.Time `json:"create_at"` // time creation
}

// Token to auth API router
type Token struct {
	ID          int64             `json:"id"`          // Cookie id
	User        int64             `json:"user_id"`     // User ID
	Token       string            `json:"token"`       // Token value in hex code
	Permissions []TokenPermission `json:"permissions"` // Token permission
	CreateAt    time.Time         `json:"create_at"`   // time creation
	UpdateAt    time.Time         `json:"update_at"`   // Date to update any row in database
}
