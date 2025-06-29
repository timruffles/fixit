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

func TestUserCanSolveOwnIssueButNotVerifyOwnSolution(t *testing.T) {
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

	// Step 4: Create a solution post as a reply to the issue (this should work now)
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

	// Step 5: Try to create a verification post as a reply to own solution (this should fail)
	verificationTitle := "Verification for solution"
	verificationFields := map[string]string{
		"title":       verificationTitle,
		"body":        "This solution works great!",
		"tags":        "verification,test",
		"community":   communityName,
		"reply_to_id": solutionPostID,
		"post_type":   "verification",
	}

	verificationResp, err := postMultipartForm(client, testServer.URL+"/api/post/create", verificationFields)
	require.NoError(t, err)
	defer verificationResp.Body.Close()

	// Should return 400 Bad Request because users can't verify their own solutions
	assert.Equal(t, http.StatusBadRequest, verificationResp.StatusCode)

	// Check that the error message is about not being able to verify own solutions
	verificationBody := readResponseBody(t, verificationResp)
	assert.Contains(t, verificationBody, "Sorry - you can't verify your own solution. Wait till someone notices your good deed")
}
