package routers

import (
	"errors"
	"net/http"

	"sirherobrine23.com.br/go-bds/bds/routers/utils"
	webTemplates "sirherobrine23.com.br/go-bds/bds/templates"
)

var WebRoute *http.ServeMux = http.NewServeMux()

func init() {
	WebRoute.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/img/logo.ico", http.StatusMovedPermanently)
	})

	WebRoute.HandleFunc("GET /status", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		webTemplates.StatusTemplate(w, false, errors.New("example error"))
	})
	WebRoute.HandleFunc("POST /login", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte("not implemented"))
	})

	WebRoute.HandleFunc("GET /login", func(w http.ResponseWriter, r *http.Request) {
		info := webTemplates.LoadTemplate("users/auth/signin.tmpl")
		if info == nil {
			utils.JsonResponse(w, 500, map[string]string{"error": "not found"})
			return
		}

		if err := info.Execute(w, map[string]any{"Title": "test"}); err != nil {
			utils.JsonResponse(w, 500, map[string]string{"error": err.Error()})
			return
		}
	})

	WebRoute.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
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
