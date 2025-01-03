package ui

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
)

//go:embed static/*
var StaticFS embed.FS

//go:embed icon/*.svg
var icons embed.FS

//go:embed *.gotmpl
var templateFiles embed.FS

func Templates() *template.Template {
	templates := template.New("")
	templates = templates.Funcs(template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s) //nolint:gosec
		},
		"icon": func(name string, size int) template.HTML {
			b, err := icons.ReadFile("icon/" + name + ".svg")
			if err != nil {
				return "icon not found"
			}

			if size > 0 {
				b = bytes.ReplaceAll(b, []byte("width=\"24\""), []byte(fmt.Sprintf("width=\"%d\"", size)))
				b = bytes.ReplaceAll(b, []byte("height=\"24\""), []byte(fmt.Sprintf("height=\"%d\"", size)))
			}

			return template.HTML(b) //nolint:gosec
		},
	})

	return template.Must(templates.ParseFS(templateFiles, "*.gotmpl"))
}
