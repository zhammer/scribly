package internal

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
	StoryID int
	Story   *Story `pg:"rel:has-one"`
	Hidden  bool
}

type Me struct {
	UserID  int
	User    *User        `pg:"rel:has-one"`
	Stories []*UserStory `pg:"rel:has-many"`
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
