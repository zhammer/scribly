package main

import (
	"encoding/gob"
	"net/http"

	"scribly/internal"

	"github.com/gorilla/sessions"
	"github.com/mitchellh/mapstructure"
)

type SessionHelper struct {
	store       *sessions.CookieStore
	sessionName string
}

func (s *SessionHelper) GetUser(r *http.Request) (*internal.User, error) {
	session, _ := s.store.Get(r, s.sessionName)
	user := internal.User{}
	if err := mapstructure.Decode(session.Values["user"], &user); err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, nil
	}
	return &user, nil
}

func (s *SessionHelper) SaveUser(user *internal.User, r *http.Request, w http.ResponseWriter) error {
	session, _ := s.store.Get(r, s.sessionName)
	session.Values["user"] = *user
	return session.Save(r, w)
}

func NewSessionHelper(secret string) *SessionHelper {
	cookieStore := sessions.NewCookieStore([]byte(secret))
	return &SessionHelper{store: cookieStore, sessionName: "scribly-session"}
}

func init() {
	gob.Register(internal.User{})
}
