package web

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"sirherobrine23.com.br/go-bds/bds/modules/datas"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/permission"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/user"
	"sirherobrine23.com.br/go-bds/bds/modules/web/templates"
	static "sirherobrine23.com.br/go-bds/bds/modules/web/web_src"

	"fmt"
)

const (
	ContextUser = "ctxUser"
	CookieName  = "bdsCookie"
)

func getUser(r *http.Request) user.User {
	switch v := r.Context().Value(ContextUser).(type) {
	case user.User:
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

	// Start new handler
	router := chi.NewMux()

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
				println(err.Error())
				pageConfig.External["Error"] = "User not exist"
				webTemplates.Render("users/auth/signin.tmpl", w, pageConfig)
				return
			}

			// Ignore if disabled user
			if user.Permission() == permission.Unknown {
				pageConfig.External["Error"] = "User not exist"
				webTemplates.Render("users/auth/signin.tmpl", w, pageConfig)
				return
			}

			pass, err := user.Password()
			if err != nil {
				println(err.Error())
				pageConfig.External["Error"] = "cannot auth user"
				webTemplates.Render("users/auth/signin.tmpl", w, pageConfig)
				return
			}

			ok, err := pass.Check(password)
			if err != nil {
				println(err.Error())
				pageConfig.External["Error"] = "cannot auth user, password"
				webTemplates.Render("users/auth/signin.tmpl", w, pageConfig)
				return
			} else if ok {
				newCookie, err := config.Cookie.CreateCookie(user.ID())
				if err != nil {
					println(err.Error())
					pageConfig.External["Error"] = "cannot auth user, error on make cookie"
					webTemplates.Render("users/auth/signin.tmpl", w, pageConfig)
					return
				}

				http.SetCookie(w, &http.Cookie{Name: CookieName, Value: newCookie})
				http.Redirect(w, r, "/", http.StatusSeeOther)
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
		webTemplates.Render("users/auth/register.tmpl", w, pageConfig)
	})

	// Global 404 page error
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(ContextUser).(user.User)
		if !ok {
			user = nil
		}

		w.WriteHeader(http.StatusNotFound)
		webTemplates.Render404(w, &templates.RenderData{
			Title: "Page not found",
			Lang:  "en-us",
			User:  user,
			External: map[string]any{
				"ErrorMsg": "Page request not found",
			},
		})
	})

	// Catch panic and set context with user info
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			user, ok := r.Context().Value(ContextUser).(user.User)
			if !ok {
				user = nil
			}

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
