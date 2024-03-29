package internal

import (
	"fmt"
	"scribly/pkg/helpers"
	"strings"
	"time"

	"github.com/uptrace/bun"
)

type EmailVerificationState string

const (
	EmailVerificationStatePending  = EmailVerificationState("pending")
	EmailVerificationStateVerified = EmailVerificationState("verified")
)

type User struct {
	ID                      int `bun:",pk,autoincrement"`
	Username                string
	Email                   string
	EmailVerificationStatus EmailVerificationState
}

func (u *User) IsScribbot() bool {
	return u.Username == scribbotUsername
}

func (u *User) ValidateCanSendVerificationEmail() error {
	if u.EmailVerificationStatus == EmailVerificationStateVerified {
		return fmt.Errorf("Email already verified.")
	}
	return nil
}

func (u *User) ValidateEmailVerification(emailVerification EmailVerificationPayload) error {
	if u.EmailVerificationStatus == EmailVerificationStateVerified {
		return fmt.Errorf("Email already verified")
	}

	if emailVerification.Email != u.Email {
		return fmt.Errorf("Email in payload doesn't match user's email")
	}

	if time.Now().Sub(time.Unix(emailVerification.Timestamp, 0)) > (time.Hour * 24) {
		return fmt.Errorf("Email token expored")
	}

	return nil
}

type TurnAction string

const (
	TurnActionPass           = TurnAction("pass")
	TurnActionWrite          = TurnAction("write")
	TurnActionFinish         = TurnAction("finish")
	TurnActionWriteAndFinish = TurnAction("write_and_finish")
)

type Turn struct {
	StoryID   int
	TakenByID int `bun:"taken_by"`
	// this can't be called TakenBy.. go-pg gets confused, change the underlying column names soon
	TakenByU User `bun:"rel:belongs-to,join:taken_by=id"`
	Action   TurnAction
	Text     string `bun:"text_written"`
}

func (t *Turn) Validate() error {
	if t.Writes() && t.Text == "" {
		return fmt.Errorf("Text for a `write` turn cannot be empty.")
	}

	switch t.Action {
	case TurnActionFinish, TurnActionPass, TurnActionWrite, TurnActionWriteAndFinish:
	default:
		return fmt.Errorf("Unknown turn action '%s'", t.Action)
	}

	return nil
}

func (t *Turn) Finishes() bool {
	switch t.Action {
	case TurnActionFinish, TurnActionWriteAndFinish:
		return true
	}
	return false
}

func (t *Turn) Writes() bool {
	switch t.Action {
	case TurnActionWrite, TurnActionWriteAndFinish:
		return true
	}
	return false
}

type StoryCowriter struct {
	StoryID   int
	Story     Story `bun:"rel:belongs-to"`
	UserID    int
	User      User `bun:"rel:belongs-to"`
	TurnIndex int
}

type StoryState string

const (
	StoryStateDraft      = StoryState("draft")
	StoryStateInProgress = StoryState("in_progress")
	StoryStateDone       = StoryState("done")
)

type Story struct {
	bun.BaseModel   `bun:"select:stories_enhanced"`
	ID              int `bun:",pk,autoincrement"`
	Title           string
	State           StoryState
	CreatedByID     int             `bun:"created_by"`
	CreatedByU      *User           `bun:"rel:has-one,join:created_by=id"`
	Cowriters       []StoryCowriter `bun:"rel:has-many"`
	Turns           []Turn          `bun:"rel:has-many"`
	CurrentWriterID int
	CurrentWriter   *User `bun:"rel:belongs-to"`
}

func (s *Story) userInvolvedWithStory(user User) bool {
	if s.CreatedByID == user.ID {
		return true
	}

	for _, cowriter := range s.Cowriters {
		if cowriter.UserID == user.ID {
			return true
		}
	}

	return false
}

