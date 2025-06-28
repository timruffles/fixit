package app

import (
	"context"
	"fmt"

	"fixit/engine/auth"
	"fixit/engine/community"
	"fixit/web/list"
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

func (a *App) Start() error {
	if err := a.server.InitDB(a.cfg.DatabaseURL); err != nil {
		return err
	}

	// Setup auth
	ab, err := auth.Setup(a.server.Client(), a.cfg.Auth)
	if err != nil {
		return err
	}

	// Add auth middleware
	a.server.Router().Use(auth.Middleware(ab))

	// Mount auth routes
	a.server.Router().PathPrefix("/auth").Handler(ab.Config.Core.Router)

	repo := community.NewRepository(a.server.Client())
	if err := repo.Seed(context.Background()); err != nil {
		return err
	}

	listHandler := list.New(a.server.Client())
	a.server.RegisterHandler(listHandler)

	return a.server.Start(fmt.Sprintf(":%d", a.cfg.Port))
}
