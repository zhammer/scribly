package main

import (
	"context"
	"log"
	"scribly/cmd"
	"scribly/internal"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	cmd.Config
	OpenAIBaseURL string `envconfig:"openai_base_url" default:"https://api.openai.com"`
	OpenAIAPIKey  string `envconfig:"openai_api_key" default:"test_openai_api_key"`
}

func (c *Config) MakeScribbot() (*internal.Scribbot, error) {
	scribly, err := c.MakeScribly()
	if err != nil {
		return nil, err
	}

	openai := internal.NewHTTPOpenAIGateway(c.OpenAIAPIKey, internal.WithBaseURL(c.OpenAIAPIKey))
	return internal.NewScribbot(scribly, openai), nil
}

func main() {
	cfg := Config{}
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(err)
	}

	scribbot, err := cfg.MakeScribbot()
	if err != nil {
		log.Fatal(err)
	}

	if err := scribbot.TakeScribbotTurns(context.Background()); err != nil {
		log.Fatal(err)
	}
}
