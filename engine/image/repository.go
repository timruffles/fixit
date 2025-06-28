package image

import (
	"context"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/pkg/errors"

	"github.com/timraymond/fixit/engine/ent"
	"github.com/timraymond/fixit/engine/ent/attachment"
	"github.com/timraymond/fixit/engine/ent/file"
	"github.com/timraymond/fixit/engine/ent/post"
)

type Repository struct {
	client *ent.Client
}

func NewRepository(client *ent.Client) *Repository {
	return &Repository{
		client: client,
	}
}

func (r *Repository) CreateFile(ctx context.Context, filename string, data []byte) (*ent.File, error) {
	ext := extractExtension(filename)
	
	file, err := r.client.File.Create().
		SetFilename(filename).
		SetExtension(ext).
		SetData(data).
		Save(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	
	return file, nil
}

func (r *Repository) AttachToPost(ctx context.Context, postID, fileID uuid.UUID, caption string) (*ent.Attachment, error) {
	builder := r.client.Attachment.Create().
		SetPostID(postID).
		SetFileID(fileID)
	
	if caption != "" {
		builder.SetCaption(caption)
	}
	
	attachment, err := builder.Save(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	
	return attachment, nil
}

func (r *Repository) GetPostAttachments(ctx context.Context, postID uuid.UUID) ([]*ent.Attachment, error) {
	attachments, err := r.client.Attachment.Query().
		Where(attachment.HasPostWith(post.ID(postID))).
		WithFile().
		All(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	
	return attachments, nil
}

func (r *Repository) GetFile(ctx context.Context, fileID uuid.UUID) (*ent.File, error) {
	file, err := r.client.File.Query().
		Where(file.ID(fileID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.WithStack(err)
		}
		return nil, errors.WithStack(err)
	}
	
	return file, nil
}

func (r *Repository) DeleteAttachment(ctx context.Context, attachmentID uuid.UUID) error {
	err := r.client.Attachment.DeleteOneID(attachmentID).Exec(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	
	return nil
}

func (r *Repository) DeleteFile(ctx context.Context, fileID uuid.UUID) error {
	exists, err := r.client.Attachment.Query().
		Where(attachment.HasFileWith(file.ID(fileID))).
		Exist(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	
	if exists {
		return errors.New("cannot delete file: still has attachments")
	}
	
	err = r.client.File.DeleteOneID(fileID).Exec(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	
	return nil
}

func extractExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) > 1 {
		return strings.ToLower(parts[len(parts)-1])
	}
	return ""
}