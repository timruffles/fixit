package post

import (
	"bytes"
	_ "embed"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/timr/oss/fixit/web/handler"
	"github.com/timr/oss/fixit/web/layouts"
)

//go:embed templates/create.gohtml
var createTplS string

var createTpl = template.Must(template.New("create").Parse(createTplS))

type CreatePostData struct {
	Title string
	Body  string
	Tags  string
	Error string
}

func CreatePostGetHandler(r *http.Request) (handler.Response, error) {
	var data CreatePostData
	content, err := renderCreatePost(data)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return handler.Ok(content), nil
}

func CreatePostPostHandler(r *http.Request) (handler.Response, error) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return handler.BadInput([]byte("Failed to parse form")), nil
	}

	title := r.FormValue("title")
	body := r.FormValue("body")
	tags := r.FormValue("tags")

	data := CreatePostData{
		Title: title,
		Body:  body,
		Tags:  tags,
	}

	if title == "" {
		data.Error = "Title is required"
		content, err := renderCreatePost(data)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return handler.BadInput(content), nil
	}

	// TODO: Save post to backend

	content, err := renderCreatePost(CreatePostData{})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return handler.Ok(content), nil
}

func renderCreatePost(data CreatePostData) ([]byte, error) {
	var content bytes.Buffer
	err := createTpl.Execute(&content, data)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	layoutData := layouts.LayoutData{
		Title:   "Create Post",
		Content: template.HTML(content.String()),
	}

	return layouts.WithGeneral(layoutData)
}

type Handler struct{}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/post/create", handler.Wrap(CreatePostGetHandler)).Methods("GET")
	router.HandleFunc("/post/create", handler.Wrap(CreatePostPostHandler)).Methods("POST")
}