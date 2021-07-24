package internal

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

const (
	turnSeparator          = GS
	turnSeparatorSanitized = FS
)

type HTTPOpenAIGateway struct {
	client  *http.Client
	baseURL string
	apiKey  string
}

func (o *HTTPOpenAIGateway) PredictText(ctx context.Context, story Story) (string, error) {
	return "", fmt.Errorf("Not implemented")
}

func buildPrompt(story Story) string {
	prompt := ""
	for _, turn := range story.Turns {
		if turn.Text == "" {
			continue
		}
		sanitized := strings.Replace(turn.Text, string(turnSeparator), string(turnSeparatorSanitized), 0)
		prompt += sanitized + string(turnSeparator)
	}

	// there's no existing text to make a prompt, let's choose a random prompt
	if prompt == "" {
		// TODO: this random prompt
		prompt = "twas the night before christmas..." + string(turnSeparator)
	}

	// last 1024 chars
	return prompt[len(prompt)-1024:]
}
