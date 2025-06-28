package list

import (
	"bytes"
	"context"
	"embed"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"fixit/engine/community"
	"fixit/engine/ent"
	"fixit/web/handler"
	"fixit/web/layouts"
)

//go:embed templates/*.gohtml
var templatesFS embed.FS
var templates = template.Must(template.ParseFS(templatesFS, "templates/*.gohtml"))

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
	router.HandleFunc("/", handler.Wrap(h.handleList)).Methods("GET")
}

func (h *Handler) handleList(req *http.Request) (*handler.Response, error) {
	ctx := context.Background()
	communityID := uuid.Must(uuid.NewV4())
	filter := &community.Filter{}

	postItems, err := h.repo.ListPosts(ctx, communityID, filter)
	if err != nil {
		return nil, err
	}

	slog.Info("renderiind posts", "len", len(postItems))
	data := struct {
		Title string
		Posts []community.PostListItem
	}{
		Title: "Community Issues",
		Posts: postItems,
	}

	content, err := templatesExecute("list.gohtml", data)
	if err != nil {
		return nil, err
	}

	html, err := layouts.WithGeneral(layouts.LayoutData{
		Title:   "Community Issues",
		Content: template.HTML(content),
	})
	if err != nil {
		return nil, err
	}

	return handler.Ok(html), nil
}

func templatesExecute(name string, data any) ([]byte, error) {
	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, name, data); err != nil {
		return nil, errors.WithStack(err)
	}
	return buf.Bytes(), nil
}
