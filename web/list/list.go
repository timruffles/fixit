package list

import (
	"embed"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
)

//go:embed templates/*.html
var templates embed.FS

type Handler struct {
	tmpl *template.Template
}

type Issue struct {
	ID          int
	Title       string
	Reporter    string
	Status      string
	Comments    int
	Category    string
	TimeAgo     string
	Priority    string
}

func New() *Handler {
	tmpl := template.Must(template.ParseFS(templates, "templates/*.html"))
	return &Handler{
		tmpl: tmpl,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/", h.handleList).Methods("GET")
}

func (h *Handler) handleList(w http.ResponseWriter, r *http.Request) {
	issues := []Issue{
		{ID: 1, Title: "Large pothole on Main Street near bus stop", Reporter: "concerned_resident", Status: "Open", Comments: 12, Category: "Roads", TimeAgo: "2 hours ago", Priority: "High"},
		{ID: 2, Title: "Graffiti on playground equipment at Central Park", Reporter: "park_visitor", Status: "In Progress", Comments: 5, Category: "Vandalism", TimeAgo: "4 hours ago", Priority: "Medium"},
		{ID: 3, Title: "Fly-tipping behind grocery store on Oak Avenue", Reporter: "shop_owner", Status: "Open", Comments: 8, Category: "Waste", TimeAgo: "6 hours ago", Priority: "High"},
		{ID: 4, Title: "Broken street lamp on Elm Street", Reporter: "night_walker", Status: "Assigned", Comments: 3, Category: "Lighting", TimeAgo: "1 day ago", Priority: "Medium"},
		{ID: 5, Title: "Abandoned shopping trolleys in River Park", Reporter: "jogger123", Status: "Open", Comments: 15, Category: "Litter", TimeAgo: "1 day ago", Priority: "Low"},
		{ID: 6, Title: "Deep potholes causing car damage on Bridge Road", Reporter: "daily_commuter", Status: "Open", Comments: 23, Category: "Roads", TimeAgo: "2 days ago", Priority: "Critical"},
		{ID: 7, Title: "Graffiti tags on railway bridge", Reporter: "train_enthusiast", Status: "Resolved", Comments: 2, Category: "Vandalism", TimeAgo: "3 days ago", Priority: "Low"},
		{ID: 8, Title: "Overflowing bins near school entrance", Reporter: "parent_council", Status: "In Progress", Comments: 7, Category: "Waste", TimeAgo: "4 days ago", Priority: "Medium"},
	}

	w.Header().Set("Content-Type", "text/html")
	if err := h.tmpl.ExecuteTemplate(w, "list.html", issues); err != nil {
		slog.Error("Failed to execute template", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
