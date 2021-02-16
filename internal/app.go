package internal

import (
	"context"

	"github.com/go-pg/pg/v10"
)

type Scribly struct {
	db      *pg.DB
	emailer EmailGateway
}

func (s *Scribly) LogIn(ctx context.Context, input LoginInput) (*User, error) {
	user := User{}
	_, err := s.db.QueryOne(&user, `
		SELECT id, username FROM users WHERE username = ? AND password = crypt(?, password);
	`, input.Username, input.Password)
	if err != nil {
		return nil, err
	}
	return &user, nil
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

func (s *Scribly) Me(ctx context.Context, user *User) (*Me, error) {
	// todo: figure out how to do this in one query, if that's bettah
	// refresh our user model
	if err := s.db.Model(user).WherePK().Select(); err != nil {
		return nil, err
	}
	var userStories []UserStory
	if err := s.db.Model(&userStories).
		Where("user_id = ?", user.ID).
		Relation("Story").
		Select(); err != nil {
		return nil, err
	}

	return &Me{
		User:    user,
		Stories: userStories,
	}, nil

}

func (s *Scribly) UserStory(ctx context.Context, userID int, storyID int) (*UserStory, error) {
	story := UserStory{}
	if err := s.db.Model(&story).
		Where("story_id = ? AND user_id = ?", storyID, userID).
		Relation("Story").
		Relation("Story.Cowriters").
		Relation("Story.Cowriters.User").
		Relation("Story.CreatedBy").
		Relation("Story.Turns").
		Relation("Story.CurrentWriter").
		Select(); err != nil {
		return nil, err
	}

	return &story, nil

}

func NewScribly(db *pg.DB, emailer EmailGateway) (*Scribly, error) {
	return &Scribly{
		db:      db,
		emailer: emailer,
	}, nil
}
