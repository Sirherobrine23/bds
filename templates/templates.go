package webTemplates

import (
	"embed"
	"io/fs"
	"path/filepath"
	"strings"
	"text/template"
)

var (
	//go:embed **/*.tmpl
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
