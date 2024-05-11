package site_test

import (
	"scribly/cmd/site"
	"scribly/embed"
	"scribly/internal"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert.Contains(t, buff.String(), "There was a car<br>wow")
	// test that <horse> was escaped, by showing what it would be escaped to
	assert.Contains(t, buff.String(), "&lt;horse&gt;")

}
