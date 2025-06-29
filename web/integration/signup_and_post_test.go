package integration

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignupAndCreatePost(t *testing.T) {
	// Create a client with cookie jar to maintain session across requests
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	
	client := &http.Client{
		// Don't follow redirects automatically so we can check them
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Jar: jar,
	}

	// Generate unique credentials for this test
	timestamp := time.Now().Unix()
	username := "testuser" + strconv.FormatInt(timestamp, 10)
	email := username + "@test.com"
	password := "ValidPassword123!"

	// Step 1: Register a new user
	registerData := url.Values{
		"username":         {username},
		"email":            {email},
		"password":         {password},
		"confirm_password": {password},
	}

	registerResp, err := client.PostForm(testServer.URL+"/auth/register", registerData)
	require.NoError(t, err)
	defer registerResp.Body.Close()

	// Should redirect to home page after successful registration
	assert.Equal(t, http.StatusFound, registerResp.StatusCode)
	location := registerResp.Header.Get("Location")
	assert.Equal(t, "/", location)

	// Step 2: Create a community first (we need a community to post to)
	communityName := "test-community-" + strconv.FormatInt(timestamp, 10)
	communityFields := map[string]string{
		"name":     communityName,
		"title":    "Test Community " + strconv.FormatInt(timestamp, 10),
		"location": "Test Location",
	}

	communityResp, err := postMultipartForm(client, testServer.URL+"/api/community/create", communityFields)
	require.NoError(t, err)
	defer communityResp.Body.Close()

	// Should redirect to the new community page
	assert.Equal(t, http.StatusFound, communityResp.StatusCode)
	communityLocation := communityResp.Header.Get("Location")
	expectedCommunityURL := "/c/" + communityName
	assert.Equal(t, expectedCommunityURL, communityLocation)

	// Step 3: Create a post in the community
	postTitle := "My Test Post " + strconv.FormatInt(timestamp, 10)
	postFields := map[string]string{
		"title":     postTitle,
		"body":      "This is a test post body created by integration test",
		"tags":      "test,integration",
		"community": communityName,
	}

	postResp, err := postMultipartForm(client, testServer.URL+"/api/post/create", postFields)
	require.NoError(t, err)
	defer postResp.Body.Close()

	// Debug: check what error we're getting
	if postResp.StatusCode != http.StatusFound {
		body := readResponseBody(t, postResp)
		t.Logf("Post creation failed with status %d, body: %s", postResp.StatusCode, body)
	}

	// Should redirect to community page with post ID
	assert.Equal(t, http.StatusFound, postResp.StatusCode)
	postLocation := postResp.Header.Get("Location")
	
	// Check that it redirects to the community page
	assert.True(t, strings.HasPrefix(postLocation, "/c/"+communityName+"?posted_id="))
	
	// Extract the post ID from the redirect URL
	parts := strings.Split(postLocation, "posted_id=")
	require.Len(t, parts, 2)
	postID := parts[1]
	assert.NotEmpty(t, postID)

	// Step 4: Verify the post appears on the community page
	communityPageResp, err := client.Get(testServer.URL + "/c/" + communityName)
	require.NoError(t, err)
	defer communityPageResp.Body.Close()

	assert.Equal(t, http.StatusOK, communityPageResp.StatusCode)
	
	body := readResponseBody(t, communityPageResp)
	
	// Check that our post title appears in the community page
	assert.Contains(t, body, postTitle)
	assert.Contains(t, body, username) // The post should show the username
	
	// Check that humanized time is being used in the list page
	hasHumanizedTime := strings.Contains(body, " ago") || 
		strings.Contains(body, "now") || 
		strings.Contains(body, "seconds") ||
		strings.Contains(body, "second ago")
	assert.True(t, hasHumanizedTime, "List page should contain humanized time format")
	
	// Should NOT contain hardcoded "6 hours ago"
	assert.False(t, strings.Contains(body, "6 hours ago"), "Should not contain hardcoded time")

	// Step 5: Verify we can view the individual post page
	postPageResp, err := client.Get(testServer.URL + "/p/" + postID)
	require.NoError(t, err)
	defer postPageResp.Body.Close()

	assert.Equal(t, http.StatusOK, postPageResp.StatusCode)
	
	postPageBody := readResponseBody(t, postPageResp)
	
	// Check that the post details appear on the post page
	assert.Contains(t, postPageBody, postTitle)
	assert.Contains(t, postPageBody, "This is a test post body created by integration test")
	assert.Contains(t, postPageBody, username)
}

func TestCreatePostRequiresAuth(t *testing.T) {
	client := &http.Client{}

	// Try to create a post without being logged in
	postFields := map[string]string{
		"title":     "Unauthorized Post",
		"body":      "This should fail",
		"community": "swindon", // Use existing seeded community
	}

	// Configure client to NOT follow redirects so we can check the redirect response
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	resp, err := postMultipartForm(client, testServer.URL+"/api/post/create", postFields)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should return 302 Found redirecting to login
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	
	// Check that it redirects to login page
	location := resp.Header.Get("Location")
	assert.Equal(t, "/auth/login", location)
}

// postMultipartForm sends a multipart form request
func postMultipartForm(client *http.Client, url string, fields map[string]string) (*http.Response, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	
	for field, value := range fields {
		fw, err := w.CreateFormField(field)
		if err != nil {
			return nil, err
		}
		_, err = fw.Write([]byte(value))
		if err != nil {
			return nil, err
		}
	}
	
	w.Close()
	
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	
	return client.Do(req)
}