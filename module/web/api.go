package web

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"sirherobrine23.com.br/go-bds/bds/module/db"
	"sirherobrine23.com.br/go-bds/bds/module/server"
)

// Base of API router
var API = chi.NewMux()

// Add this if API only avaible
func ApiCaller(database db.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add database to context
		r = r.WithContext(context.WithValue(r.Context(), DatabaseContext, database))
		API.ServeHTTP(w, r)
	})
}

// Default function to response with JSON body
func jsonResponse(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	js := json.NewEncoder(w)
	js.SetIndent("", "  ")
	_ = js.Encode(body) // Ignore error
}

func init() {
	// Default router response
	API.NotFound(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, http.StatusNotFound, map[string]string{"error": "router not found", "message": "router not found, check if path is valid"})
	})
	API.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed", "message": "method not allowed, check if method is valid"})
	})

	API.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			database := Database(r.Context())
			if database == nil {
				jsonResponse(w, http.StatusInternalServerError, map[string]string{
					"error":   "database connection",
					"message": "invalid server configuration or caller, check implementaion",
				})
				return
			}

			if Authorization := r.Header.Get("Authorization"); Authorization != "" {
				if !(strings.HasPrefix(strings.ToLower(Authorization), "bearer ") || strings.HasPrefix(strings.ToLower(Authorization), "token ")) {
					jsonResponse(w, http.StatusUnauthorized, map[string]string{
						"error":   "basic auth",
						"message": "basic authentication is disabled, use bearer or token prefix",
					})
					return
				}
				Authorization = strings.TrimPrefix(strings.TrimPrefix(Authorization, "token "), "bearer ")
				token, user, err := database.Token(Authorization)
				if err != nil {
					jsonResponse(w, http.StatusUnauthorized, map[string]string{
						"error":   "auth",
						"message": err.Error(),
					})
					return
				}
				ctx := context.WithValue(r.Context(), TokenContext, token)
				ctx = context.WithValue(ctx, UserContext, user)
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r) // call next router
		})
	})

	API.Route("/server/{id:[0-9]+}", func(API chi.Router) {
		API.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				token := Token(r.Context())
				if token == nil {
					jsonResponse(w, http.StatusUnauthorized, map[string]string{
						"error":   "authoraztion",
						"message": "require token to access this route",
					})
					return
				}
				serverID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
				db := Database(r.Context())
				user := User(r.Context())

				mcServer, err := db.Server(serverID)
				if err != nil {
					switch err {
					case io.EOF:
						jsonResponse(w, http.StatusNotFound, map[string]string{"error": "server not found"})
					default:
						jsonResponse(w, http.StatusInternalServerError, map[string]string{
							"error":   "internal error",
							"message": err.Error(),
						})
					}
					return
				}

				if mcServer.Owner != user.UserID {
					friends, err := db.ServerFriends(serverID)
					if err != nil {
						switch err {
						case io.EOF:
							jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "server not found"})
						default:
							jsonResponse(w, http.StatusInternalServerError, map[string]string{
								"error":   "internal error",
								"message": err.Error(),
							})
						}
						return
					}

					friendIndex := slices.IndexFunc(friends, func(friend *server.ServerFriends) bool {
						for _, perm := range friend.Permission {
							if (perm == server.View || perm == server.Edit) && friend.UserID == user.UserID {
								return true
							}
						}

						return false
					})

					if friendIndex == -1 {
						jsonResponse(w, http.StatusNotFound, map[string]string{"error": "server not found"})
						return
					}

					// Add friend
					r = r.WithContext(context.WithValue(r.Context(), ServerFriendContext, friends[friendIndex]))
				}

				// Add server to next call
				r = r.WithContext(context.WithValue(r.Context(), ServerContext, mcServer))

				next.ServeHTTP(w, r) // call next router
			})
		})

		// Get Server info
		API.Get("/", func(w http.ResponseWriter, r *http.Request) {})

		// Delete server
		API.Delete("/", func(w http.ResponseWriter, r *http.Request) {})

		// Update server
		API.Put("/", func(w http.ResponseWriter, r *http.Request) {})

		// Server config
		API.Route("/config", func(API chi.Router) {
			API.Get("/", func(w http.ResponseWriter, r *http.Request) {})
			API.Post("/", func(w http.ResponseWriter, r *http.Request) {})
		})

		// Server players
		API.Route("/players", func(API chi.Router) {
			// Get current users if avaible
			API.Get("/", func(w http.ResponseWriter, r *http.Request) {})

			API.Route("/{username}", func(API chi.Router) {
				API.Get("/", func(w http.ResponseWriter, r *http.Request) {})    // Get current status
				API.Post("/", func(w http.ResponseWriter, r *http.Request) {})   // Post new status
				API.Delete("/", func(w http.ResponseWriter, r *http.Request) {}) // Delete status
			})
		})

		API.Route("/backup", func(API chi.Router) {
			// Get all backups
			API.Get("/", func(w http.ResponseWriter, r *http.Request) {})

			// Create new backup
			API.Post("/", func(w http.ResponseWriter, r *http.Request) {})

			// Download backup
			API.Get("/{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {})

			// Delete backup
			API.Delete("/{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {})
		})
	})

	API.Route("/servers", func(API chi.Router) {
		// Get user servers
		API.Get("/", func(w http.ResponseWriter, r *http.Request) {})

		// Create new server
		API.Post("/", func(w http.ResponseWriter, r *http.Request) {})
	})

	// Get user info
	API.Get("/user", func(w http.ResponseWriter, r *http.Request) {})            // Current
	API.Get("/user/{username}", func(w http.ResponseWriter, r *http.Request) {}) // username
}
