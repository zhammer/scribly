package internal

import (
	"context"
	"database/sql"
	"scribly/pkg/db"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

type Scribly struct {
	db             *bun.DB
	emailer        EmailGateway
	messageGateway MessageGateway
}

func (s *Scribly) LogIn(ctx context.Context, input LoginInput) (*User, error) {
	user := User{}
	err := s.db.NewRaw(`
		SELECT id, username, email, email_verification_status FROM users WHERE username = ? AND password = crypt(?, password);
	`, input.Username, input.Password).Scan(ctx, &user)
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
	err := s.db.NewRaw(`
		INSERT INTO users (username, email, password) VALUES (?, ?, crypt(?, gen_salt('bf', 8)))
		RETURNING id, username, email, email_verification_status
	`, input.Username, input.Email, input.Password).Scan(ctx, &user)
	if err != nil {
		return nil, err
	}

	s.messageGateway.AnnounceUserCreated(ctx, user)

	return &user, nil
}

func (s *Scribly) Me(ctx context.Context, user *User) (*Me, error) {
	// todo: figure out how to do this in one query, if that's bettah
	// refresh our user model
	if err := s.db.NewSelect().Model(user).WherePK().Scan(ctx); err != nil {
		return nil, err
	}
	var userStories []UserStory
	if err := s.db.NewSelect().Model(&userStories).
		Where("user_id = ?", user.ID).
		Relation("Story").
		Scan(ctx); err != nil {
		return nil, err
	}

	return &Me{
		User:    user,
		Stories: userStories,
	}, nil

}

func (s *Scribly) UserStory(ctx context.Context, user User, storyID int) (*UserStory, error) {
	story := UserStory{}
	if err := s.db.NewSelect().Model(&story).
		Where("story_id = ? AND user_id = ?", storyID, user.ID).
		Relation("Story").
		Relation("Story.Cowriters", db.WithOrderBy("story_cowriter.turn_index")).
		Relation("Story.Cowriters.User").
		Relation("Story.CreatedByU").
		Relation("Story.Turns", db.WithOrderBy("turn.created_at")).
		Relation("Story.CurrentWriter").
		Scan(ctx); err != nil {
		return nil, err
	}

	return &story, nil

}

// this is the super simple way to do this while there are only a handful of users
// on the site. return everyone who's not the current user.
func (s *Scribly) UserSuggestions(ctx context.Context, user User) ([]User, error) {
	var users []User
	if err := s.db.NewSelect().Model(&users).Where("id != ?", user.ID).Scan(ctx); err != nil {
		return users, err
	}

	return users, nil
}

func (s *Scribly) StartStory(ctx context.Context, user User, input StartStoryInput) (*Story, error) {
	story := Story{
		Title:       input.Title,
		State:       StoryStateDraft,
		CreatedByID: user.ID,
	}
	err := s.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		_, err := tx.NewInsert().Model(&story).ExcludeColumn("current_writer_id").Exec(ctx)
		if err != nil {
			return err
		}

		firstTurn := Turn{
			StoryID:   story.ID,
			TakenByID: user.ID,
			Action:    TurnActionWrite,
			Text:      input.Text,
		}
		if _, err := tx.NewInsert().Model(&firstTurn).Exec(ctx); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if err := s.db.NewSelect().Model(&story).WherePK().Relation("Turns").Scan(ctx); err != nil {
		return nil, err
	}

	return &story, nil
}

func (s *Scribly) AddCowriters(ctx context.Context, user User, storyID int, input AddCowritersInput) error {
	story := Story{ID: storyID}
	err := s.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		if err := tx.NewSelect().Model(&story).WherePK().Scan(ctx); err != nil {
			return err
		}

		if err := story.ValidateUserCanAddCowriters(user, input.Usernames()); err != nil {
			return err
		}

		var cowriterUsers []User
		if err := tx.NewSelect().Model(&cowriterUsers).Where("username ILIKE ANY (?)", pgdialect.Array(input.Usernames())).Scan(ctx); err != nil {
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
		if _, err := tx.NewInsert().Model(&cowriters).Exec(ctx); err != nil {
			return err
		}
		story.State = StoryStateInProgress
		if _, err := tx.NewUpdate().Model(&story).WherePK().ExcludeColumn("current_writer_id").Exec(ctx); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	s.messageGateway.AnnounceCowritersAdded(ctx, story)

	return nil
}

func (s *Scribly) TakeTurn(ctx context.Context, user User, storyID int, input TurnInput) error {
	turn := Turn{
		StoryID:   storyID,
		TakenByID: user.ID,
		Action:    input.Action,
		Text:      input.Text,
	}
	if err := turn.Validate(); err != nil {
		return err
	}

	story := Story{ID: storyID}

	err := s.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		if err := tx.NewSelect().Model(&story).WherePK().Relation("Turns").Scan(ctx); err != nil {
			return err
		}

		if err := story.ValidateUserCanTakeTurn(user, turn); err != nil {
			return err
		}

		if _, err := tx.NewInsert().Model(&turn).Exec(ctx); err != nil {
			return err
		}

		if turn.Finishes() {
			story.State = StoryStateDone
			if _, err := tx.NewUpdate().Model(&story).WherePK().ExcludeColumn("current_writer_id").Exec(ctx); err != nil {
				return err
			}
		}

		story.Turns = append(story.Turns, turn)
		return nil
	})

	if err != nil {
		return err
	}

	s.messageGateway.AnnounceTurnTaken(ctx, story)

	return nil
}

