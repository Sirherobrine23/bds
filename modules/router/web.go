package router

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"sirherobrine23.com.br/go-bds/bds/modules/datas"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/permission"
	"sirherobrine23.com.br/go-bds/bds/modules/web/templates"
	static "sirherobrine23.com.br/go-bds/bds/modules/web/web_src"
)

// Main webRouter
var webRouter = chi.NewRouter()

// Get Web dashboard render
func WebRouter(config *datas.DatabaseSchemas) (http.Handler, error) {
	webTemplates, err := templates.Templates()
	if err != nil {
		return nil, err
	}

	// Catch panic and set context with user info
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Append config and webTemplate to context
		r = r.WithContext(context.WithValue(
			context.WithValue(r.Context(), ContextConfig, config),
			ContextWebTemplates, webTemplates,
		))

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
			exist, userID, _ := config.Cookie.Cookie(cookie.Value)
			if exist {
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
		webRouter.ServeHTTP(w, r)
	}), nil
}

func init() {
	// Dashboard API
	webRouter.Mount("/api", WebApi)

	// Serve static files to client
	staticFiles := http.FileServerFS(static.StaticFiles)
	webRouter.Mount("/js", staticFiles)
	webRouter.Mount("/css", staticFiles)
	webRouter.Mount("/img", staticFiles)
	webRouter.Mount("/fonts", staticFiles)
	webRouter.Handle("GET /favicon.ico", http.RedirectHandler("/img/logo.ico", http.StatusMovedPermanently))

	// Home server
	webRouter.Get("/", func(w http.ResponseWriter, r *http.Request) {
		if getUser(r) != nil {
			http.Redirect(w, r, "/servers", http.StatusSeeOther) // redirect to servers list
			return
		}

		webTemplates := getTemplates(r)
		webTemplates.Render("public/home.tmpl", w, &templates.RenderData{
			Title:         "Home",
			Lang:          "en-us",
			PageIsInstall: false,
			User:          getUser(r),
			External:      map[string]any{},
		})
	})

	// Auth page
	webRouter.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if getUser(r) != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		webTemplates := getTemplates(r)
		config := getConfig(r)
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

	webRouter.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if getUser(r) != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		webTemplates := getTemplates(r)
		config := getConfig(r)
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

	// Global 404 page error
	webRouter.NotFound(func(w http.ResponseWriter, r *http.Request) {
		user := getUser(r)
		getTemplates(r).Render404(w, &templates.RenderData{
			Title: "Page not found",
			Lang:  "en-us",
			User:  user,
			External: map[string]any{
				"ErrorMsg": "Page request not found",
			},
		})
	})

	webRouter.MethodNotAllowed(webRouter.NotFoundHandler())
}
