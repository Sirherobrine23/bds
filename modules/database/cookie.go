package db

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

type Cookie struct {
	ID          int64     `json:"id" xorm:"pk autoincr"`                 // Cookie ID
	User        *User     `json:"user" xorm:"user"`                      // User info
	CookieValue string    `json:"cookie" xorm:"'cookie' notnull unique"` // Cookie value
	ValidAt     time.Time `json:"valid_at" xorm:"'valid' notnull"`       // Cookie time to valid
}

func (cookie *Cookie) SetupCookie() error {
	buff := make([]byte, 16)
	if _, err := rand.Read(buff); err != nil {
		return err
	}
	cookie.CookieValue = hex.EncodeToString(buff)                    // Append cookie struct
	cookie.ValidAt = time.Now().UTC().Add(time.Hour * 24 * 31 * 125) // Add time to struct
	_, err := DatabaseConnection.InsertOne(cookie)
	return err
}
