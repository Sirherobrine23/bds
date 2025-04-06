package routers

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	db "sirherobrine23.com.br/go-bds/bds/modules/database"
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
	WebRoute.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/img/logo.ico", http.StatusMovedPermanently)
	})

	WebRoute.HandleFunc("GET /status", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		webTemplates.StatusTemplate(w, false, errors.New("example error"))
	})

	// Auth page
	WebRoute.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if cookie, _ := r.Cookie(CookieName); cookie != nil {
			var cookies []db.Cookie
			if err := db.DatabaseConnection.Where("cookie.cookie == ?", cookie.Value).Find(&cookies); err == nil && len(cookies) > 0 {
				http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
				return
			}
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
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		singinPage.Execute(w, map[string]any{"Title": "Login"})
	})

	WebRoute.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if cookie, _ := r.Cookie(CookieName); cookie != nil {
			var cookies []db.Cookie
			if err := db.DatabaseConnection.Where("cookie.cookie == ?", cookie.Value).Find(&cookies); err == nil && len(cookies) > 0 {
				http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
				return
			}
		}

		registerPage := webTemplates.LoadTemplate("users/auth/register.tmpl")
		if r.Method == "POST" {
			switch r.Header.Get("Content-Type") {
			case "application/x-www-form-urlencoded":
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
						http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
						return
					}

					http.SetCookie(w, &http.Cookie{Name: CookieName, Value: newCookie.CookieValue, MaxAge: int(time.Now().Sub(newCookie.ValidAt))})
					http.Redirect(w, r, "/", http.StatusFound)
					return
				}
			}
		}

		registerPage.Execute(w, map[string]any{"Title": "Register new user"})
	})

	WebRoute.HandleFunc("GET /server/{id}", func(w http.ResponseWriter, r *http.Request) {
		info := webTemplates.LoadTemplate("server/server/control.tmpl")
		if info == nil {
			utils.JsonResponse(w, 500, map[string]string{"error": "not found"})
			return
		}

		if err := info.Execute(w, map[string]any{"Title": "test"}); err != nil {
			utils.JsonResponse(w, 500, map[string]string{"error": err.Error()})
			return
		}
	})

	WebRoute.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		info := webTemplates.LoadTemplate("public/home.tmpl")
		if info == nil {
			utils.JsonResponse(w, 500, map[string]string{"error": "not found"})
			return
		}

		if err := info.Execute(w, map[string]any{"Title": "test"}); err != nil {
			utils.JsonResponse(w, 500, map[string]string{"error": err.Error()})
			return
		}
	})
}
