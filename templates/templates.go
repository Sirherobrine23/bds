package webTemplates

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
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
