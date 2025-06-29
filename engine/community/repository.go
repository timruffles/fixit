package community

import (
	"context"

	"github.com/pkg/errors"

	"fixit/engine/ent"
	"fixit/engine/ent/community"
	"fixit/engine/ent/post"
)

type Filter struct {
	Location string
}

type PostListItem struct {
	*ent.Post
	Username string
}

type Repository struct {
	client *ent.Client
}

func NewRepository(client *ent.Client) *Repository {
	return &Repository{
		client: client,
	}
}

func (r *Repository) ListPosts(ctx context.Context, communitySlug string, filter *Filter) ([]PostListItem, error) {
	// First find the community by slug
	comm, err := r.client.Community.Query().
		Where(community.NameEQ(communitySlug)).
		Only(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Then query posts for that community
	posts, err := r.client.Post.Query().
		Where(post.HasCommunityWith(community.ID(comm.ID))).
		WithUser().
		All(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var items []PostListItem
	for _, post := range posts {
		item := PostListItem{
			Post:     post,
			Username: post.Edges.User.Username,
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *Repository) Seed(ctx context.Context) error {
	count, err := r.client.Post.Query().Count(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	if count > 0 {
		return nil
	}

	comm, err := r.client.Community.Create().
		SetName("swindon").
		SetTitle("Swindon Community").
		SetLocation("Swindon, UK").
		Save(ctx)
	if err != nil {
		return err
	}

	// Create example users
	users := []struct {
		username string
		email    string
		password string
	}{
		{"alice", "alice@example.com", "password123"},
		{"bob", "bob@example.com", "password123"},
		{"charlie", "charlie@example.com", "password123"},
		{"diana", "diana@example.com", "password123"},
	}

	var createdUsers []*ent.User
	for _, userData := range users {
		user, err := r.client.User.Create().
			SetUsername(userData.username).
			SetEmail(userData.email).
			SetPassword(userData.password).
			Save(ctx)
		if err != nil {
			return errors.WithStack(err)
		}
		createdUsers = append(createdUsers, user)
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

	for i, title := range posts {
		userIndex := i % len(createdUsers)
		_, err := r.client.Post.Create().
			SetTitle(title).
			SetCommunityID(comm.ID).
			SetUser(createdUsers[userIndex]).
			Save(ctx)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (r *Repository) ForFrontpage(ctx context.Context, filter Filter) ([]*ent.Community, error) {
	query := r.client.Community.Query()

	if filter.Location != "" {
		query = query.Where(community.LocationEQ(filter.Location))
	}

	communities, err := query.All(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return communities, nil
}
