package app

import (
	"fixit/web/list"
	"fixit/web/server"
)

type App struct {
	server *server.Server
}

func New() *App {
	srv := server.New()

	listHandler := list.New()
	srv.RegisterHandler(listHandler)

	return &App{
		server: srv,
	}
}

func (a *App) Start(addr string) error {
	return a.server.Start(addr)
}
