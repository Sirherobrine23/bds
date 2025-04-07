package db

type MinecraftServers struct {
	ServerID   int64  `json:"id" xorm:"'id' pk autoincr"`       // Server ID
	User       *User  `json:"user" xorm:"'user' notnull"`       // Server origin
	Name       string `json:"name" xorm:"'name' notnull"`       // Server name
	ServerType string `json:"server" xorm:"'server' notnull"`   // Server type, Bedrock, Java, etc ...
	Version    string `json:"version" xorm:"'version' notnull"` // Server version
}
