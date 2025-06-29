package post

import (
	"context"

	"github.com/aarondl/authboss/v3"
	"github.com/gofrs/uuid/v5"
	"github.com/pkg/errors"

	"fixit/engine/ent"
	"fixit/engine/ent/post"
)

type Repository struct {
	client *ent.Client
	ab     *authboss.Authboss
}

type Auth interface {
}

func New(client *ent.Client) *Repository {
	return &Repository{
		client: client,
	}
}

type PostCreateFields struct {
	Title       string     `json:"title,omitempty"`
	Body        string     `json:"body,omitempty"`
	Role        post.Role  `json:"role,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	ReplyTo     *uuid.UUID `json:"replyTo,omitempty"`
	UserEmail   string     `json:"userEmail,omitempty"`
	CommunityID uuid.UUID  `json:"communityID,omitempty"`
}

func (r *Repository) Create(ctx context.Context, fields PostCreateFields, user *ent.User) (*ent.Post, error) {
	// Create fields with user ID for validation
	validationFields := PostCreateFields{
		Title:       fields.Title,
		Body:        fields.Body,
		Role:        fields.Role,
		Tags:        fields.Tags,
		ReplyTo:     fields.ReplyTo,
		UserEmail:   fields.UserEmail,
		CommunityID: fields.CommunityID,
	}

	// Validate role-specific requirements
	if err := r.validateRole(ctx, validationFields, user.ID); err != nil {
		return nil, errors.WithStack(err)
	}

	builder := r.client.Post.Create().
		SetTitle(fields.Title).
		SetRole(fields.Role).
		SetUserID(user.ID).
		SetCommunityID(fields.CommunityID)

	if fields.Body != "" {
		builder.SetBody(fields.Body)
	}

	if fields.Tags != nil {
		builder.SetTags(fields.Tags)
	}

	if fields.ReplyTo != nil {
		builder.SetReplyTo(*fields.ReplyTo)
	}

	post, err := builder.Save(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return post, nil
}

func (r *Repository) GetByIDWithReplies(ctx context.Context, id uuid.UUID) (*ent.Post, error) {
	post, err := r.client.Post.Query().
		Where(post.ID(id)).
		WithUser().
		WithCommunity().
		WithReplies(func(q *ent.PostQuery) {
			q.WithUser().Order(ent.Asc(post.FieldCreatedAt))
		}).
		Only(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return post, nil
}

func (r *Repository) validateRole(ctx context.Context, fields PostCreateFields, userID uuid.UUID) error {
	switch fields.Role {
	case post.RoleSolution:
		return r.validateSolutionRole(ctx, fields, userID)
	case post.RoleVerification:
		return r.validateVerificationRole(ctx, fields, userID)
	default:
		return nil
	}
}

func (r *Repository) validateSolutionRole(ctx context.Context, fields PostCreateFields, userID uuid.UUID) error {
	if fields.ReplyTo == nil {
		return errors.New("solution posts must reply to an existing post")
	}

	// Check that the parent post exists, is top-level, and has 'issue' role
	parentPost, err := r.client.Post.Query().
		Where(post.ID(*fields.ReplyTo)).
		WithUser().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return errors.New("parent post not found")
		}
		return errors.WithStack(err)
	}

	// Parent must be top-level (no reply_to)
	if parentPost.ReplyTo != nil {
		return errors.New("solution posts can only reply to top-level posts")
	}

	// Parent must have 'issue' role
	if parentPost.Role != post.RoleIssue {
		return errors.New("solution posts can only reply to posts with 'issue' role")
	}

	return nil
}

func (r *Repository) validateVerificationRole(ctx context.Context, fields PostCreateFields, userID uuid.UUID) error {
	if fields.ReplyTo == nil {
		return errors.New("verification posts must reply to an existing post")
	}

	// Check that the parent post exists and has PostRoleSolution role
	parentPost, err := r.client.Post.Query().
		Where(post.ID(*fields.ReplyTo)).
		WithUser().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return errors.New("parent post not found")
		}
		return errors.WithStack(err)
	}

	if parentPost.Role != post.RoleSolution {
		return errors.New("verification posts can only reply to solution posts")
	}

	// Check that the user is not replying to their own post
	if parentPost.Edges.User != nil && parentPost.Edges.User.ID == userID {
		return errors.New("users cannot reply to their own posts with verification role")
	}

	return nil
}
