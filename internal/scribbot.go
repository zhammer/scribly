package internal

import (
	"context"
	"sync"
)

type ScribbotOption func(s *Scribbot)

type Scribbot struct {
	*Scribly
	scribbotUserLock           sync.Mutex
	scribbotUser               *User
	scribbotShouldWriteChooser func() bool
}

func NewScribbot(scribly *Scribly, opts ...ScribbotOption) *Scribbot {
	s := Scribbot{
		Scribly:                    scribly,
		scribbotShouldWriteChooser: func() bool { return true },
	}
	for _, opt := range opts {
		opt(&s)
	}
	return &s
}

func (s *Scribbot) getScribbotUser() (*User, error) {
	s.scribbotUserLock.Lock()
	defer s.scribbotUserLock.Unlock()

	if s.scribbotUser != nil {
		return s.scribbotUser, nil
	}

	user := User{}
	if err := s.db.Model(&user).Where("username = ?", scribbotUsername).Select(); err != nil {
		return nil, err
	}

	s.scribbotUser = &user
	return s.scribbotUser, nil
}

func (s *Scribly) TakeScribbotTurn(ctx context.Context, story Story, scribbot User) error {
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
	scribbot, err := s.getScribbotUser()
	if err != nil {
		return err
	}

	var stories []Story // find all stories where it's AI's turn

	for _, story := range stories {
		if !s.scribbotShouldWriteChooser() {
			continue
		}
		if err := s.TakeScribbotTurn(ctx, story, *scribbot); err != nil {
			return err
		}
	}

	return nil

}
