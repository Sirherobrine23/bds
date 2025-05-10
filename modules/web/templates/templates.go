package templates

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"maps"
	"net/http"
	"path/filepath"
	"runtime"
	"slices"

	"sirherobrine23.com.br/go-bds/bds/modules/datas/user"
)

// Templates files
//
//go:embed */*
var TemplatesFiles embed.FS

// Base to base/{head, footer and head_navbar}
type RenderData struct {
	Title         string
	Lang          string
	PageIsInstall bool
	User          *user.User

	External map[string]any
}

func (b RenderData) toRender(t *TemplateRender, stack []byte) map[string]any {
	newData := map[string]any{}
	maps.Insert(newData, maps.All(b.External))
	newData["Title"] = b.Title
	newData["Lang"] = b.Lang
	newData["PageIsInstall"] = b.PageIsInstall
	newData["User"] = b.User
	newData["ShowTrace"] = t.ShowTrace
	newData["Trace"] = string(stack)
	return newData
}

// Struct with templates
type TemplateRender struct {
	Root      *template.Template
	ShowTrace bool
}

// Load all templates and return [*TemplateRender] to process html templates
func Templates() (WebTemplate *TemplateRender, err error) {
	WebTemplate = &TemplateRender{Root: template.New("root")}
	err = fs.WalkDir(TemplatesFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			var templateBody []byte
			if templateBody, err = TemplatesFiles.ReadFile(path); err == nil {
				tmpl := WebTemplate.Root.New(filepath.ToSlash(path))
				_, err = tmpl.Parse(string(templateBody))
			}
		}
		return err
	})
	return
}

// List all templates loaded into [*template.Template]
func (t *TemplateRender) Templates() (names []string) {
	for _, tmpl := range t.Root.Templates() {
		if tmpl.Name() == "root" {
			continue
		}
		names = append(names, tmpl.Name())
	}
	slices.Sort(names)
	return
}

// Render template
func (t *TemplateRender) Render(name string, w io.Writer, data *RenderData) error {
	if !slices.Contains(t.Templates(), name) {
		return fmt.Errorf("template not exists: %q", name)
	} else if data == nil {
		data = &RenderData{External: map[string]any{}}
	}
	return t.Root.Lookup(name).Execute(w, data.toRender(t, nil))
}

func (t *TemplateRender) Render400(w io.Writer, data *RenderData) {
	if data == nil {
		data = &RenderData{
			Title:         "Bad request",
			Lang:          "en-us",
			PageIsInstall: false,
			User:          nil,
			External: map[string]any{
				"Error": "Bad request",
			},
		}
	}

	if httpWrite, ok := w.(http.ResponseWriter); ok {
		httpWrite.WriteHeader(http.StatusBadRequest)
	}

	t.Root.Lookup("status/400.tmpl").Execute(w, data.toRender(t, nil))
}

// Render 404 page
func (t *TemplateRender) Render404(w io.Writer, data *RenderData) {
	if data == nil {
		data = &RenderData{
			Title:         "page not found",
			Lang:          "en-us",
			PageIsInstall: false,
			User:          nil,
			External:      map[string]any{},
		}
	}

	if httpWrite, ok := w.(http.ResponseWriter); ok {
		httpWrite.WriteHeader(http.StatusNotFound)
	}

	t.Root.Lookup("status/404.tmpl").Execute(w, data.toRender(t, nil))
}

// Render Backend error with caller stack
func (t *TemplateRender) Render5xx(w io.Writer, data *RenderData) {
	stackBuff := make([]byte, 1024)
	for {
		if n := runtime.Stack(stackBuff, false); n < len(stackBuff) {
			stackBuff = stackBuff[:n]
			break
		}
		stackBuff = make([]byte, 2*len(stackBuff))
	}

	if httpWrite, ok := w.(http.ResponseWriter); ok {
		httpWrite.WriteHeader(http.StatusInternalServerError)
	}

	t.Root.Lookup("status/500.tmpl").Execute(w, data.toRender(t, stackBuff))
}
