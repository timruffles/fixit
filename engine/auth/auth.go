package auth

import (
	"net/http"

	"github.com/aarondl/authboss/v3"
	_ "github.com/aarondl/authboss/v3/auth"
	"github.com/aarondl/authboss/v3/defaults"
	_ "github.com/aarondl/authboss/v3/register"
	"github.com/gorilla/mux"

	"fixit/engine/ent"
	webauth "fixit/web/auth"
)

type Config struct {
	SessionKey  string `env:"SESSION_KEY" envDefault:"your-32-byte-secret-key-here!!"`
	SendGridKey string `env:"SENDGRID_API_KEY"`
	FromEmail   string `env:"FROM_EMAIL" envDefault:"noreply@fixit.local"`
	FromName    string `env:"FROM_NAME" envDefault:"FixIt"`
	RootURL     string `env:"ROOT_URL" envDefault:"http://localhost:8080"`
}

func Setup(client *ent.Client, cfg Config) (*authboss.Authboss, error) {
	ab := authboss.New()

	// Set up defaults first
	defaults.SetCore(&ab.Config, false, false)

	// Configure storage
	ab.Config.Storage.Server = NewStorer(client)
	sessionStorer := NewSessionStorer([]byte(cfg.SessionKey))
	ab.Config.Storage.SessionState = sessionStorer
	ab.Config.Storage.CookieState = sessionStorer

	// Configure paths
	ab.Config.Paths.Mount = "/auth"
	ab.Config.Paths.AuthLoginOK = "/"
	
	ab.Config.Paths.RegisterOK = "/"
	ab.Config.Paths.RootURL = cfg.RootURL

	// Use our custom renderer
	ab.Config.Core.ViewRenderer = webauth.NewRenderer()

	if cfg.SendGridKey != "" {
		ab.Config.Core.Mailer = NewMailer(cfg.SendGridKey, cfg.FromName, cfg.FromEmail)
	}

	if err := ab.Init(); err != nil {
		return nil, err
	}

	return ab, nil
}

func Middleware(ab *authboss.Authboss) []mux.MiddlewareFunc {
	return []mux.MiddlewareFunc{
		func(next http.Handler) http.Handler {
			return ab.LoadClientStateMiddleware(next)
		},
	}
}
