package router

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/permission"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/server"
	"sirherobrine23.com.br/go-bds/bds/modules/web/templates"
)

func init() {
	webRouter.Route("/servers", func(webRouter chi.Router) {
		webRouter.Get("/", func(w http.ResponseWriter, r *http.Request) {
			user := getUser(r)
			if user == nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			webTemplates := getTemplates(r)
			config := getConfig(r)
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

		webRouter.Get("/new", func(w http.ResponseWriter, r *http.Request) {
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
			getTemplates(r).Render("server/new_server.tmpl", w, pageConfig)
		})

		webRouter.Post("/new", func(w http.ResponseWriter, r *http.Request) {
			user := getUser(r)
			if user == nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			webTemplates := getTemplates(r)
			config := getConfig(r)
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

		webRouter.Get("/{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
			user := getUser(r)
			if user == nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			webTemplates := getTemplates(r)
			config := getConfig(r)
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
			webTemplates.Render("server/server/home.tmpl", w, pageConfig)
		})
	})
}
