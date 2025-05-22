package users

// Token permission
type TokenPermission int

// Token permissions
const (
	Unknown      TokenPermission = iota // Disabled token
	CreateServer                        // Create server
	DeleteServer                        // Delete server
	UpdateServer                        // Update Server
)
