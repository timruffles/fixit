package post_test

import (
	"context"
	"testing"

	"github.com/gofrs/uuid/v5"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"fixit/engine/config"
	"fixit/engine/ent"
	"fixit/engine/ent/enttest"
	entPost "fixit/engine/ent/post"
	"fixit/engine/factory"
	"fixit/engine/post"
)

func TestRepository_CreatePostGraph(t *testing.T) {
	client := setupTestDB(t)
	ctx := context.Background()
	repo := post.New(client)

	// Create test users
	issueAuthor := factory.User(t, client, "issue-author-*")
	solutionAuthor := factory.User(t, client, "solution-author-*")
	verificationAuthor := factory.User(t, client, "verification-author-*")

	// Create test community
	community := factory.Community(t, client, "test-community-*")

	// Step 1: Create an issue post
	issuePost, err := repo.Create(ctx, post.PostCreateFields{
		Title:       "Test Issue Post",
		Role:        entPost.RoleIssue,
		Tags:        []string{"issue", "test"},
		CommunityID: community.ID,
	}, issueAuthor)
	require.NoError(t, err)
	assert.NotNil(t, issuePost)
	assert.Equal(t, "Test Issue Post", issuePost.Title)
	assert.Equal(t, entPost.RoleIssue, issuePost.Role)
	assert.Contains(t, issuePost.Tags, "issue")

	// Step 2: Create a solution post replying to the issue
	solutionPost, err := repo.Create(ctx, post.PostCreateFields{
		Title:       "Test Solution Post",
		Role:        entPost.RoleSolution,
		ReplyTo:     &issuePost.ID,
		CommunityID: community.ID,
	}, solutionAuthor)
	require.NoError(t, err)
	assert.NotNil(t, solutionPost)
	assert.Equal(t, "Test Solution Post", solutionPost.Title)
	assert.Equal(t, entPost.RoleSolution, solutionPost.Role)
	assert.Equal(t, issuePost.ID, *solutionPost.ReplyTo)

	// Step 3: Create a verification post replying to the solution
	verificationPost, err := repo.Create(ctx, post.PostCreateFields{
		Title:       "Test Verification Post",
		Role:        entPost.RoleVerification,
		ReplyTo:     &solutionPost.ID,
		CommunityID: community.ID,
	}, verificationAuthor)
	require.NoError(t, err)
	assert.NotNil(t, verificationPost)
	assert.Equal(t, "Test Verification Post", verificationPost.Title)
	assert.Equal(t, entPost.RoleVerification, verificationPost.Role)
	assert.Equal(t, solutionPost.ID, *verificationPost.ReplyTo)

	// Verify the graph structure
	// Load posts with edges
	loadedIssue, err := client.Post.Query().
		Where(entPost.ID(issuePost.ID)).
		WithReplies().
		Only(ctx)
	require.NoError(t, err)
	assert.Len(t, loadedIssue.Edges.Replies, 1)
	assert.Equal(t, solutionPost.ID, loadedIssue.Edges.Replies[0].ID)

	loadedSolution, err := client.Post.Query().
		Where(entPost.ID(solutionPost.ID)).
		WithParent().
		WithReplies().
		Only(ctx)
	require.NoError(t, err)
	assert.Equal(t, issuePost.ID, loadedSolution.Edges.Parent.ID)
	assert.Len(t, loadedSolution.Edges.Replies, 1)
	assert.Equal(t, verificationPost.ID, loadedSolution.Edges.Replies[0].ID)
}

func TestRepository_ValidationErrors(t *testing.T) {
	client := setupTestDB(t)
	ctx := context.Background()
	repo := post.New(client)

	// Create test users and community
	user := factory.User(t, client, "test-user-*")
	otherUser := factory.User(t, client, "other-test-user-*")
	community := factory.Community(t, client, "test-community-*")

	// Test: Solution post without reply_to
	_, err := repo.Create(ctx, post.PostCreateFields{
		Title:       "Invalid Solution",
		Role:        entPost.RoleSolution,
		CommunityID: community.ID,
	}, user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "solution posts must reply to an existing post")

	// Test: Solution post replying to non-existent post
	nonExistentID := uuid.Must(uuid.NewV7())
	_, err = repo.Create(ctx, post.PostCreateFields{
		Title:       "Invalid Solution",
		Role:        entPost.RoleSolution,
		ReplyTo:     &nonExistentID,
		CommunityID: community.ID,
	}, user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parent post not found")

	// Create an issue post without 'issue' tag
	nonIssuePost, err := repo.Create(ctx, post.PostCreateFields{
		Title:       "Non-Issue Post",
		Role:        entPost.RoleIssue,
		Tags:        []string{"discussion"},
		CommunityID: community.ID,
	}, user)
	require.NoError(t, err)

	// Test: Solution post replying to post without 'issue' tag
	_, err = repo.Create(ctx, post.PostCreateFields{
		Title:       "Invalid Solution",
		Role:        entPost.RoleSolution,
		ReplyTo:     &nonIssuePost.ID,
		CommunityID: community.ID,
	}, user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "solution posts can only reply to posts with 'issue' tag")

	// Create a proper issue post
	issuePost, err := repo.Create(ctx, post.PostCreateFields{
		Title:       "Valid Issue",
		Role:        entPost.RoleIssue,
		Tags:        []string{"issue"},
		CommunityID: community.ID,
	}, user)
	require.NoError(t, err)

	// Create a solution post (must be different user)
	solutionPost, err := repo.Create(ctx, post.PostCreateFields{
		Title:       "Valid Solution",
		Role:        entPost.RoleSolution,
		ReplyTo:     &issuePost.ID,
		CommunityID: community.ID,
	}, otherUser)
	require.NoError(t, err)

	// Test: Solution post replying to another solution (not top-level)
	_, err = repo.Create(ctx, post.PostCreateFields{
		Title:       "Invalid Solution",
		Role:        entPost.RoleSolution,
		ReplyTo:     &solutionPost.ID,
		CommunityID: community.ID,
	}, user) // Can be same or different user, the issue is that it's replying to a non-top-level post
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "solution posts can only reply to top-level posts")

	// Test: Verification post without reply_to
	_, err = repo.Create(ctx, post.PostCreateFields{
		Title:       "Invalid Verification",
		Role:        entPost.RoleVerification,
		CommunityID: community.ID,
	}, user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "verification posts must reply to an existing post")

	// Test: Verification post replying to non-solution post
	_, err = repo.Create(ctx, post.PostCreateFields{
		Title:       "Invalid Verification",
		Role:        entPost.RoleVerification,
		ReplyTo:     &issuePost.ID,
		CommunityID: community.ID,
	}, user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "verification posts can only reply to solution posts")
}

