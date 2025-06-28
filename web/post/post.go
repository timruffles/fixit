package post

import (
	"bytes"
	"context"
	_ "embed"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"fixit/engine/ent"
	"fixit/engine/post"
	"fixit/web/handler"
	"fixit/web/layouts"
)

//go:embed templates/create.gohtml
var createTplS string

//go:embed templates/show.gohtml
var showTplS string

//go:embed templates/community_header.gohtml
var communityHeaderTplS string

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
}

var createTpl = template.Must(template.New("create").Funcs(templateFuncs).Parse(createTplS))
var showTpl = template.Must(template.New("show").Funcs(templateFuncs).Parse(showTplS))

func init() {
	// Parse community_header as a template that can be included
	template.Must(showTpl.New("community_header").Funcs(templateFuncs).Parse(communityHeaderTplS))
}

type CreatePostData struct {
	Title       string
	Body        string
	Tags        string
	Error       string
	CommunityID string
}

type ShowPostData struct {
	ID                   uuid.UUID
	Title                string
	User                 *ent.User
	Community            *ent.Community
	CreatedAt            time.Time
	Tags                 []string
	Role                 string
	HasAcceptedSolution  bool
	Solutions            []*PostReply
	ChatMessages         []*PostReply
}

type PostReply struct {
	ID                  uuid.UUID
	Title               string
	User                *ent.User
	CreatedAt           time.Time
	Role                string
	IsAccepted          bool
	HasVerifications    bool
	VerificationCount   int
}

func (h *Handler) CreatePostPostHandler(r *http.Request) (handler.Response, error) {
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

type Handler struct{
	postRepo *post.Repository
}

func New(postRepo *post.Repository) *Handler {
	return &Handler{
		postRepo: postRepo,
	}
}

func (h *Handler) CreatePostGetHandler(r *http.Request) (handler.Response, error) {
	vars := mux.Vars(r)
	communityID := vars["slug"]

	data := CreatePostData{
		CommunityID: communityID,
	}

	content, err := renderCreatePost(data)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return handler.Ok(content), nil
}

func (h *Handler) ShowPostHandler(r *http.Request) (handler.Response, error) {
	vars := mux.Vars(r)
	postIDStr := vars["id"]
	
	postID, err := uuid.FromString(postIDStr)
	if err != nil {
		return handler.BadInput([]byte("Invalid post ID")), nil
	}

	post, err := h.postRepo.GetByIDWithReplies(context.Background(), postID)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Process replies into solutions and chat messages
	var solutions []*PostReply
	var chatMessages []*PostReply
	hasAcceptedSolution := false

	for _, reply := range post.Edges.Replies {
		replyData := &PostReply{
			ID:        reply.ID,
			Title:     reply.Title,
			User:      reply.Edges.User,
			CreatedAt: reply.CreatedAt,
			Role:      string(reply.Role),
		}

		switch reply.Role {
		case "solution":
			// TODO: Check if this solution is accepted and count verifications
			replyData.IsAccepted = false // TODO: implement acceptance logic
			replyData.HasVerifications = false // TODO: count verification replies
			replyData.VerificationCount = 0 // TODO: count verification replies
			if replyData.IsAccepted {
				hasAcceptedSolution = true
			}
			solutions = append(solutions, replyData)
		case "chat":
			chatMessages = append(chatMessages, replyData)
		}
	}

	data := ShowPostData{
		ID:                  post.ID,
		Title:               post.Title,
		User:                post.Edges.User,
		Community:           post.Edges.Community,
		CreatedAt:           post.CreatedAt,
		Tags:                post.Tags,
		Role:                string(post.Role),
		HasAcceptedSolution: hasAcceptedSolution,
		Solutions:           solutions,
		ChatMessages:        chatMessages,
	}

	content, err := renderShowPost(data)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return handler.Ok(content), nil
}

func renderShowPost(data ShowPostData) ([]byte, error) {
	var content bytes.Buffer
	err := showTpl.Execute(&content, data)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	layoutData := layouts.LayoutData{
		Title:   data.Title,
		Content: template.HTML(content.String()),
	}

	return layouts.WithGeneral(layoutData)
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/c/{slug}/post", handler.Wrap(h.CreatePostGetHandler)).Methods("GET")
	router.HandleFunc("/api/post/create", handler.Wrap(h.CreatePostPostHandler)).Methods("POST")
	router.HandleFunc("/p/{id}", handler.Wrap(h.ShowPostHandler)).Methods("GET")
}
