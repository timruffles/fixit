package frontpage

import (
	"bytes"
	"context"
	_ "embed"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/aarondl/authboss/v3"
	"github.com/dustin/go-humanize"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"fixit/engine/auth"
	"fixit/engine/community"
	"fixit/engine/config"
	"fixit/engine/ent"
	"fixit/web/handler"
	"fixit/web/layouts"
)

//go:embed templates/frontpage.gohtml
var frontpageTplS string

var templateFuncs = template.FuncMap{
	"upper": strings.ToUpper,
	"slice": func(s string, start, end int) string {
		if start >= len(s) {
			return ""
		}
		if end > len(s) {
			end = len(s)
		}
		return s[start:end]
	},
	"humanizeTime": func(t time.Time) string {
		return humanize.Time(t)
	},
}

var frontpageTpl = template.Must(template.New("frontpage").Funcs(templateFuncs).Parse(frontpageTplS))

type Handler struct {
	communityRepo *community.Repository
	ab            *authboss.Authboss
}

type FrontpageData struct {
	AppName       string
	Communities   []*ent.Community
	IsLoggedIn    bool
	Username      string
}

func New(communityRepo *community.Repository, ab *authboss.Authboss) *Handler {
	return &Handler{
		communityRepo: communityRepo,
		ab:            ab,
	}
}

func (h *Handler) FrontpageHandler(r *http.Request) (handler.Response, error) {
	ctx := context.Background()

	// Check if user is authenticated
	user, isLoggedIn := auth.RequireAuth(h.ab, r)
	var username string
	if isLoggedIn {
		username = user.GetUsername()
	}

	// Get location from query parameter if provided
	location := r.URL.Query().Get("location")
	filter := community.Filter{
		Location: location,
	}

	communities, err := h.communityRepo.ForFrontpage(ctx, filter)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	data := FrontpageData{
		AppName:     config.AppName,
		Communities: communities,
		IsLoggedIn:  isLoggedIn,
		Username:    username,
	}

	content, err := renderFrontpage(data)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return handler.Ok(content), nil
}

func renderFrontpage(data FrontpageData) ([]byte, error) {
	var content bytes.Buffer
	err := frontpageTpl.Execute(&content, data)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	layoutData := layouts.LayoutData{
		Title:   data.AppName + " - Community Issue Tracker",
		Content: template.HTML(content.String()),
	}

	return layouts.WithGeneral(layoutData)
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/", handler.Wrap(h.FrontpageHandler)).Methods("GET")
}
