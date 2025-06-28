package handler

import "net/http"

type Fn func(r *http.Request) (*Response, error)

func Wrap(fn Fn) func(http.ResponseWriter,
	*http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		res, err := fn(request)
		if err != nil {
			writer.WriteHeader(500)
			return
		}

		st := res.Status
		if st == 0 {
			st = 200
		}
		writer.WriteHeader(st)
		writer.Write(res.Content)
	}

}

func Ok(content []byte) *Response {
	return &Response{
		Status:  200,
		Content: content,
	}
}

type Response struct {
	Status  int
	Content []byte
}
