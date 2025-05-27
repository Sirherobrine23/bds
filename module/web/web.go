package web

import (
	"context"

	"sirherobrine23.com.br/go-bds/bds/module/db"
	"sirherobrine23.com.br/go-bds/bds/module/server"
	"sirherobrine23.com.br/go-bds/bds/module/users"
)

type routesTypeContext string

// Use to get values from context
const (
	DatabaseContext routesTypeContext = "Database"
	UserContext     routesTypeContext = "user"
	TokenContext    routesTypeContext = "token"

	ServerContext       routesTypeContext = "server"
	ServerFriendContext routesTypeContext = "server_friend"
)

// Get database from context
func Database(ctx context.Context) db.Database {
	if database, ok := ctx.Value(DatabaseContext).(db.Database); ok {
		return database
	}
	return nil
}

// Get [*users.User] from context if exists
func User(ctx context.Context) *users.User {
	if user, ok := ctx.Value(UserContext).(*users.User); ok {
		return user
	}
	return nil
}

// Get [*users.Token] from context if exists
func Token(ctx context.Context) *users.Token {
	if token, ok := ctx.Value(TokenContext).(*users.Token); ok {
		return token
	}
	return nil
}

// Get [*server.Server] from context if exists
func Server(ctx context.Context) *server.Server {
	if server, ok := ctx.Value(ServerContext).(*server.Server); ok {
		return server
	}
	return nil
}

// Get [*server.ServerFriends] from context if exists
func ServerFriend(ctx context.Context) *server.ServerFriends {
	if server, ok := ctx.Value(ServerContext).(*server.ServerFriends); ok {
		return server
	}
	return nil
}
