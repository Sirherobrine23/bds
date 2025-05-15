package router

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"sirherobrine23.com.br/go-bds/bds/modules/web/templates"
)

func init() {
	webRouter.Route("/users", func(webRouter chi.Router) {
		webRouter.Get("/{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
			currentUser, webTemplates := getUser(r), getTemplates(r)
			pageConfig := &templates.RenderData{User: currentUser, External: map[string]any{}, Title: "Unknown", Lang: "en-us"}
			if currentUser != nil {
				pageConfig.Title = currentUser.Name
			}

			id, _ := strconv.ParseInt(r.PathValue("id"), 8, 64)
			user, err := getConfig(r).User.ByID(id)
			if err != nil {
				webTemplates.Render404(w, nil)
				return
			}

			pageConfig.External["GetUser"] = user
			webTemplates.Render("users/user.tmpl", w, pageConfig)
		})
		webRouter.Get("/{id:[0-9]+}/img", func(w http.ResponseWriter, r *http.Request) {})
	})
}
