package cmd

import (
	"context"
	"fmt"

	"github.com/go-pg/pg/v10"
)

type DBLogger struct {
}

func (d DBLogger) BeforeQuery(c context.Context, q *pg.QueryEvent) (context.Context, error) {
	return c, nil
}

func (d DBLogger) AfterQuery(c context.Context, q *pg.QueryEvent) error {
	query, _ := q.FormattedQuery()
	fmt.Println(string(query))
	return nil
}
