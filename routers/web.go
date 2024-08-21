package routers

import (
	"net/http"

	"sirherobrine23.com.br/go-bds/bds/routers/utils"
	webTemplates "sirherobrine23.com.br/go-bds/bds/templates"
)

var WebRoute *http.ServeMux = http.NewServeMux()

func init() {
	WebRoute.HandleFunc("GET /info", func(w http.ResponseWriter, r *http.Request) {
		utils.JsonResponse(w, 200, webTemplates.WebTemplate.Templates())
	})

	WebRoute.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		info := webTemplates.LoadTemplate("public/info.tmpl")
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
