package community

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"

	"fixit/engine/community"
	handler "fixit/web/handler"
	"fixit/web/layouts"
)

//go:embed templates/create.gohtml
var createTplS string

var createTpl = template.Must(template.New("create").Parse(createTplS))

type CreateData struct {
	Name           string
	Title          string
	Location       string
	BannerImageURL string
	Latitude       string
	Longitude      string
	Error          string
}

type Handler struct {
	store *sessions.CookieStore
	repo  *community.Repository
}

func New(sessionKey []byte, repo *community.Repository) *Handler {
	return &Handler{
		store: sessions.NewCookieStore(sessionKey),
		repo:  repo,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/community/new", handler.Wrap(h.CreateGetHandler)).Methods("GET")
	router.HandleFunc("/api/community/create", handler.Wrap(h.CreatePostHandler)).Methods("POST")
}

func (h *Handler) CreateGetHandler(r *http.Request) (handler.Response, error) {
	session, err := h.store.Get(r, "fixit_session")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	data := CreateData{}

	// Check for flash data
	if flashes := session.Flashes("form_data"); len(flashes) > 0 {
		if jsonData, ok := flashes[0].(string); ok {
			if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
				// Log error but continue with empty form
			}
		}
	}

	// Check for error message
	if flashes := session.Flashes("error"); len(flashes) > 0 {
		if errMsg, ok := flashes[0].(string); ok {
			data.Error = errMsg
		}
	}

	// Don't save session here - let the handler wrapper deal with it
	// The flashes will be cleared automatically after being read

	return showCreateForm(data)
}

func (h *Handler) CreatePostHandler(r *http.Request) (handler.Response, error) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return handler.BadInput([]byte("Failed to parse form")), nil
	}

	name := r.FormValue("name")
	title := r.FormValue("title")
	location := r.FormValue("location")
	bannerImageURL := r.FormValue("banner_image_url")
	latitude := r.FormValue("latitude")
	longitude := r.FormValue("longitude")

	data := CreateData{
		Name:           name,
		Title:          title,
		Location:       location,
		BannerImageURL: bannerImageURL,
		Latitude:       latitude,
		Longitude:      longitude,
	}

	// Validation
	var validationError string
	if name == "" {
		validationError = "Community name is required"
	} else if title == "" {
		validationError = "Community title is required"
	}

	if validationError != "" {
		// Store form data and error in session
		session, err := h.store.Get(r, "fixit_session")
		if err != nil {
			return nil, errors.WithStack(err)
		}

		// Store form data as JSON
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		session.AddFlash(string(jsonData), "form_data")
		session.AddFlash(validationError, "error")

		return handler.RedirectTo("/community/new"), nil
	}

	// Create geography string from lat/lng
	var geography string
	if latitude != "" && longitude != "" {
		geography = fmt.Sprintf("POINT(%s %s)", longitude, latitude)
	}

	// Create community using repository
	ctx := context.Background()
	fields := community.CommunityCreateFields{
		Name:           name,
		Title:          title,
		Location:       location,
		BannerImageURL: bannerImageURL,
		Geography:      geography,
	}

	comm, err := h.repo.Create(ctx, fields)
	if err != nil {
		// Store form data and error in session
		session, sessionErr := h.store.Get(r, "fixit_session")
		if sessionErr != nil {
			return nil, errors.WithStack(sessionErr)
		}

		jsonData, jsonErr := json.Marshal(data)
		if jsonErr != nil {
			return nil, errors.WithStack(jsonErr)
		}
		session.AddFlash(string(jsonData), "form_data")
		session.AddFlash("Failed to create community: "+err.Error(), "error")

		return handler.RedirectTo("/community/new"), nil
	}

	// Success - redirect to the community page
	return handler.RedirectTo("/c/" + comm.Name), nil
}

func showCreateForm(data CreateData) (handler.Response, error) {
	var contentBuf bytes.Buffer
	err := createTpl.Execute(&contentBuf, data)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	layoutData := layouts.LayoutData{
		Title:   "Create Community",
		Content: template.HTML(contentBuf.String()),
	}

	content, err := layouts.WithGeneral(layoutData)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return handler.Ok(content), nil
}
