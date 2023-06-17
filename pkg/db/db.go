package db

import (
	"github.com/uptrace/bun"
)

type relationshipFunc func(*bun.SelectQuery) *bun.SelectQuery

func WithOrderBy(column string) relationshipFunc {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Order(column)
	}
}
