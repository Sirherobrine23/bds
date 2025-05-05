package webTemplates

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"maps"
	"path/filepath"
	"runtime"
	"slices"
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
	User          any

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
	return
}

// Render template
func (t *TemplateRender) Render(name string, w io.Writer, data *RenderData) error {
	if !slices.Contains(t.Templates(), name) {
		return fmt.Errorf("template not exists: %q", name)
	}
	return t.Root.Lookup(name).Execute(w, data.toRender(t, nil))
}

// Render 404 page
func (t *TemplateRender) Render404(w io.Writer, data *RenderData) {
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
	t.Root.Lookup("status/500.tmpl").Execute(w, data.toRender(t, stackBuff))
}
