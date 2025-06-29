package list

import (
	"bytes"
	"context"
	"embed"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"fixit/engine/community"
	"fixit/engine/ent"
	"fixit/web/handler"
	"fixit/web/layouts"
)

//go:embed templates/*.gohtml
var templatesFS embed.FS

func getTemplates() *template.Template {
	// In development, parse templates from disk for hot reloading
	// In production, use embedded templates
	if templates, err := template.ParseGlob("web/list/templates/*.gohtml"); err == nil {
		return templates
	}
	// Fallback to embedded templates
	return template.Must(template.ParseFS(templatesFS, "templates/*.gohtml"))
}

type Handler struct {
	repo *community.Repository
}

type Post struct {
	Title    string
	Reporter string
	Status   string
	Comments int
	Category string
	TimeAgo  string
	Priority string
}

func New(client *ent.Client) *Handler {
	repo := community.NewRepository(client)
	return &Handler{
		repo: repo,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/c/{slug}", handler.Wrap(h.handleList)).Methods("GET")
}

func (h *Handler) handleList(req *http.Request) (handler.Response, error) {
	ctx := context.Background()
	filter := &community.Filter{}

	vars := mux.Vars(req)
	communitySlug := vars["slug"]

	// Get community data
	comm, err := h.repo.GetBySlug(ctx, communitySlug)
	if err != nil {
		return nil, err
	}

	postItems, err := h.repo.ListPosts(ctx, communitySlug, filter)
	if err != nil {
		return nil, err
	}

	data := struct {
		Community *ent.Community
		Posts     []community.PostListItem
	}{
		Community: comm,
		Posts:     postItems,
	}

	content, err := templatesExecute("list.gohtml", data)
	if err != nil {
		return nil, err
	}

	html, err := layouts.WithGeneral(layouts.LayoutData{
		Title:   comm.Title,
		Content: template.HTML(content),
	})
	if err != nil {
		return nil, err
	}

	return handler.Ok(html), nil
}

func templatesExecute(name string, data any) ([]byte, error) {
	var buf bytes.Buffer
	templates := getTemplates()
	if err := templates.ExecuteTemplate(&buf, name, data); err != nil {
		return nil, errors.WithStack(err)
	}
	return buf.Bytes(), nil
}
