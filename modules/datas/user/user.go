package user

import "sirherobrine23.com.br/go-bds/bds/modules/datas/permission"

type UserSearch interface {
	Username(string) (User, error)
	ByID(int) (User, error)
}

// User info
type User interface {
	ID() int                           // Return unique user ID to reference in all points
	Name() string                      // User name
	Username() string                  // Username/nick name
	Permission() permission.Permission // User permissions
	Password() (Password, error)       // Return password manipulation
}

// Password check and storage
type Password interface {
	Check(password string) (bool, error) // Check password is valid
	Storage(password string) error       // Storage password
}
