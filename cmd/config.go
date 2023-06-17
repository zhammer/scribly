package cmd

import (
	"scribly/internal"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/extra/bundebug"
)

type Config struct {
	DatabaseURL     string `envconfig:"database_url" default:"postgres://scribly:pass@localhost/scribly?sslmode=disable"`
	SendgridBaseURL string `envconfig:"sendgrid_base_url" default:"https://api.sendgrid.com"`
	SendgridAPIKey  string `envconfig:"sendgrid_api_key" default:"test_sendgrid_api_key"`
	Debug           bool
}

func (c *Config) MakeScribly() (*internal.Scribly, error) {
	config, err := pgx.ParseConfig(c.DatabaseURL)
	if err != nil {
		panic(err)
	}
	config.PreferSimpleProtocol = true

	sqldb := stdlib.OpenDB(*config)
	db := bun.NewDB(sqldb, pgdialect.New())

	if c.Debug {
		db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
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
