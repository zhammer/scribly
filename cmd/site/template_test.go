package site_test

import (
	"net/http"
	"net/http/httptest"
	"scribly/cmd/site"
	"scribly/embed"
	"scribly/internal"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoryTemplate(t *testing.T) {
	userZach := internal.User{
		ID:       1,
		Username: "zach",
	}
	userGabe := internal.User{
		ID:       2,
		Username: "gabe",
	}
	userStory := internal.UserStory{
		UserID: userZach.ID,
		Story: internal.Story{
			ID:            1,
			Title:         "Test Story",
			CurrentWriter: &userZach,
			State:         internal.StoryStateInProgress,
			Cowriters: []internal.StoryCowriter{
				{
					User: userZach,
				},
				{
					User: userGabe,
				},
			},
			Turns: []internal.Turn{
				{
					TakenByID: userZach.ID,
					Action:    internal.TurnActionWrite,
					Text:      "There was a car\nwow",
				},
				{
					TakenByID: userGabe.ID,
					Action:    internal.TurnActionWrite,
					Text:      "And there was a <horse>",
				},
			},
		},
	}

	// make a buffer writer for testing output
	var buff strings.Builder

	err := embed.WebTemplates.ExecuteTemplate(&buff, "story.tmpl", site.ViewData{
		Data: &userStory,
	})
	assert.NoError(t, err)
	assert.Contains(t, buff.String(), "There was a car\nwow")
	// test that <horse> was escaped, by showing what it would be escaped to
	assert.Contains(t, buff.String(), "&lt;horse&gt;")
	// test that there's no leading/trailing whitespace around turn text
	assert.Contains(t, buff.String(), `<p id="turn-1" class="text">There was a car
wow</p>`)
	assert.Contains(t, buff.String(), `<p id="turn-2" class="text">And there was a &lt;horse&gt;</p>`)

}

func TestThemeClass(t *testing.T) {
	t.Run("returns empty string for default theme", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		vd := site.ViewData{Request: req}

		assert.Equal(t, "", vd.ThemeClass())
	})

	t.Run("returns theme-candlelit for candlelit theme", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  "scribly-style-preference",
			Value: "candlelit",
		})
		vd := site.ViewData{Request: req}

		assert.Equal(t, "theme-candlelit", vd.ThemeClass())
	})

	t.Run("returns empty string for invalid theme", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  "scribly-style-preference",
			Value: "nonexistent-theme",
		})
		vd := site.ViewData{Request: req}

		assert.Equal(t, "", vd.ThemeClass())
	})

	t.Run("returns empty string when request is nil", func(t *testing.T) {
		vd := site.ViewData{Request: nil}

		assert.Equal(t, "", vd.ThemeClass())
	})
}

func TestNextThemeIcon(t *testing.T) {
	t.Run("returns candlelit icon when on default theme", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		vd := site.ViewData{Request: req}

		assert.Equal(t, "🕯️", vd.NextThemeIcon())
	})

	t.Run("returns default icon when on candlelit theme", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  "scribly-style-preference",
			Value: "candlelit",
		})
		vd := site.ViewData{Request: req}

		assert.Equal(t, "📃", vd.NextThemeIcon())
	})

	t.Run("returns candlelit icon when request is nil", func(t *testing.T) {
		vd := site.ViewData{Request: nil}

		assert.Equal(t, "🕯️", vd.NextThemeIcon())
	})
}

func TestNextThemeName(t *testing.T) {
	t.Run("returns candlelit when on default theme", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		vd := site.ViewData{Request: req}

		assert.Equal(t, "candlelit", vd.NextThemeName())
	})

	t.Run("returns default when on candlelit theme", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  "scribly-style-preference",
			Value: "candlelit",
		})
		vd := site.ViewData{Request: req}

		assert.Equal(t, "default", vd.NextThemeName())
	})

	t.Run("returns candlelit when request is nil", func(t *testing.T) {
		vd := site.ViewData{Request: nil}

		assert.Equal(t, "candlelit", vd.NextThemeName())
	})
}

func TestThemeToggleHandler(t *testing.T) {
	cfg := site.Config{}
	cfg.DatabaseURL = "postgres://test"
	cfg.ResendAPIKey = "test"

	router, err := site.MakeRouter(cfg)
	require.NoError(t, err)

	t.Run("sets cookie to candlelit when no current theme", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/theme/toggle", nil)
		req.Header.Set("referer", "/me")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusSeeOther, w.Code)
		assert.Equal(t, "/me?", w.Header().Get("Location"))

		cookies := w.Result().Cookies()
		require.Len(t, cookies, 1)
		assert.Equal(t, "scribly-style-preference", cookies[0].Name)
		assert.Equal(t, "candlelit", cookies[0].Value)
		assert.Equal(t, "/", cookies[0].Path)
		assert.Equal(t, 365*24*60*60, cookies[0].MaxAge)
	})

	t.Run("cycles from candlelit back to default", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/theme/toggle", nil)
		req.AddCookie(&http.Cookie{
			Name:  "scribly-style-preference",
			Value: "candlelit",
		})
		req.Header.Set("referer", "/stories/123")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusSeeOther, w.Code)
		assert.Equal(t, "/stories/123?", w.Header().Get("Location"))

		cookies := w.Result().Cookies()
		require.Len(t, cookies, 1)
		assert.Equal(t, "scribly-style-preference", cookies[0].Name)
		assert.Equal(t, "default", cookies[0].Value)
	})

	t.Run("redirects to /me when no referer", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/theme/toggle", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusSeeOther, w.Code)
		assert.Equal(t, "/me?", w.Header().Get("Location"))
	})
}

