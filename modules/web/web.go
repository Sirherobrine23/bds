package web

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"sirherobrine23.com.br/go-bds/bds/modules/api"
	"sirherobrine23.com.br/go-bds/bds/modules/datas"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/permission"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/server"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/user"
	"sirherobrine23.com.br/go-bds/bds/modules/web/templates"
	static "sirherobrine23.com.br/go-bds/bds/modules/web/web_src"
)

const (
	ContextUser = "ctxUser"
	CookieName  = "bdsCookie"
)

func getUser(r *http.Request) *user.User {
	switch v := r.Context().Value(ContextUser).(type) {
	case *user.User:
		return v
	default:
		return nil
	}
}

type WebConfig struct {
	*datas.DatabaseSchemas
}

// Mount router to /api
func MountRouter(config *WebConfig) (http.Handler, error) {
	webTemplates, err := templates.Templates()
	if err != nil {
		return nil, err
	}

	api, err := api.MountRouter(&api.RouteConfig{DatabaseSchemas: config.DatabaseSchemas})
	if err != nil {
		return nil, err
	}

	// Start new handler
	router := chi.NewMux()
	router.Mount("/api", api)

	// Serve static files to client
	staticFiles := http.FileServerFS(static.StaticFiles)
	router.Mount("/js", staticFiles)
	router.Mount("/css", staticFiles)
	router.Mount("/img", staticFiles)
	router.Mount("/fonts", staticFiles)
	router.Handle("GET /favicon.ico", http.RedirectHandler("/img/logo.ico", http.StatusMovedPermanently))

	// Home server
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		webTemplates.Render("public/home.tmpl", w, &templates.RenderData{
			Title:         "Home",
			Lang:          "en-us",
			PageIsInstall: false,
			User:          getUser(r),
			External:      map[string]any{},
		})
	})

	// Auth page
	router.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if getUser(r) != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		pageConfig := &templates.RenderData{External: map[string]any{}}
		if r.Method == http.MethodPost {
			var username, password string

			switch r.Header.Get("Content-Type") {
			case "application/x-www-form-urlencoded":
				if err := r.ParseForm(); err != nil {
					println(err.Error())
					pageConfig.External["Error"] = fmt.Sprintf("cannot parse Form body: %s", err)
					webTemplates.Render("users/auth/signin.tmpl", w, pageConfig)
					return
				}

				// Get username and password
				username, password = r.Form.Get("username"), r.Form.Get("password")
			default:
				webTemplates.Render("users/auth/signin.tmpl", w, pageConfig)
				return
			}

			user, err := config.User.Username(username)
			if err != nil {
				pageConfig.External["Error"] = fmt.Sprintf("User not exist: %s", err)
				webTemplates.Render("users/auth/signin.tmpl", w, pageConfig)
				return
			}

			// Ignore if disabled user
			if user.Permission == permission.Unknown {
				pageConfig.External["Error"] = "Unauthorized login"
				webTemplates.Render("users/auth/signin.tmpl", w, pageConfig)
				return
			}

			ok, err := user.Password.Check(password)
			switch {
			case ok:
				var newCookie string
				if newCookie, err = config.Cookie.CreateCookie(user.ID); err == nil {
					http.SetCookie(w, &http.Cookie{Name: CookieName, Value: newCookie})
					http.Redirect(w, r, "/", http.StatusSeeOther)
					return
				}
				fallthrough
			default:
				pageConfig.External["Error"] = "Cannot auth"
				if err != nil {
					pageConfig.External["Error"] = fmt.Sprintf("Cannot auth: %s", err)
				}
			}
		}

		webTemplates.Render("users/auth/signin.tmpl", w, pageConfig)
	})

	router.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if getUser(r) != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		pageConfig := &templates.RenderData{External: map[string]any{}}
		if r.Method == http.MethodPost {
			var name, username, email, password string
			switch r.Header.Get("Content-Type") {
			case "application/x-www-form-urlencoded":
				if err := r.ParseForm(); err != nil {
					println(err.Error())
					pageConfig.External["Error"] = fmt.Sprintf("cannot parse Form body: %s", err)
					webTemplates.Render("users/auth/register.tmpl", w, pageConfig)
					return
				}
				name, username, email, password = r.Form.Get("name"), r.Form.Get("username"), r.Form.Get("email"), r.Form.Get("password")
			default:
				webTemplates.Render("users/auth/register.tmpl", w, pageConfig)
				return
			}

			// Create user
			user, err := config.User.Create(name, username, email, password)
			if err != nil {
				pageConfig.External["Error"] = fmt.Sprintf("Cannot create new user: %s", err)
				webTemplates.Render("users/auth/register.tmpl", w, pageConfig)
				return
			}

			// Set cookie
			redirectTo := "/login"
			if newCookie, err := config.Cookie.CreateCookie(user.ID); err == nil {
				http.SetCookie(w, &http.Cookie{Name: CookieName, Value: newCookie})
				redirectTo = "/"
			}

			http.Redirect(w, r, redirectTo, http.StatusSeeOther)
			return
		}
		webTemplates.Render("users/auth/register.tmpl", w, pageConfig)
	})

	router.Get("/servers", func(w http.ResponseWriter, r *http.Request) {
		user := getUser(r)
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		pageConfig := &templates.RenderData{
			Lang:  "en-us",
			User:  user,
			Title: fmt.Sprintf("%s Servers", user.Username),
			External: map[string]any{
				"Servers": []*server.Server{},
			},
		}

		servers, err := config.Servers.ByOwner(user.ID)
		if err != nil {
			pageConfig.External["Error"] = err.Error()
			webTemplates.Render400(w, pageConfig)
			return
		}

		pageConfig.External["Servers"] = servers
		webTemplates.Render("server/server_list.tmpl", w, pageConfig)
	})

	router.Get("/servers/new", func(w http.ResponseWriter, r *http.Request) {
		user := getUser(r)
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		pageConfig := &templates.RenderData{
			Lang:     "en-us",
			User:     user,
			Title:    fmt.Sprintf("%s new server", user.Username),
			External: map[string]any{},
		}
		webTemplates.Render("server/new_server.tmpl", w, pageConfig)
	})

	router.Post("/servers/new", func(w http.ResponseWriter, r *http.Request) {
		user := getUser(r)
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if !user.Permission.IsRoot() && !user.Permission.Contains(permission.WebCreateServer) {
			webTemplates.Render400(w, &templates.RenderData{
				Title:         "You not have permission to create server",
				Lang:          "en-us",
				PageIsInstall: false,
				User:          user,
				External: map[string]any{
					"Error": "you not have permission to create server, contact admin",
				},
			})
			return
		}

		pageConfig := &templates.RenderData{
			Lang:     "en-us",
			User:     user,
			Title:    fmt.Sprintf("%s new server", user.Username),
			External: map[string]any{},
		}

		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			pageConfig.External["Error"] = "invalid body"
			webTemplates.Render("server/new_server.tmpl", w, pageConfig)
			return
		}

		if err := r.ParseForm(); err != nil {
			pageConfig.External["Error"] = err.Error()
			webTemplates.Render("server/new_server.tmpl", w, pageConfig)
			return
		}

		switch v := r.Form.Get("server"); v {
		case "bedrock", "java", "pocketmine", "spigot", "purpur", "paper", "folia", "velocity":
		default:
			pageConfig.External["Error"] = fmt.Sprintf("Invalid server type input: %s", v)
			webTemplates.Render("server/new_server.tmpl", w, pageConfig)
			return
		}

		var serverType server.ServerType
		switch r.Form.Get("server") {
		case "bedrock":
			serverType = server.Bedrock
		case "java":
			serverType = server.Java
		case "pocketmine":
			serverType = server.Pocketmine
		case "spigot":
			serverType = server.SpigotMC
		case "purpur":
			serverType = server.PurpurMC
		case "paper":
			serverType = server.PaperMC
		case "folia":
			serverType = server.FoliaMC
		case "velocity":
			serverType = server.VelocityMC
		}

		serverInfo, err := config.Servers.CreateServer(r.Form.Get("servername"), "latest", serverType, user)
		if err != nil {
			pageConfig.External["Error"] = fmt.Sprintf("Cannot make new server: %s", err)
			webTemplates.Render("server/new_server.tmpl", w, pageConfig)
			return
		}

		// Redirect client to admin page
		http.Redirect(w, r, fmt.Sprintf("/servers/%d", serverInfo.ID), http.StatusSeeOther)
	})

	router.Get("/servers/{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		user := getUser(r)
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		pageConfig := &templates.RenderData{User: user, External: map[string]any{}, Title: "Unknown Server", Lang: "en-us"}

		serverID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			webTemplates.Render404(w, pageConfig)
			return
		}

		server, err := config.Servers.ByID(serverID)
		if err != nil {
			pageConfig.External["Error"] = fmt.Sprintf("Server not exist or error: %s", err)
			webTemplates.Render400(w, pageConfig)
			return
		}

		// Check is avaible to edit
		if !user.Permission.IsRoot() {
			haveUser, ok := server.Owners.UserID(user.ID)
			if !ok || (!haveUser.Permission.Contains(permission.ServerOwner) && !haveUser.Permission.Contains(permission.ServerEdit|permission.ServerView)) {
				pageConfig.External["Error"] = "Not have permission to access this server"
				webTemplates.Render404(w, pageConfig)
				return
			}
		}

		pageConfig.External["Server"] = server
		webTemplates.Render("server/server.tmpl", w, pageConfig)
	})

	// Global 404 page error
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		user := getUser(r)
		webTemplates.Render404(w, &templates.RenderData{
			Title: "Page not found",
			Lang:  "en-us",
			User:  user,
			External: map[string]any{
				"ErrorMsg": "Page request not found",
			},
		})
	})

	router.MethodNotAllowed(router.NotFoundHandler())

	// Catch panic and set context with user info
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			user := getUser(r)
			if err := recover(); err != nil {
				webTemplates.Render5xx(w, &templates.RenderData{
					Title: "Internal error",
					Lang:  "en-us",
					User:  user,
					External: map[string]any{
						"ErrorMsg": fmt.Errorf("backend error: %s", err).Error(),
					},
				})
			}
		}()

		if cookie, err := r.Cookie(CookieName); err == nil {
			exist, userID, err := config.Cookie.Cookie(cookie.Value)
			if err != nil {
				webTemplates.Render5xx(w, &templates.RenderData{
					Title: "Internal error",
					Lang:  "en-us",
					External: map[string]any{
						"ErrorMsg": fmt.Errorf("backend error: %s", err).Error(),
					},
				})
				return
			} else if exist {
				user, err := config.User.ByID(userID)
				if err != nil {
					webTemplates.Render5xx(w, &templates.RenderData{
						Title: "Internal error",
						Lang:  "en-us",
						External: map[string]any{
							"ErrorMsg": fmt.Errorf("backend error: %s", err).Error(),
						},
					})
					return
				}
				r = r.WithContext(context.WithValue(r.Context(), ContextUser, user))
			}
		}

		// Caller api router handler
		router.ServeHTTP(w, r)
	}), nil
}
