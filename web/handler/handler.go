package handler

import (
	"log/slog"
	"net/http"
	"net/url"

	"github.com/aarondl/authboss/v3"
	"github.com/gorilla/sessions"

	"fixit/web/errors"
)

// Fn is a high level handler that synchrounsly returns (Response, error) rather than
// imperatively controlling the response
type Fn func(r *http.Request) (Response, error)

type Register interface {
	Route(method string, path string, fn Fn)
	Get(method string, path string, fn Fn)
	Post(method string, path string, fn Fn)
}

// Wrap a handler at a high level
func Wrap(fn Fn) func(http.ResponseWriter,
	*http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		resI, err := fn(request)
		if err != nil {
			slog.Error("error in handler", "err", err)
			errors.Handle500(writer, request, err)
			return
		}

		switch res := resI.(type) {
		case *ResponseBuffered:
			st := res.Status
			if st == 0 {
				st = 200
			}
			writer.WriteHeader(st)
			writer.Write(res.Content)
		case *Redirect:
			if res.Permanent {
				http.Redirect(writer, request, res.To.String(), http.StatusMovedPermanently)
			} else {
				http.Redirect(writer, request, res.To.String(), http.StatusFound)
			}
		case *RedirectWithSession:
			// Save session before redirect
			if err := res.Session.Save(request, writer); err != nil {
				slog.Error("error saving session", "err", err)
				errors.Handle500(writer, request, err)
				return
			}
			http.Redirect(writer, request, res.To, http.StatusFound)
		case *AuthRequiredRedirect:
			// Save authboss state before redirect - for now just redirect
			// The flash message should already be set in the state
			http.Redirect(writer, request, res.To, http.StatusFound)
		default:
			slog.Error("unknown response type")
			errors.Handle500(writer, request, nil)
		}
	}

}

type ResponseBuffered struct {
	Content []byte
	// Status defaults to 200
	Status int
}

func (r ResponseBuffered) isResponse() {}

type Redirect struct {
	To        url.URL
	Permanent bool
}

func (r *Redirect) isResponse() {}

var _ Response = &Redirect{}

type RedirectWithSession struct {
	To      string
	Session *sessions.Session
	Request *http.Request
}

func (r *RedirectWithSession) isResponse() {}

var _ Response = &RedirectWithSession{}

type AuthRequiredRedirect struct {
	State authboss.ClientState
	AB    *authboss.Authboss
	To    string
}

func (a *AuthRequiredRedirect) isResponse() {}

var _ Response = &AuthRequiredRedirect{}

func BadInput(content []byte) Response {
	return &ResponseBuffered{
		Status:  400,
		Content: content,
	}
}

func NotFound(content []byte) Response {
	return &ResponseBuffered{
		Status:  400,
		Content: content,
	}
}

func Ok(content []byte) Response {
	return &ResponseBuffered{
		Status:  200,
		Content: content,
	}
}

type Response interface {
	isResponse()
}

func RedirectTo(to string) Response {
	u, _ := url.Parse(to)
	return &Redirect{
		To:        *u,
		Permanent: false,
	}
}

func RedirectWithSessionTo(to string, session *sessions.Session, r *http.Request) Response {
	return &RedirectWithSession{
		To:      to,
		Session: session,
		Request: r,
	}
}
