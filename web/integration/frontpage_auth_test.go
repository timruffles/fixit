package integration

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFrontpageAuthDisplay(t *testing.T) {
	// Test 1: Anonymous user should see login/signup links
	t.Run("Anonymous user sees login/signup links", func(t *testing.T) {
		resp, err := http.Get(testServer.URL + "/")
		require.NoError(t, err)
		defer resp.Body.Close()

		body := readResponseBody(t, resp)
		assert.Contains(t, body, "Login")
		assert.Contains(t, body, "Sign Up")
		assert.Contains(t, body, "/auth/login")
		assert.Contains(t, body, "/auth/register")
	})

	// Test 2: Authenticated user should see username
	t.Run("Authenticated user sees username", func(t *testing.T) {
		// Create a client with cookie jar to maintain session
		jar, err := cookiejar.New(nil)
		require.NoError(t, err)

		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Jar: jar,
		}

		// Generate unique credentials
		timestamp := time.Now().Unix()
		username := "testuser" + strconv.FormatInt(timestamp, 10)
		email := username + "@test.com"
		password := "ValidPassword123!"

		// Register a new user
		registerData := url.Values{
			"username":         {username},
			"email":            {email},
			"password":         {password},
			"confirm_password": {password},
		}

		registerResp, err := client.PostForm(testServer.URL+"/auth/register", registerData)
		require.NoError(t, err)
		defer registerResp.Body.Close()

		// Should redirect after successful registration
		assert.Equal(t, http.StatusFound, registerResp.StatusCode)

		// Now visit the frontpage as authenticated user
		frontpageResp, err := client.Get(testServer.URL + "/")
		require.NoError(t, err)
		defer frontpageResp.Body.Close()

		body := readResponseBody(t, frontpageResp)
		
		// Should show username and not show login/signup links
		assert.Contains(t, body, "@"+username)
		assert.NotContains(t, body, "Login")
		assert.NotContains(t, body, "Sign Up")
		assert.NotContains(t, body, "/auth/login")
		assert.NotContains(t, body, "/auth/register")
	})
}