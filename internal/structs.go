package internal

import (
	"fmt"
	"scribly/pkg/helpers"
)

type EmailVerificationState string

const (
	EmailVerificationStatePending  = EmailVerificationState("pending")
	EmailVerificationStateVerified = EmailVerificationState("verified")
)

type User struct {
	ID                      int
	Username                string
	Email                   string
	EmailVerificationStatus EmailVerificationState
}

type TurnAction string

const (
	TurnActionPass           = TurnAction("pass")
	TurnActionWrite          = TurnAction("write")
	TurnActionFinish         = TurnAction("finish")
	TurnActionWriteAndFinish = TurnAction("write_and_finish")
)

type Turn struct {
	TakenByID   int
	TakenBy     *User `pg:"rel:has-one"`
	Action      TurnAction
	TextWritten string
}

type StoryState string

const (
	StoryStateDraft      = StoryState("draft")
	StoryStateInProgress = StoryState("in_progress")
	StoryStateDone       = StoryState("done")
)

type Story struct {
	ID          int
	Title       string
	State       StoryState
	CreatedByID int
	CreatedBy   *User  `pg:"rel:has-one"`
	Cowriters   []User `pg:"rel:has-many"`
	Turns       []Turn `pg:"rel:has-many"`
}

func (s *Story) CurrentWritersTurn() *User {
	if s.State != StoryStateDraft {
		return nil
	}

	currentWriterIndex := len(s.Turns) % len(s.Cowriters)
	return &s.Cowriters[currentWriterIndex]
}

// this should be a view
type UserStory struct {
	UserID  int
	StoryID int
	Story   *Story `pg:"rel:has-one"`
	Hidden  bool
}

type Me struct {
	User    *User
	Stories []UserStory
}

type Email struct {
	Subject string
	Body    string
	To      User
}

type EmailVerificationTokenPayload struct {
	UserID    int
	Email     string
	Timestamp float64
}

type SignUpInput struct {
	Username string `schema:"username,required"`
	Password string `schema:"password,required"`
	Email    string `schema:"email,required"`
}

func (s *SignUpInput) Validate() error {
	if len(s.Password) < 8 {
		return fmt.Errorf("Password must be longer than 8 characters")
	}

	if len(s.Username) < 4 || !helpers.IsAlphaNumeric(s.Username) {
		return fmt.Errorf("Username must be longer than 4 characters and only consist of alphanumeric characters")
	}

	if !helpers.IsValidEmail(s.Email) {
		return fmt.Errorf("Invalid email address format")
	}

	return nil
}
