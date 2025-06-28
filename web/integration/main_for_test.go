package integration

import (
	"fmt"
	"net/http/httptest"
	"os"
	"testing"

	"fixit/engine/auth"
	"fixit/engine/config"
	"fixit/web/app"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

var testServer *httptest.Server
var testConfig app.Config

func setup() {
	testConfig = app.Config{
		DatabaseURL: config.TestDBURL,
		Port:        0,
		Auth: auth.Config{
			SessionKey:  "test-32-byte-secret-key-here!!!",
			SendGridKey: "",
			FromEmail:   "test@example.com",
			FromName:    "Test FixIt",
			RootURL:     "http://localhost:8080",
		},
	}

	testApp, err := app.New(testConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to create test app: %v", err))
	}

	if err := testApp.Initialize(); err != nil {
		panic(fmt.Sprintf("Failed to init test app: %v", err))
	}

	testServer = httptest.NewServer(testApp.Router())
	testConfig.Auth.RootURL = testServer.URL
}

func teardown() {
	if testServer != nil {
		testServer.Close()
	}
}
