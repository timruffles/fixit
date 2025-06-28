package post

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/pkg/errors"

	"fixit/engine/ent"
	"fixit/engine/ent/post"
)

type Repository struct {
	client *ent.Client
}

func New(client *ent.Client) *Repository {
	return &Repository{
		client: client,
	}
}

type PostCreateFields struct {
	Title       string     `json:"title,omitempty"`
	Role        post.Role  `json:"role,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	ReplyTo     *uuid.UUID `json:"replyTo,omitempty"`
	UserID      uuid.UUID  `json:"userID,omitempty"`
	CommunityID uuid.UUID  `json:"communityID,omitempty"`
}

func (r *Repository) Create(ctx context.Context, fields PostCreateFields) (*ent.Post, error) {
	// Validate role-specific requirements
	if err := r.validateRole(ctx, fields); err != nil {
		return nil, errors.WithStack(err)
	}

	builder := r.client.Post.Create().
		SetTitle(fields.Title).
		SetRole(fields.Role).
		SetUserID(fields.UserID).
		SetCommunityID(fields.CommunityID)

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

func (r *Repository) validateRole(ctx context.Context, fields PostCreateFields) error {
	switch fields.Role {
	case post.RoleSolution:
		return r.validateSolutionRole(ctx, fields)
	case post.RoleVerification:
		return r.validateVerificationRole(ctx, fields)
	default:
		return nil
	}
}

func (r *Repository) validateSolutionRole(ctx context.Context, fields PostCreateFields) error {
	if fields.ReplyTo == nil {
		return errors.New("solution posts must reply to an existing post")
	}

	// Check that the parent post exists, is top-level, and has 'issue' tag
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

	// Parent must have 'issue' tag
	hasIssueTag := false
	for _, tag := range parentPost.Tags {
		if tag == "issue" {
			hasIssueTag = true
			break
		}
	}
	if !hasIssueTag {
		return errors.New("solution posts can only reply to posts with 'issue' tag")
	}

	// Check that the user is not replying to their own post
	if parentPost.Edges.User != nil && parentPost.Edges.User.ID == fields.UserID {
		return errors.New("users cannot reply to their own posts with solution role")
	}

	return nil
}

func (r *Repository) validateVerificationRole(ctx context.Context, fields PostCreateFields) error {
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
	if parentPost.Edges.User != nil && parentPost.Edges.User.ID == fields.UserID {
		return errors.New("users cannot reply to their own posts with verification role")
	}

	return nil
}