func TestRepository_SelfReplyValidation(t *testing.T) {
	client := setupTestDB(t)
	ctx := context.Background()
	repo := post.New(client)

	// Create test users and community
	user := factory.User(t, client, "self-reply-user-*")
	otherUser := factory.User(t, client, "self-reply-other-*")
	community := factory.Community(t, client, "self-reply-community-*")

	// Create an issue post by user
	issuePost, err := repo.Create(ctx, post.PostCreateFields{
		Title:       "User's Issue",
		Role:        entPost.RoleIssue,
		Tags:        []string{"issue"},
		CommunityID: community.ID,
	}, user)
	require.NoError(t, err)

	// Test: User cannot reply to their own issue with a solution
	_, err = repo.Create(ctx, post.PostCreateFields{
		Title:       "Self Solution",
		Role:        entPost.RoleSolution,
		ReplyTo:     &issuePost.ID,
		CommunityID: community.ID,
	}, user) // Same user as issue author
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "users cannot reply to their own posts with solution role")

	// Test: Other user CAN reply with a solution
	solutionPost, err := repo.Create(ctx, post.PostCreateFields{
		Title:       "Other User Solution",
		Role:        entPost.RoleSolution,
		ReplyTo:     &issuePost.ID,
		CommunityID: community.ID,
	}, otherUser) // Different user
	require.NoError(t, err)
	assert.NotNil(t, solutionPost)

	// Test: Solution author cannot verify their own solution
	_, err = repo.Create(ctx, post.PostCreateFields{
		Title:       "Self Verification",
		Role:        entPost.RoleVerification,
		ReplyTo:     &solutionPost.ID,
		CommunityID: community.ID,
	}, otherUser) // Same user as solution author
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "users cannot reply to their own posts with verification role")

	// Test: Original issue author CAN verify the solution
	verificationPost, err := repo.Create(ctx, post.PostCreateFields{
		Title:       "Issue Author Verification",
		Role:        entPost.RoleVerification,
		ReplyTo:     &solutionPost.ID,
		CommunityID: community.ID,
	}, user) // Different user (original issue author)
	require.NoError(t, err)
	assert.NotNil(t, verificationPost)
}

func TestRepository_GetByIDWithReplies(t *testing.T) {
	client := setupTestDB(t)
	ctx := context.Background()
	repo := post.New(client)

	// Create test data
	user := factory.User(t, client, "get-by-id-user-*")
	community := factory.Community(t, client, "get-by-id-community-*")

	// Create main post
	mainPost, err := repo.Create(ctx, post.PostCreateFields{
		Title:       "Test Issue Post",
		Role:        entPost.RoleIssue,
		Tags:        []string{"issue", "test"},
		CommunityID: community.ID,
	}, user)
	require.NoError(t, err)

	// Create a reply
	otherUser := factory.User(t, client, "get-by-id-other-*")
	_, err = repo.Create(ctx, post.PostCreateFields{
		Title:       "Test Solution",
		Role:        entPost.RoleSolution,
		ReplyTo:     &mainPost.ID,
		CommunityID: community.ID,
	}, otherUser)
	require.NoError(t, err)

	// Test GetByIDWithReplies
	retrievedPost, err := repo.GetByIDWithReplies(ctx, mainPost.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrievedPost)
	assert.Equal(t, mainPost.ID, retrievedPost.ID)
	assert.Equal(t, "Test Issue Post", retrievedPost.Title)

	// Check that edges are loaded
	assert.NotNil(t, retrievedPost.Edges.User)
	assert.Equal(t, user.ID, retrievedPost.Edges.User.ID)
	assert.NotNil(t, retrievedPost.Edges.Community)
	assert.Equal(t, community.ID, retrievedPost.Edges.Community.ID)

	// Check replies are loaded
	assert.Len(t, retrievedPost.Edges.Replies, 1)
	assert.Equal(t, "Test Solution", retrievedPost.Edges.Replies[0].Title)
	assert.NotNil(t, retrievedPost.Edges.Replies[0].Edges.User)
}

func setupTestDB(t *testing.T) *ent.Client {
	// Use enttest with migrations - it handles cleaning up for us
	// The WithMigrateOptions ensures we get a fresh schema each time
	opts := []enttest.Option{
		enttest.WithOptions(ent.Log(t.Log)),
		enttest.WithMigrateOptions(),
	}

	client := enttest.Open(t, "postgres", config.TestDBURL, opts...)

	return client
}