func (s *Story) ValidateCanNudge(nudger User, nudgee User) error {
	if !s.userInvolvedWithStory(nudger) {
		return fmt.Errorf("You can't send a nudge for a story you're not a part of!")
	}

	if nudgee.ID != s.CurrentWriterID {
		return fmt.Errorf("It's not %d's turn!", nudgee.ID)
	}

	if nudgee.EmailVerificationStatus != EmailVerificationStateVerified {
		return fmt.Errorf("%s hasn't verified their email yet!", nudgee.Username)
	}

	return nil
}

func (s *Story) ValidateCanHide(user User, hide UserStoryHide) error {
	if !s.userInvolvedWithStory(user) {
		return fmt.Errorf("User is not involved with story %d", s.ID)
	}
	return nil
}

func (s *Story) ValidateUserCanAddCowriters(user User, cowriters []string) error {
	if s.CreatedByID != user.ID {
		return fmt.Errorf("User %d cannot add cowriters to story %d created by %d", user.ID, s.ID, s.CreatedByID)
	}

	if s.State != StoryStateDraft {
		return fmt.Errorf("Story must be in state 'draft' to add cowriters. Story %d is in state %s.", s.ID, s.State)
	}

	if helpers.ContainsStr(cowriters, user.Username, strings.ToLower) {
		return fmt.Errorf("You cannot add yourself as a cowriter.")
	}

	return nil
}

func (s *Story) ValidateUserCanTakeTurn(user User, turn Turn) error {
	if s.CurrentWriterID != user.ID {
		return fmt.Errorf("It is not user %d's turn!", user.ID)
	}

	return nil
}

type UserStory struct {
	UserID  int `bun:",pk"`
	StoryID int
	Story   Story `bun:"rel:belongs-to"`
	Hidden  bool
}

func (u *UserStory) IsUsersTurn() bool {
	return u.Story.CurrentWriterID == u.UserID
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
	var yourTurn []UserStory
	inProgress := m.storiesWithState(StoryStateInProgress)
	for _, story := range inProgress {
		if story.IsUsersTurn() {
			yourTurn = append(yourTurn, story)
		}
	}
	return yourTurn
}
func (m *Me) WaitingForOthers() []UserStory {
	var waitingForOthers []UserStory
	inProgress := m.storiesWithState(StoryStateInProgress)
	for _, story := range inProgress {
		if !story.IsUsersTurn() {
			waitingForOthers = append(waitingForOthers, story)
		}
	}
	return waitingForOthers
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

type LoginInput struct {
	Username string `schema:"username,required"`
	Password string `schema:"password,required"`
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

type StartStoryInput struct {
	Title string `schema:"title,required"`
	Text  string `schema:"body,required"`
}

type AddCowritersInput struct {
	Person1 string `schema:"person-1,required"`
	Person2 string `schema:"person-2"`
	Person3 string `schema:"person-3"`
}

func (a *AddCowritersInput) Usernames() []string {
	usernames := []string{a.Person1}
	if a.Person2 != "" {
		usernames = append(usernames, a.Person2)
	}
	if a.Person3 != "" {
		usernames = append(usernames, a.Person3)
	}
	return usernames
}

type TurnInput struct {
	Action TurnAction `schema:"action,required"`
	Text   string     `schema:"text"`
}

// TODO: this should be smarter, specifically should return invalid usernames
// i'm just too lazy to be so bold as to write a SET in go.
func (a *AddCowritersInput) ValidateAllFound(users []User) error {
	if len(users) != len(a.Usernames()) {
		return fmt.Errorf("not all users were found")
	}

	return nil
}

type HiddenStatus string

const (
	HiddenStatusHidden   = HiddenStatus("hidden")
	HiddenStatusUnhidden = HiddenStatus("unhidden")
)

type UserStoryHide struct {
	UserID       int `bun:",pk"`
	StoryID      int `bun:",pk"`
	HiddenStatus HiddenStatus
}

type EmailVerificationPayload struct {
	UserID    int
	Email     string
	Timestamp int64
}
