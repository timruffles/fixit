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

func TestMultipleVerifications(t *testing.T) {
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

	// Step 1: Register the original user
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

	// Step 2: Create a community
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
	issueTitle := "Multi-Verification Test Issue " + strconv.FormatInt(timestamp, 10)
	issueFields := map[string]string{
		"title":     issueTitle,
		"body":      "This issue will have a solution with multiple verifications",
		"tags":      "issue,test",
		"community": communityName,
	}

	issueResp, err := postMultipartForm(client, testServer.URL+"/api/post/create", issueFields)
	require.NoError(t, err)
	defer issueResp.Body.Close()

	assert.Equal(t, http.StatusFound, issueResp.StatusCode)
	issueLocation := issueResp.Header.Get("Location")
	
	parts := strings.Split(issueLocation, "posted_id=")
	require.Len(t, parts, 2)
	issuePostID := parts[1]

	// Step 4: Create a solution
	solutionTitle := "Solution for multi-verification test"
	solutionFields := map[string]string{
		"title":       solutionTitle,
		"body":        "This is a great solution that will get multiple verifications",
		"tags":        "solution,test",
		"community":   communityName,
		"reply_to_id": issuePostID,
		"post_type":   "solution",
	}

	solutionResp, err := postMultipartForm(client, testServer.URL+"/api/post/create", solutionFields)
	require.NoError(t, err)
	defer solutionResp.Body.Close()

	assert.Equal(t, http.StatusFound, solutionResp.StatusCode)
	solutionLocation := solutionResp.Header.Get("Location")
	
	solutionParts := strings.Split(solutionLocation, "posted_id=")
	require.Len(t, solutionParts, 2)
	solutionPostID := solutionParts[1]

	// Step 5: Create multiple verifiers and verifications
	verificationTitles := []string{
		"First verification - works great!",
		"Second verification - confirmed working",
		"Third verification - tested successfully",
	}

	for i, verificationTitle := range verificationTitles {
		// Create a new verifier user
		verifierUsername := "verifier" + strconv.Itoa(i) + "_" + strconv.FormatInt(timestamp, 10)
		verifierEmail := verifierUsername + "@test.com"

		registerDataVerifier := url.Values{
			"username":         {verifierUsername},
			"email":            {verifierEmail},
			"password":         {password},
			"confirm_password": {password},
		}

		// Create client for verifier
		jarVerifier, err := cookiejar.New(nil)
		require.NoError(t, err)
		
		clientVerifier := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Jar: jarVerifier,
		}

		registerRespVerifier, err := clientVerifier.PostForm(testServer.URL+"/auth/register", registerDataVerifier)
		require.NoError(t, err)
		defer registerRespVerifier.Body.Close()

		assert.Equal(t, http.StatusFound, registerRespVerifier.StatusCode)

		// Create verification
		verificationFields := map[string]string{
			"title":       verificationTitle,
			"body":        "Verification details: " + verificationTitle,
			"tags":        "verification,test",
			"community":   communityName,
			"reply_to_id": solutionPostID,
			"post_type":   "verification",
		}

		verificationResp, err := postMultipartForm(clientVerifier, testServer.URL+"/api/post/create", verificationFields)
		require.NoError(t, err)
		defer verificationResp.Body.Close()

		assert.Equal(t, http.StatusFound, verificationResp.StatusCode)
	}

	// Step 6: Verify the issue post page shows all verifications under the solution
	issuePageResp, err := client.Get(testServer.URL + "/p/" + issuePostID)
	require.NoError(t, err)
	defer issuePageResp.Body.Close()

	assert.Equal(t, http.StatusOK, issuePageResp.StatusCode)
	
	issuePageBody := readResponseBody(t, issuePageResp)
	
	// Check that all verification titles appear in the page
	for _, verificationTitle := range verificationTitles {
		assert.Contains(t, issuePageBody, verificationTitle)
	}
	
	// Check that the verification count shows 3
	assert.Contains(t, issuePageBody, "3 verification(s)")
	
	// Check that multiple "✓ Verified" badges appear
	verifiedCount := strings.Count(issuePageBody, "✓ Verified")
	assert.Equal(t, 3, verifiedCount, "Should have 3 verification badges")
	
	// Check that the Verifications section exists
	assert.Contains(t, issuePageBody, "Verifications")
}