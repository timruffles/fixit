package post

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/aarondl/authboss/v3"
	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"fixit/engine/auth"
	"fixit/engine/community"
	"fixit/engine/ent"
	"fixit/engine/ent/post"
	postEngine "fixit/engine/post"
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
	ReplyToID   string
	PostType    string
}

type ShowPostData struct {
	ID                  uuid.UUID
	Title               string
	Body                string
	User                *ent.User
	Community           *ent.Community
	CreatedAt           time.Time
	Tags                []string
	Role                string
	HasAcceptedSolution bool
	Solutions           []*PostReply
	ChatMessages        []*PostReply
}

type PostReply struct {
	ID                uuid.UUID
	Title             string
	User              *ent.User
	CreatedAt         time.Time
	Role              string
	IsAccepted        bool
	HasVerifications  bool
	VerificationCount int
}

func (h *Handler) CreatePostPostHandler(r *http.Request) (handler.Response, error) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return handler.BadInput([]byte("Failed to parse form")), nil
	}

	title := r.FormValue("title")
	body := r.FormValue("body")
	tags := r.FormValue("tags")
	communitySlug := r.FormValue("community")
	replyToID := r.FormValue("reply_to_id")
	postType := r.FormValue("post_type")

	data := CreatePostData{
		Title:       title,
		Body:        body,
		Tags:        tags,
		CommunityID: communitySlug,
		ReplyToID:   replyToID,
		PostType:    postType,
	}

	if title == "" {
		data.Error = "Title is required"
		content, err := renderCreatePost(data)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return handler.BadInput(content), nil
	}

	if communitySlug == "" {
		data.Error = "Community is required"
		content, err := renderCreatePost(data)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return handler.BadInput(content), nil
	}

	user, isAuthenticated := auth.RequireAuth(h.ab, r)
	if !isAuthenticated {
		// For now, just redirect to login without flash message
		// TODO: Implement flash message support
		return handler.RedirectTo("/auth/login"), nil
	}

	ctx := r.Context()

	// Get community by slug to get its ID
	comm, err := h.communityRepo.GetBySlug(ctx, communitySlug)
	if err != nil {
		data.Error = "Community not found"
		content, renderErr := renderCreatePost(data)
		if renderErr != nil {
			return nil, errors.WithStack(renderErr)
		}
		return handler.BadInput(content), nil
	}

	// Parse tags
	var tagsList []string
	if tags != "" {
		tagsList = strings.Split(tags, ",")
		for i, tag := range tagsList {
			tagsList[i] = strings.TrimSpace(tag)
		}
	}

	// Determine post role based on post_type
	var postRole post.Role
	switch postType {
	case "solution":
		postRole = post.RoleSolution
	case "verification":
		postRole = post.RoleVerification
	case "chat":
		postRole = post.RoleChat
	default:
		postRole = post.RoleIssue // Default to issue for new posts
	}

	// Parse reply_to_id if provided
	var replyToUUID *uuid.UUID
	if replyToID != "" {
		parsedUUID, err := uuid.FromString(replyToID)
		if err != nil {
			data.Error = "Invalid reply_to_id format"
			content, renderErr := renderCreatePost(data)
			if renderErr != nil {
				return nil, errors.WithStack(renderErr)
			}
			return handler.BadInput(content), nil
		}
		replyToUUID = &parsedUUID
	}

	// Create post using repository
	fields := postEngine.PostCreateFields{
		Title:       title,
		Body:        body,
		Role:        postRole,
		Tags:        tagsList,
		CommunityID: comm.ID,
		ReplyTo:     replyToUUID,
	}

	createdPost, err := h.postRepo.Create(ctx, fields, user.User)
	if err != nil {
		data.Error = "Failed to create post: " + err.Error()
		content, renderErr := renderCreatePost(data)
		if renderErr != nil {
			return nil, errors.WithStack(renderErr)
		}
		return handler.BadInput(content), nil
	}

	// Redirect to community page with post ID
	redirectURL := fmt.Sprintf("/c/%s?posted_id=%s", communitySlug, createdPost.ID.String())
	return handler.RedirectTo(redirectURL), nil
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

type Handler struct {
	postRepo      *postEngine.Repository
	communityRepo *community.Repository
	ab            *authboss.Authboss
}

func New(postRepo *postEngine.Repository, communityRepo *community.Repository, ab *authboss.Authboss) *Handler {
	return &Handler{
		postRepo:      postRepo,
		communityRepo: communityRepo,
		ab:            ab,
	}
}

func (h *Handler) CreatePostGetHandler(r *http.Request) (handler.Response, error) {
	vars := mux.Vars(r)
	communityID := vars["slug"]

	// Parse query parameters
	replyToID := r.URL.Query().Get("reply_to_id")
	postType := r.URL.Query().Get("post_type")

	data := CreatePostData{
		CommunityID: communityID,
		ReplyToID:   replyToID,
		PostType:    postType,
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

	postEntity, err := h.postRepo.GetByIDWithReplies(context.Background(), postID)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Process replies into solutions and chat messages
	var solutions []*PostReply
	var chatMessages []*PostReply
	hasAcceptedSolution := false

	for _, reply := range postEntity.Edges.Replies {
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
			replyData.IsAccepted = false       // TODO: implement acceptance logic
			replyData.HasVerifications = false // TODO: count verification replies
			replyData.VerificationCount = 0    // TODO: count verification replies
			if replyData.IsAccepted {
				hasAcceptedSolution = true
			}
			solutions = append(solutions, replyData)
		case "chat":
			chatMessages = append(chatMessages, replyData)
		}
	}

	data := ShowPostData{
		ID:                  postEntity.ID,
		Title:               postEntity.Title,
		Body:                postEntity.Body,
		User:                postEntity.Edges.User,
		Community:           postEntity.Edges.Community,
		CreatedAt:           postEntity.CreatedAt,
		Tags:                postEntity.Tags,
		Role:                string(postEntity.Role),
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
