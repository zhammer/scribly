package scribly

import (
	"embed"
	"html/template"
)

//go:embed templates/email/*
var templateFS embed.FS
var EmailTemplates = template.Must(template.ParseFS(templateFS, "templates/email/*tmpl"))

//go:embed static/style.css
var CSS string

//go:embed static/
var Static embed.FS
