package router

import (
	"encoding/json"
	"net/http"

	"sirherobrine23.com.br/go-bds/bds/modules/datas"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/permission"
	"sirherobrine23.com.br/go-bds/bds/modules/datas/user"
	"sirherobrine23.com.br/go-bds/bds/modules/web/templates"
)

type contextType int

const (
	CookieName             = "bdsCookie"
	_          contextType = iota

	ContextConfig
	ContextWebTemplates
	ContextToken
	ContextTokenPerm
	ContextUser
)

func getConfig(r *http.Request) *datas.DatabaseSchemas {
	switch v := r.Context().Value(ContextConfig).(type) {
	case *datas.DatabaseSchemas:
		return v
	case datas.DatabaseSchemas:
		return &v
	default:
		return nil
	}
}

func getUser(r *http.Request) *user.User {
	switch v := r.Context().Value(ContextUser).(type) {
	case *user.User:
		return v
	default:
		return nil
	}
}

func getTokenPerm(r *http.Request) permission.Permission {
	switch v := r.Context().Value(ContextTokenPerm).(type) {
	case permission.Permission:
		return v
	default:
		return permission.Unknown
	}
}

func getTemplates(r *http.Request) *templates.TemplateRender {
	switch v := r.Context().Value(ContextWebTemplates).(type) {
	case *templates.TemplateRender:
		return v
	case templates.TemplateRender:
		return &v
	default:
		return nil
	}
}

func jsonWrite(status int, w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	js := json.NewEncoder(w)
	js.SetIndent("", "  ")
	js.Encode(data)
}
