package frontpage

import (
	"bytes"
	"context"
	_ "embed"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

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
}

type FrontpageData struct {
	AppName     string
	Communities []*ent.Community
}

func New(communityRepo *community.Repository) *Handler {
	return &Handler{
		communityRepo: communityRepo,
	}
}

func (h *Handler) FrontpageHandler(r *http.Request) (handler.Response, error) {
	ctx := context.Background()

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
