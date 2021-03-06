package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStory(t *testing.T) {
	t.Run("ValidateUserCanAddCowriters", func(t *testing.T) {
		testCases := []struct {
			name      string
			user      User
			story     Story
			cowriters []string
			expected  string
		}{
			{
				name: "Can't add myself",
				user: User{ID: 1, Username: "zach"},
				story: Story{
					ID:          1,
					State:       StoryStateDraft,
					CreatedByID: 1,
				},
				cowriters: []string{"zach"},
				expected:  "cannot add yourself",
			},
			{
				name: "Can't add myself (case insensitive)",
				user: User{ID: 1, Username: "zach"},
				story: Story{
					ID:          1,
					State:       StoryStateDraft,
					CreatedByID: 1,
				},
				cowriters: []string{"ZACH"},
				expected:  "cannot add yourself",
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				err := testCase.story.ValidateUserCanAddCowriters(testCase.user, testCase.cowriters)
				assert.Contains(t, err.Error(), testCase.expected)
			})
		}
	})
}
