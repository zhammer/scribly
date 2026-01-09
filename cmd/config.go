package cmd

import (
	"scribly/internal"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/extra/bundebug"
)

type Config struct {
	DatabaseURL   string `envconfig:"database_url" default:"postgres://scribly:pass@localhost/scribly?sslmode=disable"`
	ResendBaseURL string `envconfig:"resend_base_url" default:"https://api.resend.com"`
	ResendAPIKey  string `envconfig:"resend_api_key" default:"test_resend_api_key"`
	Debug         bool
}

func (c *Config) MakeScribly() (*internal.Scribly, error) {
	config, err := pgx.ParseConfig(c.DatabaseURL)
	if err != nil {
		panic(err)
	}
	config.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	sqldb := stdlib.OpenDB(*config)
	db := bun.NewDB(sqldb, pgdialect.New())

	if c.Debug {
		db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}

	resend := internal.NewResendClient(c.ResendBaseURL, c.ResendAPIKey)
	messageGateway := internal.GoroutineMessageGateway{}

	scribly, err := internal.NewScribly(db, resend, &messageGateway)
	if err != nil {
		return nil, err
	}
	messageGateway.Scribly(scribly)

	return scribly, nil
}
