package internal

import (
	"context"

	"github.com/go-pg/pg/v10"
)

type Scribly struct {
	db      *pg.DB
	emailer EmailGateway
}

func (s *Scribly) LogIn(ctx context.Context, username string, password string) (User, error) {
	return User{}, ErrNotImplemented
}

func (s *Scribly) SignUp(ctx context.Context, username string, password string, email string) (User, error) {
	return User{}, ErrNotImplemented
}

func NewScribly(db *pg.DB, emailer EmailGateway) (*Scribly, error) {
	return &Scribly{
		db:      db,
		emailer: emailer,
	}, nil
}
