package factory_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"fixit/engine/factory"
)

func TestPlaceholder(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
	}{
		{"single asterisk", "user-*"},
		{"multiple asterisks", "user-*-test-*"},
		{"no asterisk", "user-test"},
		{"asterisk at start", "*-user"},
		{"asterisk at end", "user-*"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := factory.Placeholder(tt.pattern)
			
			// Should not contain any asterisks
			assert.NotContains(t, result, "*")
			
			// Should not be empty
			assert.NotEmpty(t, result)
			
			// If pattern had no asterisk, result should be identical
			if !strings.Contains(tt.pattern, "*") {
				assert.Equal(t, tt.pattern, result)
			}
		})
	}
}

func TestPlaceholderUniqueness(t *testing.T) {
	pattern := "user-*"
	
	// Generate multiple placeholders and ensure they're unique
	results := make(map[string]bool)
	for i := 0; i < 100; i++ {
		result := factory.Placeholder(pattern)
		assert.False(t, results[result], "Placeholder should generate unique values")
		results[result] = true
	}
}

func TestPlaceholderFormat(t *testing.T) {
	result := factory.Placeholder("user-*")
	
	// Should start with "user-"
	assert.True(t, strings.HasPrefix(result, "user-"))
	
	// The random part should be URL-safe base64 (no + or /)
	randomPart := strings.TrimPrefix(result, "user-")
	assert.NotContains(t, randomPart, "+")
	assert.NotContains(t, randomPart, "/")
	assert.NotContains(t, randomPart, "=") // No padding
}