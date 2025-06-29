package integration

import (
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

func TestPostReplyFunctionality(t *testing.T) {
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

	// Step 2: Create a community first
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

	// Step 3: Create an issue post
	issueTitle := "My Test Issue " + strconv.FormatInt(timestamp, 10)
	issueFields := map[string]string{
		"title":     issueTitle,
		"body":      "This is a test issue that needs a solution",
		"tags":      "issue,test",
		"community": communityName,
	}

	issueResp, err := postMultipartForm(client, testServer.URL+"/api/post/create", issueFields)
	require.NoError(t, err)
	defer issueResp.Body.Close()

	// Should redirect to community page with post ID
	assert.Equal(t, http.StatusFound, issueResp.StatusCode)
	issueLocation := issueResp.Header.Get("Location")
	
	// Extract the issue post ID from the redirect URL
	parts := strings.Split(issueLocation, "posted_id=")
	require.Len(t, parts, 2)
	issuePostID := parts[1]
	assert.NotEmpty(t, issuePostID)

	// Step 4: Create a solution post as a reply to the issue (users can now solve their own issues)
	solutionTitle := "Solution for " + issueTitle
	solutionFields := map[string]string{
		"title":       solutionTitle,
		"body":        "This is a solution to the issue",
		"tags":        "solution,test",
		"community":   communityName,
		"reply_to_id": issuePostID,
		"post_type":   "solution",
	}

	solutionResp, err := postMultipartForm(client, testServer.URL+"/api/post/create", solutionFields)
	require.NoError(t, err)
	defer solutionResp.Body.Close()

	// Should redirect to community page with solution post ID
	assert.Equal(t, http.StatusFound, solutionResp.StatusCode)
	solutionLocation := solutionResp.Header.Get("Location")
	
	// Extract the solution post ID from the redirect URL
	solutionParts := strings.Split(solutionLocation, "posted_id=")
	require.Len(t, solutionParts, 2)
	solutionPostID := solutionParts[1]
	assert.NotEmpty(t, solutionPostID)

	// Step 5: Create a verification post as a reply to the solution using a different user (users can't verify their own solutions)
	username3 := "verifier" + strconv.FormatInt(timestamp, 10)
	email3 := username3 + "@test.com"
	
	registerData3 := url.Values{
		"username":         {username3},
		"email":            {email3},
		"password":         {password},
		"confirm_password": {password},
	}

	// Create third client for the verifier user
	jar3, err := cookiejar.New(nil)
	require.NoError(t, err)
	
	client3 := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Jar: jar3,
	}

	registerResp3, err := client3.PostForm(testServer.URL+"/auth/register", registerData3)
	require.NoError(t, err)
	defer registerResp3.Body.Close()

	assert.Equal(t, http.StatusFound, registerResp3.StatusCode)

	verificationTitle := "Verification for solution"
	verificationFields := map[string]string{
		"title":       verificationTitle,
		"body":        "This solution works great!",
		"tags":        "verification,test",
		"community":   communityName,
		"reply_to_id": solutionPostID,
		"post_type":   "verification",
	}

	verificationResp, err := postMultipartForm(client3, testServer.URL+"/api/post/create", verificationFields)
	require.NoError(t, err)
	defer verificationResp.Body.Close()

	// Should redirect to community page with verification post ID
	assert.Equal(t, http.StatusFound, verificationResp.StatusCode)

	// Step 6: Create a chat post as a reply to the issue
	chatTitle := "Discussion about " + issueTitle
	chatFields := map[string]string{
		"title":       chatTitle,
		"body":        "Let me know if you need more clarification",
		"tags":        "chat,test",
		"community":   communityName,
		"reply_to_id": issuePostID,
		"post_type":   "chat",
	}

	chatResp, err := postMultipartForm(client, testServer.URL+"/api/post/create", chatFields)
	require.NoError(t, err)
	defer chatResp.Body.Close()

	// Should redirect to community page with chat post ID
	assert.Equal(t, http.StatusFound, chatResp.StatusCode)

	// Step 7: Verify the issue post page shows the solution and chat replies
	issuePageResp, err := client.Get(testServer.URL + "/p/" + issuePostID)
	require.NoError(t, err)
	defer issuePageResp.Body.Close()

	assert.Equal(t, http.StatusOK, issuePageResp.StatusCode)
	
	issuePageBody := readResponseBody(t, issuePageResp)
	
	// Check that the issue post appears
	assert.Contains(t, issuePageBody, issueTitle)
	assert.Contains(t, issuePageBody, "This is a test issue that needs a solution")
	
	// Check that the solution appears in the Solutions section
	assert.Contains(t, issuePageBody, solutionTitle)
	assert.Contains(t, issuePageBody, "Solutions")
	
	// Check that the chat appears in the Discussion section
	assert.Contains(t, issuePageBody, chatTitle)
	assert.Contains(t, issuePageBody, "Discussion")
	
	// Check that the "Solve This" link is present and points to the correct URL
	expectedSolveURL := "/c/" + communityName + "/post?reply_to_id=" + issuePostID + "&post_type=solution"
	assert.Contains(t, issuePageBody, expectedSolveURL)
	
	// Check that the "Reply" link is present for chat
	expectedReplyURL := "/c/" + communityName + "/post?reply_to_id=" + issuePostID + "&post_type=chat"
	assert.Contains(t, issuePageBody, expectedReplyURL)
}

func TestPostReplyGetHandler(t *testing.T) {
	// Create a client with cookie jar to maintain session across requests
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	
	client := &http.Client{
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

	assert.Equal(t, http.StatusFound, registerResp.StatusCode)

	// Step 2: Create a community first
	communityName := "test-community-" + strconv.FormatInt(timestamp, 10)
	communityFields := map[string]string{
		"name":     communityName,
		"title":    "Test Community " + strconv.FormatInt(timestamp, 10),
		"location": "Test Location",
	}

	communityResp, err := postMultipartForm(client, testServer.URL+"/api/community/create", communityFields)
	require.NoError(t, err)
	defer communityResp.Body.Close()

	assert.Equal(t, http.StatusFound, communityResp.StatusCode)

	// Step 3: Create an issue post
	issueFields := map[string]string{
		"title":     "Test Issue " + strconv.FormatInt(timestamp, 10),
		"body":      "This is a test issue",
		"tags":      "issue,test",
		"community": communityName,
	}

	issueResp, err := postMultipartForm(client, testServer.URL+"/api/post/create", issueFields)
	require.NoError(t, err)
	defer issueResp.Body.Close()

	assert.Equal(t, http.StatusFound, issueResp.StatusCode)
	
	// Extract the issue post ID from the redirect URL
	issueLocation := issueResp.Header.Get("Location")
	parts := strings.Split(issueLocation, "posted_id=")
	require.Len(t, parts, 2)
	issuePostID := parts[1]

	// Step 4: Test the post create page with query parameters
	createPageURL := testServer.URL + "/c/" + communityName + "/post?reply_to_id=" + issuePostID + "&post_type=solution"
	createPageResp, err := client.Get(createPageURL)
	require.NoError(t, err)
	defer createPageResp.Body.Close()

	assert.Equal(t, http.StatusOK, createPageResp.StatusCode)
	
	createPageBody := readResponseBody(t, createPageResp)
	
	// Check that the form contains the hidden fields for reply_to_id and post_type
	assert.Contains(t, createPageBody, `name="reply_to_id" value="`+issuePostID+`"`)
	assert.Contains(t, createPageBody, `name="post_type" value="solution"`)
}

