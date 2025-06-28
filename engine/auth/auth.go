package auth

import (
	"net/http"

	"github.com/aarondl/authboss/v3"
	_ "github.com/aarondl/authboss/v3/auth"
	"github.com/aarondl/authboss/v3/defaults"
	_ "github.com/aarondl/authboss/v3/register"

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

	ab.Config.Storage.Server = NewStorer(client)
	ab.Config.Storage.SessionState = NewSessionStorer([]byte(cfg.SessionKey))
	ab.Config.Storage.CookieState = NewSessionStorer([]byte(cfg.SessionKey))

	ab.Config.Paths.Mount = "/auth"
	ab.Config.Paths.RootURL = cfg.RootURL

	// Use our custom renderer
	ab.Config.Core.ViewRenderer = webauth.NewRenderer()

	if cfg.SendGridKey != "" {
		ab.Config.Core.Mailer = NewMailer(cfg.SendGridKey, cfg.FromName, cfg.FromEmail)
	}

	defaults.SetCore(&ab.Config, false, false)

	if err := ab.Init(); err != nil {
		return nil, err
	}

	return ab, nil
}

func Middleware(ab *authboss.Authboss) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return ab.LoadClientStateMiddleware(next)
	}
}
