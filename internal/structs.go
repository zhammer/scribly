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
	StoryID     int
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
	ID              int
	Title           string
	State           StoryState
	CreatedByID     int    `pg:"created_by"`
	CreatedBy       *User  `pg:"rel:has-one"`
	Cowriters       []User `pg:"rel:has-many"`
	Turns           []Turn `pg:"rel:has-many"`
	CurrentWriterID int    `pg:"-"`
}

func (s *Story) CurrentWritersTurn() *User {
	if s.State != StoryStateDraft {
		return nil
	}

	currentWriterIndex := len(s.Turns) % len(s.Cowriters)
	return &s.Cowriters[currentWriterIndex]
}

// this should be a view
// todo: add field to check if it's the user's turn
type UserStory struct {
	UserID  int
	StoryID int
	Story   Story `pg:"rel:has-one"`
	Hidden  bool
}

type Me struct {
	User    *User
	Stories []UserStory
}

func (m *Me) storiesWithState(state StoryState) []UserStory {
	var out []UserStory
	for _, story := range m.Stories {
		if story.Story.State == state {
			out = append(out, story)
		}
	}
	return out
}

func (m *Me) Drafts() []UserStory {
	return m.storiesWithState(StoryStateDraft)
}
func (m *Me) YourTurn() []UserStory {
	// todo: check turn
	return m.storiesWithState(StoryStateInProgress)
}
func (m *Me) WaitingForOthers() []UserStory {
	// todo: check turn
	return m.storiesWithState(StoryStateInProgress)
}
func (m *Me) Done() []UserStory {
	return m.storiesWithState(StoryStateDone)
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
