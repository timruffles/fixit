package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFrontPage(t *testing.T) {
	resp, err := http.Get(testServer.URL + "/")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/html")

	body := readResponseBody(t, resp)

	assert.Contains(t, body, "<title>Community Issues</title>")
	// Check that the page title is set correctly in the header

	assert.Contains(t, body, "Large pothole on Main Street")
	assert.Contains(t, body, "Graffiti on playground equipment")
	assert.Contains(t, body, "alice")
	assert.Contains(t, body, "bob")
}
