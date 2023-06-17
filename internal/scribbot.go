package internal

import (
	"context"
	"fmt"
	"scribly/pkg/db"
	"sync"
)

type ScribbotOption func(s *Scribbot)

type Scribbot struct {
	*Scribly
	openai                     OpenAIGateway
	scribbotShouldWriteChooser func() bool

	scribbotUserLock sync.Mutex
	scribbotUser     *User
}

func NewScribbot(scribly *Scribly, openai OpenAIGateway, opts ...ScribbotOption) *Scribbot {
	s := Scribbot{
		Scribly:                    scribly,
		openai:                     openai,
		scribbotShouldWriteChooser: func() bool { return true },
	}
	for _, opt := range opts {
		opt(&s)
	}
	return &s
}

func (s *Scribbot) getScribbotUser(ctx context.Context) (*User, error) {
	s.scribbotUserLock.Lock()
	defer s.scribbotUserLock.Unlock()

	if s.scribbotUser != nil {
		return s.scribbotUser, nil
	}

	user := User{}
	if _, err := s.db.NewSelect().Model(&user).Where("username = ?", scribbotUsername).Exec(ctx); err != nil {
		return nil, err
	}

	s.scribbotUser = &user
	return s.scribbotUser, nil
}

func (s *Scribbot) takeScribbotTurn(ctx context.Context, story Story, scribbot User) error {
	fmt.Printf("scribbot taking turn on story %d\n", story.ID)
	text, err := s.openai.GenerateTurnText(ctx, story)
	if err != nil {
		return err
	}

	action := TurnActionWrite
	if Odds(1, 20) {
		action = TurnActionWriteAndFinish
	}

	turnInput := TurnInput{
		Text:   text,
		Action: action,
	}

	return s.TakeTurn(ctx, scribbot, story.ID, turnInput)
}

func (s *Scribbot) TakeScribbotTurns(ctx context.Context) error {
	scribbot, err := s.getScribbotUser(ctx)
	if err != nil {
		return err
	}

	var stories []Story
	if _, err := s.db.NewSelect().Model(&stories).
		Where("current_writer_id = ?", scribbot.ID).
		Relation("Turns", db.WithOrderBy("turn.created_at")).
		Exec(ctx); err != nil {
		return err
	}
	fmt.Printf("found %d stories where it is scribbot's turn\n", len(stories))

	var storiesToTakeTurn []Story
	for _, story := range stories {
		if s.scribbotShouldWriteChooser() {
			storiesToTakeTurn = append(storiesToTakeTurn, story)
		}
	}
	fmt.Printf("scribbot will take its turn on %d stories\n", len(storiesToTakeTurn))

	for _, story := range storiesToTakeTurn {
		if err := s.takeScribbotTurn(ctx, story, *scribbot); err != nil {
			return err
		}
	}

	return nil

}
