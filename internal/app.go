package internal

import "context"

type Scribly struct {
}

func (s *Scribly) LogIn(ctx context.Context, username string, password string) (User, error) {
	return User{}, ErrNotImplemented
}

func (s *Scribly) SignUp(ctx context.Context, username string, password string, email string) (User, error) {
	return User{}, ErrNotImplemented
}
