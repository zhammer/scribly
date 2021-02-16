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
		Relation("Story.CreatedByU").
		Relation("Story.Turns").
		Relation("Story.CurrentWriter").
		Select(); err != nil {
		return nil, err
	}

	return &story, nil

}

func (s *Scribly) StartStory(ctx context.Context, user User, input StartStoryInput) (*Story, error) {
	story := Story{
		Title:       input.Title,
		State:       StoryStateDraft,
		CreatedByID: user.ID,
	}
	err := s.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		_, err := tx.Model(&story).ExcludeColumn("current_writer_id").Insert()
		if err != nil {
			return err
		}

		firstTurn := Turn{
			StoryID:   story.ID,
			TakenByID: user.ID,
			Action:    TurnActionWrite,
			Text:      input.Text,
		}
		if _, err := tx.Model(&firstTurn).Insert(); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if err := s.db.Model(&story).WherePK().Relation("Turns").Select(); err != nil {
		return nil, err
	}

	return &story, nil
}

func (s *Scribly) AddCowriters(ctx context.Context, user User, storyID int, input AddCowritersInput) error {
	story := Story{ID: storyID}
	err := s.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		if err := tx.Model(&story).Select(); err != nil {
			return err
		}

		if err := story.ValidateUserCanAddCowriters(user, input.Usernames()); err != nil {
			return err
		}

		var cowriterUsers []User
		if err := tx.Model(&cowriterUsers).WhereIn("username IN (?)", input.Usernames()).Select(); err != nil {
			return err
		}

		if err := input.ValidateAllFound(cowriterUsers); err != nil {
			return err
		}

		var cowriters []StoryCowriter
		for index, cowriter := range append([]User{user}, cowriterUsers...) {
			cowriters = append(cowriters, StoryCowriter{
				UserID:    cowriter.ID,
				StoryID:   storyID,
				TurnIndex: index,
			})
		}
		if _, err := tx.Model(&cowriters).Insert(); err != nil {
			return err
		}
		story.State = StoryStateInProgress
		if _, err := tx.Model(&story).WherePK().ExcludeColumn("current_writer_id").Update(); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func NewScribly(db *pg.DB, emailer EmailGateway) (*Scribly, error) {
	return &Scribly{
		db:      db,
		emailer: emailer,
	}, nil
}
