package factory

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"fixit/engine/ent"
)

// User creates a test user with a unique username
func User(t *testing.T, client *ent.Client, usernamePattern string) *ent.User {
	username := Placeholder(usernamePattern)
	user, err := client.User.Create().
		SetUsername(username).
		SetEmail(username + "@example.com").
		SetPassword("password123").
		Save(context.Background())
	require.NoError(t, err)
	return user
}

// Community creates a test community with a unique name
func Community(t *testing.T, client *ent.Client, namePattern string) *ent.Community {
	name := Placeholder(namePattern)
	community, err := client.Community.Create().
		SetName(name).
		SetTitle("Test Community " + name).
		Save(context.Background())
	require.NoError(t, err)
	return community
}

// Placeholder generates a unique name by replacing '*' in the pattern with a random base64 string.
// Example: Placeholder("issue-author-*") -> "issue-author-abc123xyz"
func Placeholder(pattern string) string {
	// Generate random bytes for base64 encoding
	randomBytes := make([]byte, 9) // 9 bytes = 12 base64 chars
	rand.Read(randomBytes)

	// Encode to base64 and make URL-safe (replace + and / with - and _)
	randomStr := base64.URLEncoding.EncodeToString(randomBytes)
	// Remove padding
	randomStr = strings.TrimRight(randomStr, "=")

	// Replace '*' with the random string
	return strings.ReplaceAll(pattern, "*", randomStr)
}
