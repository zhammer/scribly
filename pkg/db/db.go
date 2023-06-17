package db

import "github.com/go-pg/pg/v10/orm"

type relationshipFunc func(*orm.Query) (*orm.Query, error)

func WithOrderBy(column string) relationshipFunc {
	return func(q *orm.Query) (*orm.Query, error) {
		return q.Order(column), nil
	}
}
