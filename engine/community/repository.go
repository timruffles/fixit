package community

import (
	"context"

	"github.com/pkg/errors"

	"fixit/engine/ent"
	"fixit/engine/ent/community"
	"fixit/engine/ent/post"
)

type CommunityCreateFields struct {
	Name           string `json:"name,omitempty"`
	Title          string `json:"title,omitempty"`
	Location       string `json:"location,omitempty"`
	BannerImageURL string `json:"bannerImageURL,omitempty"`
	Geography      string `json:"geography,omitempty"`
}

type Filter struct {
	Location string
}

type PostListItem struct {
	*ent.Post
	Username string
	Solved   bool
}

type Repository struct {
	client *ent.Client
}

func NewRepository(client *ent.Client) *Repository {
	return &Repository{
		client: client,
	}
}

func (r *Repository) Create(ctx context.Context, fields CommunityCreateFields) (*ent.Community, error) {
	builder := r.client.Community.Create().
		SetName(fields.Name).
		SetTitle(fields.Title)

	if fields.Location != "" {
		builder.SetLocation(fields.Location)
	}

	if fields.BannerImageURL != "" {
		builder.SetBannerImageURL(fields.BannerImageURL)
	}

	if fields.Geography != "" {
		builder.SetGeography(fields.Geography)
	}

	community, err := builder.Save(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return community, nil
}

func (r *Repository) GetBySlug(ctx context.Context, slug string) (*ent.Community, error) {
	comm, err := r.client.Community.Query().
		Where(community.NameEQ(slug)).
		Only(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return comm, nil
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
	for _, p := range posts {
		// Check if p has solution replies to determine if it's solved
		solved := false
		if p.Role == post.RoleIssue {
			solutionCount, err := r.client.Post.Query().
				Where(
					post.RoleEQ(post.RoleSolution),
					post.ReplyToEQ(p.ID),
				).
				Count(ctx)
			if err == nil && solutionCount > 0 {
				solved = true
			}
		}

		item := PostListItem{
			Post:     p,
			Username: p.Edges.User.Username,
			Solved:   solved,
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
