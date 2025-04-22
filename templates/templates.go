package webTemplates

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	db "sirherobrine23.com.br/go-bds/bds/modules/database"
)

var (
	//go:embed */*
	TemplatesFiles embed.FS

	WebTemplate *template.Template // Root templates
)

func init() {
	WebTemplate = template.New("root")
	if err := fs.WalkDir(TemplatesFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		println("loadding:", path)
		buf, err := TemplatesFiles.ReadFile(path)
		if err != nil {
			return err
		}
		tmpl := WebTemplate.New(filepath.ToSlash(path))
		if _, err = tmpl.Parse(string(buf)); err != nil {
			return err
		}
		return nil
	}); err != nil {
		panic(err)
	}
}

func LoadTemplate(load string) *template.Template {
	if strings.HasPrefix(load, "base") {
		return nil
	}
	return WebTemplate.Lookup(load)
}

type ErrData struct {
	Title                 string
	Signed, PageIsInstall bool
	ErrorMSg              string
}

func StatusTemplate(w io.Writer, IsSigned bool, err error) {
	templateStatus := WebTemplate.Lookup("status/500.tmpl")
	fmt.Println(templateStatus.Execute(w, ErrData{
		Title:         "Internal Error",
		Signed:        IsSigned,
		PageIsInstall: false,
		ErrorMSg:      err.Error(),
	}))
}

func StatusTemplate404(w io.Writer, IsSigned bool, message string) {
	templateStatus := WebTemplate.Lookup("status/not_found.tmpl")
	fmt.Println(templateStatus.Execute(w, map[string]any{
		"Title":   "Page not found",
		"Signed":  IsSigned,
		"Message": message,
	}))
}

func StatusTemplate500(w http.ResponseWriter, r *http.Request, err error) {
	GetUserCtx := func(r *http.Request) *db.User {
		user := r.Context().Value("bdsuser")
		switch user := user.(type) {
		case db.User:
			return &user
		case *db.User:
			return user
		default:
			return nil
		}
	}

	w.WriteHeader(500)
	templateStatus := WebTemplate.Lookup("status/500.tmpl")
	fmt.Println(templateStatus.Execute(w, map[string]any{
		"Title":    "500 Error",
		"Signed":   GetUserCtx(r) != nil,
		"User":     GetUserCtx(r),
		"ErrorMSg": err.Error(),
	}))
}
