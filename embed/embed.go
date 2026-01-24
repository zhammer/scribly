package embed

import (
	"embed"
	"html/template"
)

//go:embed templates/*
var templateFS embed.FS
var EmailTemplates = template.Must(template.ParseFS(templateFS, "templates/email/*tmpl"))
var WebTemplates = template.Must(template.ParseFS(templateFS, "templates/web/*tmpl"))
