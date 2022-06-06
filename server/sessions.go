package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/jjlock/bank/server/validate"
)

type sessionStore interface {
	create(w http.ResponseWriter, r *http.Request, username string) error
	getUser(r *http.Request) (string, error)
	expire(w http.ResponseWriter, r *http.Request) error
}

type cookieStore struct {
	store *sessions.CookieStore
}

func newCookieStore() sessionStore {
	cs := sessions.NewCookieStore(securecookie.GenerateRandomKey(32))
	// set maximum cookie age to 15 minutes
	cs.MaxAge(900)

	return &cookieStore{cs}
}

func (cs *cookieStore) create(w http.ResponseWriter, r *http.Request, username string) error {
	if err := validate.Username(username); err != nil {
		return err
	}

	session, err := cs.store.Get(r, "bank-session")
	if err != nil {
		return err
	}

	session.Values["user"] = username

	err = session.Save(r, w)
	if err != nil {
		return err
	}

	return nil
}

func (cs *cookieStore) getUser(r *http.Request) (string, error) {
	session, err := cs.store.Get(r, "bank-session")
	if err != nil {
		return "", err
	}

	username, ok := session.Values["user"].(string)
	if !ok {
		return "", errors.New("client not logged in")
	}

	if err := validate.Username(username); err != nil {
		return "", fmt.Errorf("session contains invalid username: %w", err)
	}

	return username, nil
}

func (cs *cookieStore) expire(w http.ResponseWriter, r *http.Request) error {
	session, err := cs.store.Get(r, "bank-session")
	if err != nil {
		return err
	}

	session.Values["user"] = ""
	session.Options.MaxAge = -1

	err = session.Save(r, w)
	if err != nil {
		return err
	}

	return nil
}
