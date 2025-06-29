package app

import (
	"context"
	"fmt"
	"net/http"

	"fixit/engine/auth"
	"fixit/engine/community"
	enginePost "fixit/engine/post"
	webcommunity "fixit/web/community"
	weberrors "fixit/web/errors"
	"fixit/web/frontpage"
	"fixit/web/list"
	"fixit/web/post"
	"fixit/web/server"
)

type App struct {
	server *server.Server
	cfg    Config
}

type Config struct {
	DatabaseURL string `env:"DATABASE_URL" envDefault:"postgres://fixit:password@localhost:5432/fixit?sslmode=disable"`
	Port        int    `env:"PORT" envDefault:"8080"`
	Auth        auth.Config
}

func New(cfg Config) (*App, error) {
	srv := server.New()

	return &App{
		cfg:    cfg,
		server: srv,
	}, nil
}

func (a *App) Initialize() error {
	if err := a.server.InitDB(a.cfg.DatabaseURL); err != nil {
		return err
	}

	ab, err := auth.Setup(a.server.Client(), a.cfg.Auth)
	if err != nil {
		return err
	}

	a.server.Router().Use(auth.Middleware(ab)...)

	a.server.Router().PathPrefix("/auth").Handler(http.StripPrefix("/auth", ab.Config.Core.Router))

	repo := community.NewRepository(a.server.Client())
	if err := repo.Seed(context.Background()); err != nil {
		return err
	}

	// Register frontpage handler
	frontpageHandler := frontpage.New(repo)
	a.server.RegisterHandler(frontpageHandler)

	listHandler := list.New(a.server.Client())
	a.server.RegisterHandler(listHandler)

	postRepo := enginePost.New(a.server.Client())
	postHandler := post.New(postRepo, repo, ab)
	a.server.RegisterHandler(postHandler)

	communityHandler := webcommunity.New([]byte(a.cfg.Auth.SessionKey), repo, ab)
	a.server.RegisterHandler(communityHandler)

	// Set up 404 handler for unmatched routes
	a.server.Router().NotFoundHandler = http.HandlerFunc(weberrors.NotFoundHandler)

	return nil
}

func (a *App) Router() http.Handler {
	return a.server.Router()
}

func (a *App) Start() error {
	if err := a.Initialize(); err != nil {
		return err
	}

	return a.server.Start(fmt.Sprintf(":%d", a.cfg.Port))
}
