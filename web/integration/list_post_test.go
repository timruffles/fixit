package integration

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"fixit/engine/config"
	"fixit/engine/ent"
	"fixit/engine/factory"
)

func TestFrontPage(t *testing.T) {
	// Create test data using factory package
	dbClient := createDBClient(t)
	defer dbClient.Close()

	// Create test users and community with posts
	alice := factory.User(t, dbClient, "alice-*")
	bob := factory.User(t, dbClient, "bob-*")
	community := factory.Community(t, dbClient, "test-community-*")

	// Create posts for the community
	_, err := dbClient.Post.Create().
		SetTitle("Large pothole on Main Street").
		SetBody("There's a dangerous pothole that needs fixing").
		SetUser(alice).
		SetCommunity(community).
		Save(context.Background())
	require.NoError(t, err)

	_, err = dbClient.Post.Create().
		SetTitle("Graffiti on playground equipment").
		SetBody("Playground equipment needs cleaning").
		SetUser(bob).
		SetCommunity(community).
		Save(context.Background())
	require.NoError(t, err)

	// Test the front page shows communities
	resp, err := http.Get(testServer.URL + "/")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/html")

	body := readResponseBody(t, resp)

	// Check front page content
	assert.Contains(t, body, "FixIt - Community Issue Tracker")
	assert.Contains(t, body, "Active Communities")
	assert.Contains(t, body, community.Title)
	assert.Contains(t, body, "/c/"+community.Name)

	// Test the community page shows posts and usernames
	resp, err = http.Get(testServer.URL + "/c/" + community.Name)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	communityBody := readResponseBody(t, resp)

	assert.Contains(t, communityBody, "Large pothole on Main Street")
	assert.Contains(t, communityBody, "Graffiti on playground equipment")
	assert.Contains(t, communityBody, alice.Username)
	assert.Contains(t, communityBody, bob.Username)
}

func TestPostWithImages(t *testing.T) {
	// Create test data using factory package
	dbClient := createDBClient(t)
	defer dbClient.Close()

	// Create test users and community with posts
	alice := factory.User(t, dbClient, "alice-*")
	community := factory.Community(t, dbClient, "test-community-*")

	// Create post with image
	postWithImage, err := dbClient.Post.Create().
		SetTitle("Issue with image attachment").
		SetBody("This post has an image attached").
		SetImageURL("https://example.com/test-image.jpg").
		SetUser(alice).
		SetCommunity(community).
		Save(context.Background())
	require.NoError(t, err)

	// Test the community page shows polaroid image
	resp, err := http.Get(testServer.URL + "/c/" + community.Name)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	communityBody := readResponseBody(t, resp)

	// Check that the image and polaroid styles are present
	assert.Contains(t, communityBody, "polaroid-container")
	assert.Contains(t, communityBody, "polaroid-image")
	assert.Contains(t, communityBody, "https://example.com/test-image.jpg")

	// Test the individual post page shows large image
	resp, err = http.Get(testServer.URL + "/p/" + postWithImage.ID.String())
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	postBody := readResponseBody(t, resp)

	// Check that the large image is displayed
	assert.Contains(t, postBody, "https://example.com/test-image.jpg")
	assert.Contains(t, postBody, "max-h-96") // Large image styling
}

func createDBClient(t *testing.T) *ent.Client {
	client, err := ent.Open("postgres", config.GetTestDBURL())
	require.NoError(t, err)
	return client
}