func (s *Scribly) Hide(ctx context.Context, user User, storyID int, hiddenStatus HiddenStatus) error {
	story := Story{ID: storyID}
	if err := s.db.NewSelect().Model(&story).WherePK().Relation("Cowriters").Scan(ctx); err != nil {
		return err
	}

	hide := UserStoryHide{
		UserID:       user.ID,
		StoryID:      story.ID,
		HiddenStatus: hiddenStatus,
	}

	if err := story.ValidateCanHide(user, hide); err != nil {
		return err
	}

	if _, err := s.db.NewInsert().Model(&hide).On("CONFLICT (user_id, story_id) DO UPDATE").Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Scribly) Nudge(ctx context.Context, nudger User, nudgeeID int, storyID int) error {
	nudgee := User{ID: nudgeeID}
	if err := s.db.NewSelect().Model(&nudgee).WherePK().Scan(ctx); err != nil {
		return err
	}

	story := Story{ID: storyID}
	if err := s.db.NewSelect().Model(&story).WherePK().Relation("Cowriters").Scan(ctx); err != nil {
		return err
	}

	if err := story.ValidateCanNudge(nudger, nudgee); err != nil {
		return err
	}

	email, err := BuildNudgeEmail(nudger, nudgee, story)
	if err != nil {
		return err
	}

	return s.emailer.SendEmail(ctx, *email)
}

func (s *Scribly) SendAddedToStoryEmails(ctx context.Context, storyID int) error {
	story := Story{ID: storyID}
	if err := s.db.NewSelect().Model(&story).
		WherePK().
		Relation("Cowriters").
		Relation("Cowriters.User").
		Relation("CurrentWriter").
		Relation("Turns").
		Relation("CreatedByU").
		Scan(ctx); err != nil {
		return err
	}

	emails, err := BuildAddedToStoryEmails(story)
	if err != nil {
		return err
	}

	for _, email := range emails {
		if err := s.emailer.SendEmail(ctx, email); err != nil {
			return err
		}
	}

	return nil
}

func (s *Scribly) SendTurnEmailNotifications(ctx context.Context, storyID int, turnNumber int) error {
	story := Story{ID: storyID}
	if err := s.db.NewSelect().Model(&story).WherePK().
		Relation("Turns", db.WithOrderBy("turn.created_at")).
		Relation("Turns.TakenByU").
		Relation("Cowriters", db.WithOrderBy("story_cowriter.turn_index")).
		Relation("Cowriters.User").
		Relation("CurrentWriter").
		Scan(ctx); err != nil {
		return err
	}

	emails, err := BuildTurnNotificationEmails(story, turnNumber)
	if err != nil {
		return err
	}

	for _, email := range emails {
		if err := s.emailer.SendEmail(ctx, email); err != nil {
			return err
		}
	}

	return nil
}

func (s *Scribly) SendVerificationEmail(ctx context.Context, userID int) error {
	user := User{ID: userID}
	if err := s.db.NewSelect().Model(&user).WherePK().Scan(ctx); err != nil {
		return err
	}

	if err := user.ValidateCanSendVerificationEmail(); err != nil {
		return err
	}

	verificationToken, err := buildEmailVerificationToken(user)
	if err != nil {
		return err
	}

	email, err := BuildEmailVerificationEmail(user, verificationToken)
	if err != nil {
		return err
	}

	if err := s.emailer.SendEmail(ctx, *email); err != nil {
		return err
	}

	return nil
}

func (s *Scribly) VerifyEmail(ctx context.Context, user User, token string) error {
	emailVerificationPayload, err := parseEmailVerificationToken(token)
	if err != nil {
		return err
	}

	if err := user.ValidateEmailVerification(*emailVerificationPayload); err != nil {
		return nil
	}

	user.EmailVerificationStatus = EmailVerificationStateVerified
	if _, err := s.db.NewUpdate().Model(&user).WherePK().Column("email_verification_status").Exec(ctx); err != nil {
		return err
	}

	return nil
}

func NewScribly(db *bun.DB, emailer EmailGateway, messageGateway MessageGateway) (*Scribly, error) {
	return &Scribly{
		db:             db,
		emailer:        emailer,
		messageGateway: messageGateway,
	}, nil
}
