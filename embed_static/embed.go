package embed_static

import (
	"embed"
)

//go:embed static/style.css
var CSS string

//go:embed static/*
var StaticFS embed.FS
