package routers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"sirherobrine23.com.br/go-bds/bds/modules/config"
	db "sirherobrine23.com.br/go-bds/bds/modules/database"
	web "sirherobrine23.com.br/go-bds/bds/web_src"
)

var (
	ContextCookie = "bdscookie"
	ContextUser   = "bdsuser"
	ContextToken  = "bdstoken"

	Router *http.ServeMux = http.NewServeMux() // Server Handler
)

type ServerConfig struct {
	Port         int    `ini:"PORT" json:"port"`
	PortRedirect int    `ini:"PORT_REDIRECT" json:"portRedirect"`
	ListenHTTPs  bool   `ini:"HTTPS" json:"listenHTTPs"`
	CertFile     string `ini:"CERT" json:"-"`
	KeyFile      string `ini:"KEY" json:"-"`
}

func Listen() error {
	server, err := config.ConfigProvider.GetSection("server")
	if err != nil {
		if server, err = config.ConfigProvider.NewSection("server"); err != nil {
			return err
		}
		server.NewKey("PORT", "3000")
		server.NewKey("HTTPS", "false")
	}
	server.Key("PORT").MustInt(3000)
	var dataConfig ServerConfig
	if err = server.MapTo(&dataConfig); err != nil {
		return err
	}

	listAddr, redirectAddr := fmt.Sprintf(":%d", dataConfig.Port), fmt.Sprintf(":%d", dataConfig.PortRedirect)
	fmt.Printf("Listen on %s\n", listAddr)
	if dataConfig.ListenHTTPs {
		// Redirect handler
		redirect := http.NewServeMux()
		redirect.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			head := w.Header()
			head.Set("Location", "https://"+r.Host+r.RequestURI)
			w.WriteHeader(http.StatusMovedPermanently)
		})

		fmt.Printf("Listen redirect on %s\n", listAddr)
		go http.ListenAndServe(redirectAddr, redirect)
		return http.ListenAndServeTLS(listAddr, dataConfig.CertFile, dataConfig.KeyFile, Router)
	}

	return http.ListenAndServe(listAddr, Router)
}

func init() {
	// API Path
	Router.Handle("/api/v1/", ContextApi(http.StripPrefix("/api/v1", APIv1)))

	// Static files
	statisFiles := http.FileServer(http.FS(web.StatisFiles))
	Router.Handle("GET /js/", statisFiles)
	Router.Handle("GET /fonts/", statisFiles)
	Router.Handle("GET /css/", statisFiles)
	Router.Handle("GET /img/", statisFiles)
	Router.Handle("/", ContextWeb(WebRoute))
}

func GetUserCtx(r *http.Request) *db.User {
	user := r.Context().Value(ContextUser)
	switch user := user.(type) {
	case db.User:
		return &user
	case *db.User:
		return user
	default:
		return nil
	}
}

func GetCookieCtx(r *http.Request) *db.Cookie {
	cookie := r.Context().Value(ContextCookie)
	switch cookie := cookie.(type) {
	case db.Cookie:
		return &cookie
	case *db.Cookie:
		return cookie
	default:
		return nil
	}
}

func ContextWeb(fn http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cookie, _ := r.Cookie(CookieName); cookie != nil {
			var cookies []*db.Cookie
			if err := db.DatabaseConnection.Join("FULL", "User", "User.ID = Cookie.user").Where("cookie.cookie == ?", cookie.Value).Find(&cookies); err == nil && len(cookies) > 0 {
				cookie := cookies[0]
				oldCtx := r.Context()
				r = r.WithContext(context.WithValue(context.WithValue(oldCtx, ContextCookie, cookie), ContextUser, cookie.User))
			}
		}
		fn.ServeHTTP(w, r)
	})
}

func ContextApi(fn http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if Auth := r.Header.Get("Authorization"); Auth != "" {
			switch {
			case strings.HasPrefix(Auth, "Bearer"):
				Auth = strings.TrimPrefix(Auth, "Bearer")
				var tokens []*db.Token
				if err := db.DatabaseConnection.Join("FULL", "User", "User.ID = Token.user_id").Where("cookie.token == ?", Auth).Find(&tokens); err == nil && len(tokens) > 0 {
					token := tokens[0]
					oldCtx := r.Context()
					r = r.WithContext(context.WithValue(context.WithValue(oldCtx, ContextToken, token), ContextUser, token.User))
				}
			}
		}
		fn.ServeHTTP(w, r)
	})
}
