package db

import (
	"testing"

	"sirherobrine23.com.br/go-bds/bds/module/server"
	"sirherobrine23.com.br/go-bds/bds/module/users"
)

func TestDatabase(t *testing.T) {
	client, err := NewSqliteConnection(":memory:")
	if err != nil {
		t.Error(err)
		return
	}

	*passwordToEncrypt = "testBackend"
	user, err := client.CreateNewUser(&users.User{}, &users.Password{Password: "test1234"})
	if err != nil {
		t.Errorf("cannot make new user in database: %s", err)
		return
	}

	server, err := client.CreateServer(user, &server.Server{Software: "java", Version: "1.21.0", Owner: user.UserID, Name: "Test Server"})
	if err != nil {
		t.Errorf("cannot make new server in database: %s", err)
		return
	}

	t.Logf("Server ID: %d", server.ID)
}
