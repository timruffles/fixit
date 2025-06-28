package post_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/lib/pq"

	"fixit/engine/config"
	"fixit/engine/ent"
	"fixit/engine/ent/enttest"
	entPost "fixit/engine/ent/post"
	"fixit/engine/post"
)

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

func createTestUser(t *testing.T, client *ent.Client, username string) *ent.User {
	user, err := client.User.Create().
		SetUsername(username).
		SetEmail(username + "@example.com").
		SetPassword("password123").
		Save(context.Background())
	require.NoError(t, err)
	return user
}

func createTestCommunity(t *testing.T, client *ent.Client, name string) *ent.Community {
	community, err := client.Community.Create().
		SetName(name).
		SetTitle("Test Community " + name).
		Save(context.Background())
	require.NoError(t, err)
	return community
}

func TestRepository_CreatePostGraph(t *testing.T) {
	client := setupTestDB(t)
	ctx := context.Background()
	repo := post.New(client)

	// Create test users
	issueAuthor := createTestUser(t, client, fmt.Sprintf("issue_author_%d_%d", time.Now().UnixNano(), 1))
	solutionAuthor := createTestUser(t, client, fmt.Sprintf("solution_author_%d_%d", time.Now().UnixNano(), 2))
	verificationAuthor := createTestUser(t, client, fmt.Sprintf("verification_author_%d_%d", time.Now().UnixNano(), 3))

	// Create test community
	community := createTestCommunity(t, client, fmt.Sprintf("test_community_%d", time.Now().UnixNano()))

	// Step 1: Create an issue post
	issuePost, err := repo.Create(ctx, post.PostCreateFields{
		Title:       "Test Issue Post",
		Role:        entPost.RoleIssue,
		Tags:        []string{"issue", "test"},
		UserID:      issueAuthor.ID,
		CommunityID: community.ID,
	})
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
		UserID:      solutionAuthor.ID,
		CommunityID: community.ID,
	})
	require.NoError(t, err)
	assert.NotNil(t, solutionPost)
	assert.Equal(t, "Test Solution Post", solutionPost.Title)
	assert.Equal(t, entPost.RoleSolution, solutionPost.Role)
	assert.Equal(t, issuePost.ID, solutionPost.ReplyTo)

	// Step 3: Create a verification post replying to the solution
	verificationPost, err := repo.Create(ctx, post.PostCreateFields{
		Title:       "Test Verification Post",
		Role:        entPost.RoleVerification,
		ReplyTo:     &solutionPost.ID,
		UserID:      verificationAuthor.ID,
		CommunityID: community.ID,
	})
	require.NoError(t, err)
	assert.NotNil(t, verificationPost)
	assert.Equal(t, "Test Verification Post", verificationPost.Title)
	assert.Equal(t, entPost.RoleVerification, verificationPost.Role)
	assert.Equal(t, solutionPost.ID, verificationPost.ReplyTo)

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
	user := createTestUser(t, client, fmt.Sprintf("test_user_%d", time.Now().UnixNano()))
	community := createTestCommunity(t, client, fmt.Sprintf("test_community_%d", time.Now().UnixNano()))

	// Test: Solution post without reply_to
	_, err := repo.Create(ctx, post.PostCreateFields{
		Title:       "Invalid Solution",
		Role:        entPost.RoleSolution,
		UserID:      user.ID,
		CommunityID: community.ID,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "solution posts must reply to an existing post")

	// Test: Solution post replying to non-existent post
	nonExistentID := uuid.Must(uuid.NewV7())
	_, err = repo.Create(ctx, post.PostCreateFields{
		Title:       "Invalid Solution",
		Role:        entPost.RoleSolution,
		ReplyTo:     &nonExistentID,
		UserID:      user.ID,
		CommunityID: community.ID,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parent post not found")

	// Create an issue post without 'issue' tag
	nonIssuePost, err := repo.Create(ctx, post.PostCreateFields{
		Title:       "Non-Issue Post",
		Role:        entPost.RoleIssue,
		Tags:        []string{"discussion"},
		UserID:      user.ID,
		CommunityID: community.ID,
	})
	require.NoError(t, err)

	// Test: Solution post replying to post without 'issue' tag
	_, err = repo.Create(ctx, post.PostCreateFields{
		Title:       "Invalid Solution",
		Role:        entPost.RoleSolution,
		ReplyTo:     &nonIssuePost.ID,
		UserID:      user.ID,
		CommunityID: community.ID,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "solution posts can only reply to posts with 'issue' tag")

	// Create a proper issue post
	issuePost, err := repo.Create(ctx, post.PostCreateFields{
		Title:       "Valid Issue",
		Role:        entPost.RoleIssue,
		Tags:        []string{"issue"},
		UserID:      user.ID,
		CommunityID: community.ID,
	})
	require.NoError(t, err)

	// Create a solution post
	solutionPost, err := repo.Create(ctx, post.PostCreateFields{
		Title:       "Valid Solution",
		Role:        entPost.RoleSolution,
		ReplyTo:     &issuePost.ID,
		UserID:      user.ID,
		CommunityID: community.ID,
	})
	require.NoError(t, err)

	// Test: Solution post replying to another solution (not top-level)
	_, err = repo.Create(ctx, post.PostCreateFields{
		Title:       "Invalid Solution",
		Role:        entPost.RoleSolution,
		ReplyTo:     &solutionPost.ID,
		UserID:      user.ID,
		CommunityID: community.ID,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "solution posts can only reply to top-level posts")

	// Test: Verification post without reply_to
	_, err = repo.Create(ctx, post.PostCreateFields{
		Title:       "Invalid Verification",
		Role:        entPost.RoleVerification,
		UserID:      user.ID,
		CommunityID: community.ID,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "verification posts must reply to an existing post")

	// Test: Verification post replying to non-solution post
	_, err = repo.Create(ctx, post.PostCreateFields{
		Title:       "Invalid Verification",
		Role:        entPost.RoleVerification,
		ReplyTo:     &issuePost.ID,
		UserID:      user.ID,
		CommunityID: community.ID,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "verification posts can only reply to solution posts")
}

func TestRepository_SelfReplyValidation(t *testing.T) {
	client := setupTestDB(t)
	ctx := context.Background()
	repo := post.New(client)

	// Create test users and community
	user := createTestUser(t, client, fmt.Sprintf("test_user_%d", time.Now().UnixNano()))
	otherUser := createTestUser(t, client, fmt.Sprintf("other_user_%d", time.Now().UnixNano()))
	community := createTestCommunity(t, client, fmt.Sprintf("test_community_%d", time.Now().UnixNano()))

	// Create an issue post by user
	issuePost, err := repo.Create(ctx, post.PostCreateFields{
		Title:       "User's Issue",
		Role:        entPost.RoleIssue,
		Tags:        []string{"issue"},
		UserID:      user.ID,
		CommunityID: community.ID,
	})
	require.NoError(t, err)

	// Test: User cannot reply to their own issue with a solution
	_, err = repo.Create(ctx, post.PostCreateFields{
		Title:       "Self Solution",
		Role:        entPost.RoleSolution,
		ReplyTo:     &issuePost.ID,
		UserID:      user.ID, // Same user as issue author
		CommunityID: community.ID,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "users cannot reply to their own posts with solution role")

	// Test: Other user CAN reply with a solution
	solutionPost, err := repo.Create(ctx, post.PostCreateFields{
		Title:       "Other User Solution",
		Role:        entPost.RoleSolution,
		ReplyTo:     &issuePost.ID,
		UserID:      otherUser.ID, // Different user
		CommunityID: community.ID,
	})
	require.NoError(t, err)
	assert.NotNil(t, solutionPost)

	// Test: Solution author cannot verify their own solution
	_, err = repo.Create(ctx, post.PostCreateFields{
		Title:       "Self Verification",
		Role:        entPost.RoleVerification,
		ReplyTo:     &solutionPost.ID,
		UserID:      otherUser.ID, // Same user as solution author
		CommunityID: community.ID,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "users cannot reply to their own posts with verification role")

	// Test: Original issue author CAN verify the solution
	verificationPost, err := repo.Create(ctx, post.PostCreateFields{
		Title:       "Issue Author Verification",
		Role:        entPost.RoleVerification,
		ReplyTo:     &solutionPost.ID,
		UserID:      user.ID, // Different user (original issue author)
		CommunityID: community.ID,
	})
	require.NoError(t, err)
	assert.NotNil(t, verificationPost)
}