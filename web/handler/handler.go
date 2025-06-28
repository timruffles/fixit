package handler

import (
	"errors"
	"net/http"
)

// Fn is a high level handler that synchrounsly returns (Response, error) rather than
// imperatively controlling the response
type Fn func(r *http.Request) (Response, error)

// Wrap a handler at a high level
func Wrap(fn Fn) func(http.ResponseWriter,
	*http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		resI, err := fn(request)
		if err != nil {
			writer.WriteHeader(500)
			return
		}

		res, ok := resI.(*ResponseBuffered)
		if !ok {
			err = errors.New("unimeplemented response type")
		}

		st := res.Status
		if st == 0 {
			st = 200
		}
		writer.WriteHeader(st)
		writer.Write(res.Content)
	}

}

type ResponseBuffered struct {
	Content []byte
	// Status defaults to 200
	Status int
}

func (r ResponseBuffered) isResponse() {}

var _ Response = ResponseBuffered{}

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
