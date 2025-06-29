package config

import "os"

const (
	AppName = "FixIt"
)

// GetTestDBURL returns the test database URL from environment variable or default
func GetTestDBURL() string {
	if dbURL := os.Getenv("TEST_DATABASE_URL"); dbURL != "" {
		return dbURL
	}
	return "postgres://fixit:password@localhost:5432/fixit_test?sslmode=disable"
}
