// Module only api structs
package router

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	app "sirherobrine23.com.br/go-bds/bds/modules"

	"sirherobrine23.com.br/go-bds/bds/modules/datas"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/permission"
)

// JSON message to return
type ErrorResponse struct {
	From    string `json:"error"`
	Message string `json:"message,omitempty"`
}

type AppVersion struct {
	Version string        `json:"version"`
	Uptime  time.Duration `json:"uptime"`
}

var apiRouter = chi.NewMux()

// Mount router api with new config
func ApiRouter(config *datas.DatabaseSchemas) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WebApi.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ContextConfig, config)))
	})
}

// Process API request with config from context
var WebApi = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	config := getConfig(r)
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			jsonWrite(http.StatusInternalServerError, w, ErrorResponse{
				From:    "Internal Server Error",
				Message: fmt.Sprintf("error: %s", err),
			})
			return
		}
	}()

	// Set token context
	if auth := r.Header.Get("Authentication"); auth != "" && (strings.HasPrefix(auth, "Bearer") || strings.HasPrefix(auth, "token")) {
		tokenString := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(auth, "token"), "Bearer"))
		ok, userID, perm, err := config.Token.Check(tokenString)
		switch {
		case err != nil:
			jsonWrite(http.StatusBadRequest, w, ErrorResponse{
				From:    "token",
				Message: err.Error(),
			})
			return
		case !ok, perm == permission.Unknown:
			jsonWrite(http.StatusBadRequest, w, ErrorResponse{
				From:    "token",
				Message: "token not exists",
			})
			return
		}

		user, err := config.User.ByID(userID)
		if err != nil {
			jsonWrite(http.StatusBadRequest, w, ErrorResponse{
				From:    "user",
				Message: err.Error(),
			})
			return
		}

		r = r.WithContext(context.WithValue(
			context.WithValue(
				context.WithValue(r.Context(),
					ContextUser, user),
				ContextToken, tokenString),
			ContextTokenPerm, perm),
		)
	}

	// Caller api router handler
	apiRouter.ServeHTTP(w, r)
})

func init() {
	apiRouter.Get("/", func(w http.ResponseWriter, r *http.Request) {
		jsonWrite(http.StatusOK, w, AppVersion{
			Version: app.AppVersion,
			Uptime:  time.Since(app.StartTime),
		})
	})

	apiRouter.Get("/servers", func(w http.ResponseWriter, r *http.Request) {
		user := getUser(r)
		if user == nil {
			jsonWrite(http.StatusBadRequest, w, ErrorResponse{From: "authorization", Message: "require token to get servers"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		js := json.NewEncoder(w)
		js.SetIndent("", "  ")

		config := getConfig(r)
		servers, err := config.Servers.ByOwner(user.ID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			js.Encode(ErrorResponse{From: "error", Message: err.Error()})
			return
		}
		w.WriteHeader(http.StatusOK)
		js.Encode(servers)
	})

	apiRouter.NotFound(func(w http.ResponseWriter, r *http.Request) {
		jsonWrite(http.StatusNotFound, w, ErrorResponse{From: "api path not found"})
	})
	apiRouter.MethodNotAllowed(apiRouter.NotFoundHandler())
}
