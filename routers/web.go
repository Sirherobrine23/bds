package routers

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	db "sirherobrine23.com.br/go-bds/bds/modules/database"
	"sirherobrine23.com.br/go-bds/bds/modules/versions"
	"sirherobrine23.com.br/go-bds/bds/routers/utils"
	webTemplates "sirherobrine23.com.br/go-bds/bds/templates"
)

const minPasswordLength = 8

var (
	hasUpperRegex   = regexp.MustCompile(`[A-Z]`)
	hasSpecialRegex = regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>/?~]`)
	usernameCheck   = regexp.MustCompile(`^[a-zA-Z0-9_.-]{3,20}$`)

	CookieName                = "bdscookie"
	WebRoute   *http.ServeMux = http.NewServeMux()
)

func isValidPasswordCombinedCheck(password string) bool {
	if len(password) < minPasswordLength {
		return false
	}
	if !hasUpperRegex.MatchString(password) {
		return false
	}
	if !hasSpecialRegex.MatchString(password) {
		return false
	}
	return true
}

func init() {
	WebRoute.Handle("GET /favicon.ico", http.RedirectHandler("/img/logo.ico", http.StatusMovedPermanently))

	WebRoute.HandleFunc("GET /user", func(w http.ResponseWriter, r *http.Request) {})

	WebRoute.HandleFunc("GET /user/img", func(w http.ResponseWriter, r *http.Request) {
		userInfo := GetUserCtx(r)
		if userInfo == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		RandomPng(w)
	})

	// Auth page
	WebRoute.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if GetCookieCtx(r) != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		singinPage := webTemplates.LoadTemplate("users/auth/signin.tmpl")
		if r.Method == "POST" {
			switch r.Header.Get("Content-Type") {
			case "application/x-www-form-urlencoded":
				if err := r.ParseForm(); err != nil {
					println(err.Error())
					singinPage.Execute(w, map[string]any{"Title": "Login", "Error": fmt.Sprintf("cannot parse Form: %s", err)})
					return
				}
			default:
				singinPage.Execute(w, map[string]any{"Title": "Login"})
				return
			}

			username, password := r.Form.Get("username"), r.Form.Get("password")
			if !usernameCheck.MatchString(username) || !isValidPasswordCombinedCheck(password) {
				singinPage.Execute(w, map[string]any{"Title": "Login", "Error": "invalid username or password"})
				return
			}

			var Auth []db.Token
			if err := db.DatabaseConnection.Join("FULL", "User", "User.ID = Token.user_id").Where("token.token == \"\" AND User.Username == ?", username).Find(&Auth); err != nil || len(Auth) == 0 {
				errMessage := "Username not exists"
				if err != nil {
					errMessage += ": " + err.Error()
				}
				singinPage.Execute(w, map[string]any{"Title": "Login", "Error": errMessage})
				return
			}

			UserAuth := Auth[0]
			if err := UserAuth.Compare(password); err != nil {
				singinPage.Execute(w, map[string]any{"Title": "Login", "Error": err.Error()})
				return
			}

			var newCookie db.Cookie
			newCookie.User = Auth[0].User
			if err := newCookie.SetupCookie(); err != nil {
				singinPage.Execute(w, map[string]any{"Title": "test"})
				return
			}
			http.SetCookie(w, &http.Cookie{Name: CookieName, Value: newCookie.CookieValue, MaxAge: int(newCookie.ValidAt.Sub(time.Now().UTC()))})
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		singinPage.Execute(w, map[string]any{"Title": "Login"})
	})

	WebRoute.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if GetCookieCtx(r) != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		registerPage := webTemplates.LoadTemplate("users/auth/register.tmpl")
		if r.Method == "POST" {
			if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
				if err := r.ParseForm(); err == nil {
					name, username, email, password := r.Form.Get("name"), r.Form.Get("username"), r.Form.Get("email"), r.Form.Get("password")
					if name == "" || len(strings.Fields(name)) <= 1 {
						registerPage.Execute(w, map[string]any{"Title": "Register new user", "Error": "Set full name", "Name": name, "Username": username, "Email": email})
						return
					}
					if !usernameCheck.MatchString(username) {
						registerPage.Execute(w, map[string]any{"Title": "Register new user", "Error": "Username require mostly complexe, Example: Google@1234@", "Name": name, "Username": username, "Email": email})
						return
					}
					if !isValidPasswordCombinedCheck(password) {
						registerPage.Execute(w, map[string]any{"Title": "Register new user", "Error": "Password require mostly complexe, Example: Google@1234@", "Name": name, "Username": username, "Email": email})
						return
					}

					newUser := &db.User{
						Name:     name,
						Username: username,
						Email:    email,
						Banned:   false,
						Active:   true,
					}

					if err := newUser.CreateUser(); err != nil {
						registerPage.Execute(w, map[string]any{"Title": "Register new user", "Error": fmt.Sprintf("cannot set User in database: %s", err), "Name": name, "Username": username, "Email": email})
						return
					}

					_, err := db.CreatePassword(password, newUser)
					if err != nil {
						db.DatabaseConnection.Delete(newUser)
						registerPage.Execute(w, map[string]any{"Title": "Register new user", "Error": fmt.Sprintf("cannot set password in DB: %s", err), "Name": name, "Username": username, "Email": email})
						return
					}

					var newCookie db.Cookie
					newCookie.User = newUser
					if err = newCookie.SetupCookie(); err != nil {
						http.Redirect(w, r, "/login", http.StatusSeeOther)
						return
					}

					http.SetCookie(w, &http.Cookie{Name: CookieName, Value: newCookie.CookieValue, MaxAge: int(time.Now().Sub(newCookie.ValidAt))})
					http.Redirect(w, r, "/", http.StatusSeeOther)
					return
				}
			}
		}

		registerPage.Execute(w, map[string]any{"Title": "Register new user"})
	})

	WebRoute.HandleFunc("GET /servers", func(w http.ResponseWriter, r *http.Request) {
		userInfo := GetUserCtx(r)
		if userInfo == nil {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		var serversList []*db.MinecraftServers
		if err := db.DatabaseConnection.Where("MinecraftServers.user == ?", userInfo.ID).Find(&serversList); err != nil {
			w.WriteHeader(500)
			webTemplates.StatusTemplate(w, false, err)
			return
		}

		info := webTemplates.LoadTemplate("server/servers.tmpl")
		ModelOptions := map[string]any{
			"Title":   fmt.Sprintf("%s Servers", userInfo.Username),
			"Signed":  userInfo != nil,
			"User":    userInfo,
			"Servers": serversList,
		}

		info.Execute(w, ModelOptions)
	})

	WebRoute.HandleFunc("/servers/new", func(w http.ResponseWriter, r *http.Request) {
		userInfo := GetUserCtx(r)
		if userInfo == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		info := webTemplates.LoadTemplate("server/servers_new.tmpl")
		ModelOptions := map[string]any{
			"Title":  "Create new Server",
			"Signed": true,
			"User":   userInfo,
		}

		if r.Method == "POST" {
			if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
				if err := r.ParseForm(); err != nil {
					ModelOptions["Error"] = err.Error()
					info.Execute(w, ModelOptions)
					return
				}

				serverType, serverName := r.Form.Get("server"), r.Form.Get("servername")
				if !usernameCheck.MatchString(serverName) {
					ModelOptions["Error"] = "Set valid name to server"
					info.Execute(w, ModelOptions)
					return
				}

				switch serverType {
				case "bedrock", "java", "spigot", "purpur", "paper", "folia", "velocity":
				default:
					ModelOptions["Error"] = fmt.Sprintf("Invalid server type input: %s", serverType)
					info.Execute(w, ModelOptions)
					return
				}

				exist, err := db.DatabaseConnection.Where("user == ? AND server == ? AND name == ?", userInfo.ID, serverType, serverName).Exist(&db.MinecraftServers{})
				if err != nil {
					ModelOptions["Error"] = err.Error()
					info.Execute(w, ModelOptions)
					return
				} else if exist {
					ModelOptions["Error"] = "Server ared exist"
					info.Execute(w, ModelOptions)
					return
				}

				newServer := &db.MinecraftServers{
					ServerType: serverType,
					User:       userInfo,
					Name:       serverName,
				}

				switch serverType {
				case "bedrock":
					ver := versions.BedrockVersions.LatestStable()
					newServer.Version = ver.Version
				case "java":
					ver := versions.JavaVersions[len(versions.JavaVersions)-1]
					newServer.Version = ver.Version()
				case "spigot":
					ver := versions.SpigotVersions[len(versions.SpigotVersions)-1]
					newServer.Version = ver.Version()
				case "purpur":
					ver := versions.PurpurVersions[len(versions.PurpurVersions)-1]
					newServer.Version = ver.Version()
				case "paper":
					ver := versions.PaperVersions[len(versions.PaperVersions)-1]
					newServer.Version = ver.Version()
				case "folia":
					ver := versions.FoliaVersions[len(versions.FoliaVersions)-1]
					newServer.Version = ver.Version()
				case "velocity":
					ver := versions.VelocityVersions[len(versions.VelocityVersions)-1]
					newServer.Version = ver.Version()
				}

				// Insert to database
				if _, err := db.DatabaseConnection.InsertOne(newServer); err != nil {
					ModelOptions["Error"] = fmt.Sprintf("cannot insert in database new server: %s", err)
					info.Execute(w, ModelOptions)
					return
				}

				// Redirect client to admin page
				http.Redirect(w, r, fmt.Sprintf("/server/%d", newServer.ServerID), http.StatusSeeOther)
				return
			}
		}
		info.Execute(w, ModelOptions)
	})

	WebRoute.HandleFunc("GET /server/{serverTarget}", func(w http.ResponseWriter, r *http.Request) {
		userInfo := GetUserCtx(r)
		if userInfo == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		Server := db.MinecraftServers{}
		serverTarget := r.PathValue("serverTarget")

		if serverID, err := strconv.ParseInt(serverTarget, 10, 64); err == nil {
			if _, err := db.DatabaseConnection.Where("user == ? AND id == ?", userInfo.ID, serverID).Get(&Server); err != nil {
				w.WriteHeader(http.StatusNotFound)
				webTemplates.StatusTemplate404(w, true, "Server not exists")
				return
			}
		} else {
			if _, err := db.DatabaseConnection.Where("user == ? AND name == ?", userInfo.ID, serverTarget).Get(&Server); err != nil {
				w.WriteHeader(http.StatusNotFound)
				webTemplates.StatusTemplate404(w, true, "Server not exists")
				return
			}
		}

		info := webTemplates.LoadTemplate("server/server/control.tmpl")
		ModelOptions := map[string]any{
			"Title":  fmt.Sprintf("Control - %s", Server.Name),
			"Signed": true,
			"User":   userInfo,
			"Server": Server,
		}

		info.Execute(w, ModelOptions)
	})

	WebRoute.HandleFunc("GET /server/{serverTarget}/console", func(w http.ResponseWriter, r *http.Request) {})
	WebRoute.HandleFunc("GET /server/{serverTarget}/settings", func(w http.ResponseWriter, r *http.Request) {})
	WebRoute.HandleFunc("GET /server/{serverTarget}/software", func(w http.ResponseWriter, r *http.Request) {})
	WebRoute.HandleFunc("GET /server/{serverTarget}/players", func(w http.ResponseWriter, r *http.Request) {})
	WebRoute.HandleFunc("GET /server/{serverTarget}/files", func(w http.ResponseWriter, r *http.Request) {})

	WebRoute.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		userInfo := GetUserCtx(r)
		info := webTemplates.LoadTemplate("public/home.tmpl")
		if info == nil {
			utils.JsonResponse(w, 500, map[string]string{"error": "not found"})
			return
		}

		ModelOptions := map[string]any{
			"Title":  "Home page",
			"Signed": userInfo != nil,
			"User":   userInfo,
		}

		if err := info.Execute(w, ModelOptions); err != nil {
			utils.JsonResponse(w, 500, map[string]string{"error": err.Error()})
			return
		}
	})
}
