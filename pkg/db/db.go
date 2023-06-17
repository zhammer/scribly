package db

import "github.com/uptrace/bun"

type relationshipFunc func(*bun.SelectQuery) (*bun.SelectQuery, error)

func WithOrderBy(column string) relationshipFunc {
	return func(q *bun.SelectQuery) (*bun.SelectQuery, error) {
		return q.Order(column), nil
	}
}
