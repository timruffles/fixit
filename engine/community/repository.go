package community

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/pkg/errors"

	"fixit/engine/ent"
)

type Filter struct{}

type Repository struct {
	client *ent.Client
}

func NewRepository(client *ent.Client) *Repository {
	return &Repository{
		client: client,
	}
}

func (r *Repository) ListPosts(ctx context.Context, communityID uuid.UUID, filter *Filter) ([]*ent.Post, error) {
	posts, err := r.client.Post.Query().All(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return posts, nil
}

func (r *Repository) Seed(ctx context.Context) error {
	count, err := r.client.Post.Query().Count(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	
	if count > 0 {
		return nil
	}

	posts := []string{
		"Large pothole on Main Street near bus stop",
		"Graffiti on playground equipment at Central Park",
		"Fly-tipping behind grocery store on Oak Avenue",
		"Broken street lamp on Elm Street",
		"Abandoned shopping trolleys in River Park",
		"Deep potholes causing car damage on Bridge Road",
		"Graffiti tags on railway bridge",
		"Overflowing bins near school entrance",
	}

	for _, title := range posts {
		_, err := r.client.Post.Create().
			SetTitle(title).
			Save(ctx)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}
