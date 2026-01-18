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
	t.Run("returns next theme icon in rotation", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		vd := site.ViewData{Request: req}

		// Should rotate to the next theme after default
		nextIcon := vd.NextThemeIcon()
		assert.NotEmpty(t, nextIcon)
	})

	t.Run("returns different icon for each theme", func(t *testing.T) {
		// Test that cycling through all themes eventually returns to start
		icons := []string{}

		req := httptest.NewRequest("GET", "/", nil)

		// Get icons for a full cycle (should wrap around)
		for i := 0; i < 4; i++ {
			vd := site.ViewData{Request: req}
			icon := vd.NextThemeIcon()
			icons = append(icons, icon)

			// Update cookie to next theme
			themeName := vd.NextThemeName()
			req = httptest.NewRequest("GET", "/", nil)
			req.AddCookie(&http.Cookie{
				Name:  "scribly-style-preference",
				Value: themeName,
			})
		}

		// After cycling through, we should see at least 2 different icons
		uniqueIcons := make(map[string]bool)
		for _, icon := range icons {
			uniqueIcons[icon] = true
		}
		assert.GreaterOrEqual(t, len(uniqueIcons), 2)
	})

	t.Run("returns consistent icon when request is nil", func(t *testing.T) {
		vd := site.ViewData{Request: nil}

		icon1 := vd.NextThemeIcon()
		icon2 := vd.NextThemeIcon()
		assert.Equal(t, icon1, icon2)
	})
}

func TestNextThemeName(t *testing.T) {
	t.Run("returns next theme name in rotation", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		vd := site.ViewData{Request: req}

		// Should return a valid theme name
		nextName := vd.NextThemeName()
		assert.NotEmpty(t, nextName)
	})

	t.Run("cycles through all themes and wraps around", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		vd := site.ViewData{Request: req}
		startingTheme := vd.NextThemeName()

		// Cycle through themes
		currentTheme := startingTheme
		seen := []string{currentTheme}

		for i := 0; i < 5; i++ {
			req = httptest.NewRequest("GET", "/", nil)
			req.AddCookie(&http.Cookie{
				Name:  "scribly-style-preference",
				Value: currentTheme,
			})
			vd = site.ViewData{Request: req}
			currentTheme = vd.NextThemeName()
			seen = append(seen, currentTheme)
		}

		// Should eventually cycle back to starting theme
		foundStart := false
		for i := 1; i < len(seen); i++ {
			if seen[i] == startingTheme {
				foundStart = true
				break
			}
		}
		assert.True(t, foundStart, "Theme rotation should cycle back to start")
	})

	t.Run("returns consistent name when request is nil", func(t *testing.T) {
		vd := site.ViewData{Request: nil}

		name1 := vd.NextThemeName()
		name2 := vd.NextThemeName()
		assert.Equal(t, name1, name2)
	})
}

func TestThemeToggleHandler(t *testing.T) {
	cfg := site.Config{}
	cfg.DatabaseURL = "postgres://test"
	cfg.ResendAPIKey = "test"

	router, err := site.MakeRouter(cfg)
	require.NoError(t, err)

	t.Run("sets cookie to next theme when no current theme", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/theme/toggle", nil)
		req.Header.Set("referer", "/me")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusSeeOther, w.Code)
		assert.Equal(t, "/me?", w.Header().Get("Location"))

		cookies := w.Result().Cookies()
		require.Len(t, cookies, 1)
		assert.Equal(t, "scribly-style-preference", cookies[0].Name)
		assert.NotEmpty(t, cookies[0].Value, "Should set a theme")
		assert.NotEqual(t, "default", cookies[0].Value, "Should not stay on default")
		assert.Equal(t, "/", cookies[0].Path)
		assert.Equal(t, 365*24*60*60, cookies[0].MaxAge)
	})

	t.Run("cycles through themes", func(t *testing.T) {
		// Get first theme
		req := httptest.NewRequest("POST", "/theme/toggle", nil)
		req.Header.Set("referer", "/stories/123")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		cookies := w.Result().Cookies()
		require.Len(t, cookies, 1)
		firstTheme := cookies[0].Value

		// Toggle again with first theme
		req = httptest.NewRequest("POST", "/theme/toggle", nil)
		req.AddCookie(&http.Cookie{
			Name:  "scribly-style-preference",
			Value: firstTheme,
		})
		req.Header.Set("referer", "/stories/123")
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusSeeOther, w.Code)
		assert.Equal(t, "/stories/123?", w.Header().Get("Location"))

		cookies = w.Result().Cookies()
		require.Len(t, cookies, 1)
		assert.Equal(t, "scribly-style-preference", cookies[0].Name)
		secondTheme := cookies[0].Value
		assert.NotEqual(t, firstTheme, secondTheme, "Should cycle to a different theme")
	})

	t.Run("redirects to /me when no referer", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/theme/toggle", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusSeeOther, w.Code)
		assert.Equal(t, "/me?", w.Header().Get("Location"))
	})
}

