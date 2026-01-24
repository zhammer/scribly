package embed

import (
	"embed"
	"html/template"
	"io/fs"
)

//go:embed templates/*
var templateFS embed.FS
var EmailTemplates = template.Must(template.ParseFS(templateFS, "templates/email/*tmpl"))
var WebTemplates = template.Must(template.ParseFS(templateFS, "templates/web/*tmpl"))

// NOTE: The only reason for this nested public/static structure is for vercel, which does not
// provide great mechanics for only serving specific static assets when using functions, from what I can tell.
// (See: https://community.vercel.com/t/unwanted-serving-of-static-files/905, as well as related issues.)
// We nest static in public/ (which is arbitrary, it could as well be nested/static) so that we can tell
// vercel to _only_ include embed/public in the output directory.

//go:embed public/static/style.css
var CSS string

//go:embed public/static/*
var nestFS embed.FS
var StaticFS = must(fs.Sub(nestFS, "public"))

func must(f fs.FS, err error) fs.FS {
	if err != nil {
		panic(err)
	}
	return f
}
