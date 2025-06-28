package integration

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthLoginPage(t *testing.T) {
	resp, err := http.Get(testServer.URL + "/auth/login")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/html")

	body := readResponseBody(t, resp)

	assert.Contains(t, body, "Sign in to your account")
	assert.Contains(t, body, `<input id="email" name="email" type="email"`)
	assert.Contains(t, body, `<input id="password" name="password" type="password"`)
	assert.Contains(t, body, `action="/auth/login" method="POST"`)
	assert.Contains(t, body, "Don't have an account? Sign up")
}

func TestAuthRegisterPage(t *testing.T) {
	resp, err := http.Get(testServer.URL + "/auth/register")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/html")

	body := readResponseBody(t, resp)

	assert.Contains(t, body, "Create your account")
	assert.Contains(t, body, `<input id="username" name="username" type="text"`)
	assert.Contains(t, body, `<input id="email" name="email" type="email"`)
	assert.Contains(t, body, `<input id="password" name="password" type="password"`)
	assert.Contains(t, body, `<input id="confirm_password" name="confirm_password" type="password"`)
	assert.Contains(t, body, `action="/auth/register" method="POST"`)
	assert.Contains(t, body, "Already have an account? Sign in")
}

func TestPostsRenderWithUsernames(t *testing.T) {
	resp, err := http.Get(testServer.URL + "/")
	require.NoError(t, err)
	defer resp.Body.Close()

	body := readResponseBody(t, resp)

	expectedUsers := []string{"alice", "bob", "charlie", "diana"}
	foundUsers := 0

	for _, user := range expectedUsers {
		if strings.Contains(body, user) {
			foundUsers++
		}
	}

	assert.GreaterOrEqual(t, foundUsers, 1, "At least one seeded user should appear in the posts")
}

func TestSignupWithValidationErrors(t *testing.T) {
	// Prepare form data that will trigger validation errors (weak password)
	formData := url.Values{
		"username": {"testuser"},
		"email":    {"test@example.com"},
		"password": {"weak"},
	}

	resp, err := http.PostForm(testServer.URL+"/auth/register", formData)
	require.NoError(t, err)
	defer resp.Body.Close()

	body := readResponseBody(t, resp)

	// Should return 200 with validation error displayed
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Check that the form is re-displayed (indicating validation failure)
	assert.Contains(t, body, "Create your account")
	assert.Contains(t, body, `action="/auth/register" method="POST"`)

	// Check for validation error indicators
	hasError := strings.Contains(body, "Does not match password") ||
		strings.Contains(body, "Must contain at least") ||
		strings.Contains(body, "text-red-600")

	assert.True(t, hasError, "Expected validation errors to be displayed for weak password")
}

func TestSignupSuccess(t *testing.T) {
	// Prepare form data with valid credentials including matching passwords
	// Use a unique email to avoid conflicts with seeded data
	timestamp := time.Now().Unix()
	email := fmt.Sprintf("user%d@test.com", timestamp)

	formData := url.Values{
		"username":         {fmt.Sprintf("user%d", timestamp)},
		"email":            {email},
		"password":         {"ValidPassword123!"},
		"confirm_password": {"ValidPassword123!"},
	}

	t.Logf("Submitting form data: %+v", formData)

	resp, err := http.PostForm(testServer.URL+"/auth/register", formData)
	require.NoError(t, err)
	defer resp.Body.Close()

	body := readResponseBody(t, resp)

	t.Logf("Response status: %d", resp.StatusCode)
	if len(body) > 500 {
		t.Logf("Response body (truncated): %s...", body[:500])
	} else {
		t.Logf("Response body: %s", body)
	}

	// Check for successful registration
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// A successful registration should show the main page, not the registration form
	isMainPage := strings.Contains(body, "Community Issues") &&
		strings.Contains(body, "Large pothole on Main Street")

	isRegistrationForm := strings.Contains(body, "Create your account")

	assert.True(t, isMainPage, "Expected to be redirected to main page after successful registration")
	assert.False(t, isRegistrationForm, "Should not show registration form after successful signup")
}

func TestLoginSuccess(t *testing.T) {
	// First create a user account
	timestamp := time.Now().Unix()
	email := fmt.Sprintf("loginuser%d@test.com", timestamp)
	password := "LoginPassword123!"

	// Register the user
	signupData := url.Values{
		"username":         {fmt.Sprintf("loginuser%d", timestamp)},
		"email":            {email},
		"password":         {password},
		"confirm_password": {password},
	}

	resp, err := http.PostForm(testServer.URL+"/auth/register", signupData)
	require.NoError(t, err)
	resp.Body.Close()

	// Now test login with the same credentials
	loginData := url.Values{
		"email":    {email},
		"password": {password},
	}

	resp, err = http.PostForm(testServer.URL+"/auth/login", loginData)
	require.NoError(t, err)
	defer resp.Body.Close()

	body := readResponseBody(t, resp)

	// Check for successful login
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// A successful login should show the main page, not the login form
	isMainPage := strings.Contains(body, "Community Issues") &&
		strings.Contains(body, "Large pothole on Main Street")

	isLoginForm := strings.Contains(body, "Sign in to your account")

	assert.True(t, isMainPage, "Expected to be redirected to main page after successful login")
	assert.False(t, isLoginForm, "Should not show login form after successful login")
}

func TestLoginWithInvalidCredentials(t *testing.T) {
	// Attempt login with non-existent user
	loginData := url.Values{
		"email":    {"nonexistent@test.com"},
		"password": {"WrongPassword123!"},
	}

	resp, err := http.PostForm(testServer.URL+"/auth/login", loginData)
	require.NoError(t, err)
	defer resp.Body.Close()

	body := readResponseBody(t, resp)

	// Should return 200 with login form re-displayed and error message
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Should show the login form again (indicating failed login)
	assert.Contains(t, body, "Sign in to your account")
	assert.Contains(t, body, `action="/auth/login" method="POST"`)

	// Should show an error message
	hasError := strings.Contains(body, "text-red-600") ||
		strings.Contains(body, "error") ||
		strings.Contains(body, "invalid") ||
		strings.Contains(body, "incorrect")

	assert.True(t, hasError, "Expected error message for invalid login credentials")
}

func readResponseBody(t *testing.T, resp *http.Response) string {
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return string(body)
}
