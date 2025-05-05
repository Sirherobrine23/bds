// Module only api structs
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	app "sirherobrine23.com.br/go-bds/bds/modules"

	"sirherobrine23.com.br/go-bds/bds/modules/datas/permission"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/token"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/user"
)

const (
	ContextToken     string = "token"
	ContextTokenPerm string = "tokenPerm"
	ContextUser      string = "tokenUser"
)

// JSON message to return
type ErrorResponse struct {
	From    string `json:"error"`
	Message string `json:"message,omitempty"`
}

type RouteConfig struct {
	Token token.Token
	User  user.UserSearch
}

type AppVersion struct {
	Version string                `json:"version"`
	Uptime  time.Duration         `json:"uptime"`
	Perm    permission.Permission `json:"user_permission"`
}

// Mount router to /api
func MountRouter(config *RouteConfig) (http.Handler, error) {
	router := http.NewServeMux()

	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		js := json.NewEncoder(w)
		js.SetIndent("", "  ")
		js.Encode(AppVersion{
			Version: app.AppVersion,
			Uptime:  time.Now().Sub(app.StartTime),
			Perm:    r.Context().Value(ContextTokenPerm).(permission.Permission),
		})
	})

	// Catch panic and set context with user info
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				js := json.NewEncoder(w)
				js.SetIndent("", "  ")
				js.Encode(ErrorResponse{
					From:    "Internal Server Error",
					Message: fmt.Sprintf("error: %s", err),
				})
				return
			}
		}()

		// Set context
		if v := r.Header.Get("Authentication"); v != "" {
			switch {
			case strings.HasPrefix(v, "Bearer"), strings.HasPrefix(v, "token"):
				tokenString := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(v, "token"), "Bearer"))
				ok, userID, perm, err := config.Token.Check(tokenString)
				switch {
				case err != nil:
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)
					js := json.NewEncoder(w)
					js.SetIndent("", "  ")
					js.Encode(ErrorResponse{
						From:    "token",
						Message: err.Error(),
					})
					return
				case !ok, perm == permission.Unknown:
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)
					js := json.NewEncoder(w)
					js.SetIndent("", "  ")
					js.Encode(ErrorResponse{
						From:    "token",
						Message: "token not exists",
					})
					return
				}

				user, err := config.User.ByID(userID)
				if err != nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)
					js := json.NewEncoder(w)
					js.SetIndent("", "  ")
					js.Encode(ErrorResponse{
						From:    "user",
						Message: err.Error(),
					})
					return
				}

				ctx := r.Context()
				ctx = context.WithValue(ctx, ContextUser, user)
				ctx = context.WithValue(ctx, ContextToken, tokenString)
				ctx = context.WithValue(ctx, ContextTokenPerm, perm)
				r = r.WithContext(ctx)
			case strings.HasPrefix(v, "basic"):
				w.WriteHeader(http.StatusBadRequest)
				js := json.NewEncoder(w)
				js.SetIndent("", "  ")
				js.Encode(ErrorResponse{
					From:    "authentication",
					Message: "basic authentication is disabled",
				})
				return
			default:
				w.WriteHeader(http.StatusBadRequest)
				js := json.NewEncoder(w)
				js.SetIndent("", "  ")
				js.Encode(ErrorResponse{
					From:    "authentication",
					Message: "Require Authentication header",
				})
				return
			}
		} else {
			ctx := r.Context()
			ctx = context.WithValue(ctx, ContextTokenPerm, permission.Unknown)
			r = r.WithContext(ctx)
		}

		// Caller api router handler
		router.ServeHTTP(w, r)
	}), nil
}
