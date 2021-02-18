package internal

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path"

	"github.com/vanng822/go-premailer/premailer"
)

var css string

type viewData struct {
	Data       interface{}
	websiteURL string
}

func (v viewData) StoryLink(story Story) template.HTML {
	return template.HTML(fmt.Sprintf(`<a href="%s/stories/%d">%s</a>`, v.websiteURL, story.ID, story.Title))
}

func (v viewData) CSS() template.HTML {
	return template.HTML("<style>" + css + "</style>")
}

func BuildNudgeEmail(nudger User, nudgee User, story Story) (*Email, error) {
	subject := fmt.Sprintf("%s nudged you to take your turn on %s", nudger.Username, story.Title)
	data := map[string]interface{}{
		"Story":  story,
		"Nudger": nudger,
	}
	body, err := renderTemplateWithCSS("nudge.tmpl", data)
	if err != nil {
		return nil, err
	}

	return &Email{
		Subject: subject,
		Body:    body,
		To:      nudgee,
	}, nil
}

func renderTemplateWithCSS(templateName string, data interface{}) (string, error) {
	template, err := template.ParseFiles("goemailtemplates/_layout.tmpl", path.Join("goemailtemplates", templateName))
	if err != nil {
		return "", err
	}
	var buffer bytes.Buffer
	viewData := viewData{Data: data, websiteURL: os.Getenv("WEBSITE_URL")}
	if err := template.ExecuteTemplate(&buffer, templateName, viewData); err != nil {
		return "", err
	}
	prem, err := premailer.NewPremailerFromBytes(buffer.Bytes(), premailer.NewOptions())
	if err != nil {
		return "", err
	}

	return prem.Transform()
}

func init() {
	styleCSS, err := os.ReadFile("static/style.css")
	if err != nil {
		panic(err)
	}

	css = string(styleCSS)
}
