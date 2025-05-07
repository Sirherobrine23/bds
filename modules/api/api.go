// Module only api structs
package api

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
	*datas.DatabaseSchemas
}

type AppVersion struct {
	Version string        `json:"version"`
	Uptime  time.Duration `json:"uptime"`
}

func getUser(r *http.Request) user.User {
	switch v := r.Context().Value(ContextUser).(type) {
	case user.User:
		return v
	default:
		return nil
	}
}

func getTokenPerm(r *http.Request) permission.Permission {
	switch v := r.Context().Value(ContextTokenPerm).(type) {
	case permission.Permission:
		return v
	default:
		return permission.Unknown
	}
}

// Mount router to /api
func MountRouter(config *RouteConfig) (http.Handler, error) {
	api := chi.NewMux()
	api.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		js := json.NewEncoder(w)
		js.SetIndent("", "  ")
		js.Encode(AppVersion{
			Version: app.AppVersion,
			Uptime:  time.Since(app.StartTime),
		})
	})

	api.Get("/servers", func(w http.ResponseWriter, r *http.Request) {
		user := getUser(r)
		if user == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			js := json.NewEncoder(w)
			js.SetIndent("", "  ")
			js.Encode(ErrorResponse{From: "authorization", Message: "require token to get servers"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		js := json.NewEncoder(w)
		js.SetIndent("", "  ")

		servers, err := config.Servers.ByOwner(user.ID())
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			js.Encode(ErrorResponse{From: "error", Message: err.Error()})
			return
		}
		w.WriteHeader(http.StatusOK)
		js.Encode(servers)
	})

	api.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		js := json.NewEncoder(w)
		js.SetIndent("", "  ")
		js.Encode(ErrorResponse{From: "api path not found"})
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
		api.ServeHTTP(w, r)
	}), nil
}
