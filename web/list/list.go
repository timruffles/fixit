package list

import (
	"context"
	"embed"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/mux"

	"fixit/engine/community"
	"fixit/engine/ent"
)

//go:embed templates/*.html
var templates embed.FS

type Handler struct {
	tmpl *template.Template
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
	tmpl := template.Must(template.ParseFS(templates, "templates/*.html"))
	repo := community.NewRepository(client)
	return &Handler{
		tmpl: tmpl,
		repo: repo,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/", h.handleList).Methods("GET")
}

func (h *Handler) handleList(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	communityID := uuid.Must(uuid.NewV4())
	filter := &community.Filter{}

	entPosts, err := h.repo.ListPosts(ctx, communityID, filter)
	if err != nil {
		slog.Error("Failed to fetch posts", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	listTplParams := struct {
		Posts []*ent.Post
	}{
		Posts: entPosts,
	}
	if err := h.tmpl.ExecuteTemplate(w, "list.html", listTplParams); err != nil {
		slog.Error("Failed to execute template", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
