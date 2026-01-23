package site

import (
	"fmt"
	"html"
	"html/template"
	"net/http"
	"strings"

	"github.com/mssola/useragent"
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
	ua := useragent.New(v.Request.Header.Get("user-agent"))
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

// Theme represents a visual style for the app
type Theme struct {
	Name     string // "default", "candlelit", etc.
	CSSClass string // "", "theme-candlelit", etc.
	Icon     string // emoji shown for this theme
}

// availableThemes defines the rotation order
// Note: Icon represents what you'll see on the button to switch TO this theme (from the previous theme)
var availableThemes = []Theme{
	{Name: "default", CSSClass: "", Icon: "📃"},
	{Name: "candlelit", CSSClass: "theme-candlelit", Icon: "🕯️"},
	{Name: "stars", CSSClass: "theme-stars", Icon: "🌌"},
	// Future themes can be added here
}

// getCurrentTheme returns the current theme based on cookie value
func getCurrentTheme(r *http.Request) Theme {
	if r == nil {
		return availableThemes[0] // default
	}

	cookie, err := r.Cookie("scribly-style-preference")
	if err != nil {
		return availableThemes[0] // default
	}

	for _, theme := range availableThemes {
		if theme.Name == cookie.Value {
			return theme
		}
	}
	return availableThemes[0] // fallback to default
}

// getNextTheme returns the next theme in rotation
func getNextTheme(currentTheme Theme) Theme {
	for i, theme := range availableThemes {
		if theme.Name == currentTheme.Name {
			// Return next theme (wrap around to start if at end)
			return availableThemes[(i+1)%len(availableThemes)]
		}
	}
	return availableThemes[0] // fallback
}

// ThemeClass returns the CSS class for the current theme
func (v ViewData) ThemeClass() string {
	return getCurrentTheme(v.Request).CSSClass
}

// NextThemeIcon returns the icon for the next theme in rotation
func (v ViewData) NextThemeIcon() string {
	currentTheme := getCurrentTheme(v.Request)
	nextTheme := getNextTheme(currentTheme)
	return nextTheme.Icon
}

// NextThemeIcon returns the icon for the next theme in rotation
func (v ViewData) NextThemeName() string {
	currentTheme := getCurrentTheme(v.Request)
	nextTheme := getNextTheme(currentTheme)
	return nextTheme.Name
}
