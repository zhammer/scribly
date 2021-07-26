package internal

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildPrompt(t *testing.T) {
	rand.Seed(1337)
	const (
		bangbang1 = `I said "Scuse me, have you got the time"`
		bangbang2 = `He shrugged his shoulders said, "Get a spine"`
		bangbang3 = `Coconut hit me on the head, jonesin' for a fig`
	)
	testCases := []struct {
		name     string
		story    Story
		expected string
	}{
		{
			name: "basic story",
			story: Story{
				Turns: []Turn{
					{Text: bangbang1, Action: TurnActionWrite},
					{Text: bangbang2, Action: TurnActionWrite},
					{Text: ``, Action: TurnActionPass},
					{Text: bangbang3, Action: TurnActionWrite},
				},
			},
			expected: bangbang1 + turnSeparator + bangbang2 + turnSeparator + bangbang3 + turnSeparator,
		},
		{
			name: "truncates prompt",
			story: Story{
				Turns: []Turn{
					{Text: strings.Repeat("a", 1000), Action: TurnActionWrite},
					{Text: strings.Repeat("b", 999), Action: TurnActionWrite},
				},
			},
			expected: strings.Repeat("a", 23) + turnSeparator + strings.Repeat("b", 999) + turnSeparator,
		},
		{
			name:     "generates random prompt if story empty",
			story:    Story{},
			expected: pachinko[len(pachinko)-1023:] + turnSeparator,
		},
		{
			name: "sanitizes existing turn separator characters",
			story: Story{
				Turns: []Turn{
					{Text: "Hi ~" + turnSeparator + "~ <(-_-)>", Action: TurnActionWrite},
				},
			},
			expected: "Hi ~" + turnSeparatorSanitized + "~ <(-_-)>" + turnSeparator,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			prompt := buildPrompt(tt.story)
			assert.Equal(t, tt.expected, prompt)
		})
	}
}
