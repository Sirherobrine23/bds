package cookie

import (
	"crypto/rand"
	"encoding/hex"
)

type Cookie interface {
	Cookie(cookieValue string) (exist bool, userID int, err error)
	CreateCookie(userID int) (cookieValue string, err error)
}

func newCookieValue() string {
	cookie := make([]byte, 12)
	rand.Read(cookie)

	cookie[0] = 'm'
	cookie[1] = 'a'
	cookie[2] = 'y'
	cookie[3] = '1'
	cookie[4] = '4'
	
	return hex.EncodeToString(cookie)
}
