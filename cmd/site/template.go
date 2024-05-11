package site

import (
	"fmt"
	"html"
	"html/template"
	"net/http"
	"strings"

	"github.com/mssola/user_agent"
)

type ViewData struct {
	Data      interface{}
	Request   *http.Request
	title     string
	pageClass string
}

// v *ViewData doesn't seem to work. i guess go templates don't have the same
// pointer-whichever-way-you-want-it magic as actual go code.
func (v ViewData) Query(key string) string {
	return v.Request.URL.Query().Get(key)
}

func (v ViewData) Mobile() bool {
	ua := user_agent.New(v.Request.Header.Get("user-agent"))
	return ua.Mobile()
}

func (v ViewData) Title() string {
	if v.title == "" {
		return "Scribly"
	}
	return v.title
}

func (v ViewData) PageClass() string {
	if v.pageClass == "" {
		return "page"
	}
	return v.pageClass
}

func (v ViewData) Ternary(cond bool, a interface{}, b interface{}) interface{} {
	if cond {
		return a
	} else {
		return b
	}
}

func (v ViewData) Add(a int, b int) int {
	return a + b
}

func (v ViewData) NewLineify(str string) template.HTML {
	return template.HTML(strings.ReplaceAll(html.EscapeString(str), "\n", "<br>"))
}

func (v ViewData) Replace(original string, pattern string, replacement string) string {
	return strings.ReplaceAll(original, pattern, replacement)
}

func (v ViewData) Count(from uint, to uint) ([]uint, error) {
	if to < from {
		return nil, fmt.Errorf("to >= from, got to=%d from=%d", to, from)
	}

	var items []uint
	for i := from; i < to; i++ {
		items = append(items, i)
	}
	return items, nil
}

// lets us pass some other data into other templates while preserving the top-level
// request data (and other context we may add later.)
func (v ViewData) Propogate(data interface{}) ViewData {
	return ViewData{
		Data:    data,
		Request: v.Request,
	}
}
