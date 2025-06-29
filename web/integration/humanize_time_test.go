package integration

import (
	"strings"
	"testing"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/stretchr/testify/assert"
)

func TestHumanizeTimeFunction(t *testing.T) {
	// Test with various time differences to ensure humanize works
	now := time.Now()
	
	testCases := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{
			name:     "5 minutes ago",
			time:     now.Add(-5 * time.Minute),
			expected: "5 minutes ago",
		},
		{
			name:     "1 hour ago",
			time:     now.Add(-1 * time.Hour),
			expected: "1 hour ago",
		},
		{
			name:     "2 days ago", 
			time:     now.Add(-48 * time.Hour),
			expected: "2 days ago",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := humanize.Time(tc.time)
			t.Logf("Input time: %s, Humanized: %s", tc.time.Format("2006-01-02 15:04:05"), result)
			
			// Basic check that it contains "ago" for past times
			if tc.time.Before(now) {
				assert.True(t, strings.Contains(result, "ago"), "Past time should contain 'ago': %s", result)
			}
			
			// Check that it doesn't contain the year format we're replacing
			assert.False(t, strings.Contains(result, "2006"), "Should not contain absolute date format")
			assert.False(t, strings.Contains(result, "Jan 2"), "Should not contain absolute date format")
		})
	}
	
	// Test that very recent time shows as "now" or similar
	veryRecent := now.Add(-1 * time.Second)
	recentResult := humanize.Time(veryRecent)
	t.Logf("Very recent time humanized: %s", recentResult)
	
	// Should be "now" or "1 second ago" or similar
	assert.True(t, 
		strings.Contains(recentResult, "now") || 
		strings.Contains(recentResult, "second") ||
		strings.Contains(recentResult, "moment"),
		"Very recent time should be humanized appropriately: %s", recentResult)
}