package cmd

import (
	"context"
	"scribly/internal"

	"github.com/go-pg/pg/v10"
)

type Config struct {
	DatabaseURL     string `envconfig:"database_url" default:"postgres://scribly:pass@localhost/scribly?sslmode=disable"`
	SendgridBaseURL string `envconfig:"sendgrid_base_url" default:"https://api.sendgrid.com"`
	SendgridAPIKey  string `envconfig:"sendgrid_api_key" default:"test_sendgrid_api_key"`
	Debug           bool
}

func (c *Config) MakeScribly() (*internal.Scribly, error) {
	opt, err := pg.ParseURL(c.DatabaseURL)
	if err != nil {
		return nil, err
	}

	db := pg.Connect(opt)
	if err := db.Ping(context.Background()); err != nil {
		return nil, err
	}

	if c.Debug {
		db.AddQueryHook(DBLogger{})
	}

	sendgrid := internal.NewSendgridClient(c.SendgridBaseURL, c.SendgridAPIKey)
	messageGateway := internal.GoroutineMessageGateway{}

	scribly, err := internal.NewScribly(db, sendgrid, &messageGateway)
	if err != nil {
		return nil, err
	}
	messageGateway.Scribly(scribly)

	return scribly, nil
}
