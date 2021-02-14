package internal

import (
	"context"

	"github.com/go-pg/pg/v10"
)

type Scribly struct {
	db      *pg.DB
	emailer EmailGateway
}

func (s *Scribly) LogIn(ctx context.Context, username string, password string) (*User, error) {
	return &User{}, ErrNotImplemented
}

func (s *Scribly) SignUp(ctx context.Context, input SignUpInput) (*User, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	user := User{}
	_, err := s.db.QueryOne(&user, `
		INSERT INTO users (username, email, password) VALUES (?, ?, crypt(?, gen_salt('bf', 8)))
		RETURNING id, username, email, email_verification_status
	`, input.Username, input.Email, input.Password)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func NewScribly(db *pg.DB, emailer EmailGateway) (*Scribly, error) {
	return &Scribly{
		db:      db,
		emailer: emailer,
	}, nil
}
