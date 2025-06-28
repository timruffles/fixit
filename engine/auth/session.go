package auth

import (
	"net/http"

	"github.com/aarondl/authboss/v3"
	"github.com/gorilla/sessions"
)

type SessionStorer struct {
	store *sessions.CookieStore
}

func NewSessionStorer(key []byte) *SessionStorer {
	return &SessionStorer{
		store: sessions.NewCookieStore(key),
	}
}

var _ authboss.ClientStateReadWriter = (*SessionStorer)(nil)

func (s *SessionStorer) ReadState(r *http.Request) (authboss.ClientState, error) {
	session, err := s.store.Get(r, "fixit_session")
	if err != nil {
		return nil, err
	}

	state := &SessionState{
		session: session,
		request: r,
	}

	return state, nil
}

func (s *SessionStorer) WriteState(w http.ResponseWriter, state authboss.ClientState, ev []authboss.ClientStateEvent) error {
	sessionState := state.(*SessionState)
	return sessionState.session.Save(sessionState.request, w)
}

type SessionState struct {
	session *sessions.Session
	request *http.Request
}

func (s *SessionState) Get(key string) (string, bool) {
	val, ok := s.session.Values[key]
	if !ok {
		return "", false
	}
	str, ok := val.(string)
	return str, ok
}

func (s *SessionState) Put(key, value string) {
	s.session.Values[key] = value
}

func (s *SessionState) Del(key string) {
	delete(s.session.Values, key)
}
