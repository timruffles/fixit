package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"

	"fixit/engine/ent"
	"fixit/engine/ent/migrate"
)

type Handler interface {
	RegisterRoutes(router *mux.Router)
}

type Server struct {
	router   *mux.Router
	handlers []Handler
	client   *ent.Client
}

func New() *Server {
	return &Server{
		router:   mux.NewRouter(),
		handlers: make([]Handler, 0),
	}
}

func (s *Server) InitDB(databaseURL string) error {
	client, err := ent.Open("postgres", databaseURL)
	if err != nil {
		return errors.WithStack(err)
	}

	s.client = client

	ctx := context.Background()
	if err := client.Debug().Schema.Create(ctx, migrate.WithGlobalUniqueID(true)); err != nil {
		return errors.WithStack(err)
	}

	slog.Info("Database migration completed successfully")
	return nil
}

func (s *Server) Client() *ent.Client {
	return s.client
}

func (s *Server) RegisterHandler(handler Handler) {
	s.handlers = append(s.handlers, handler)
	handler.RegisterRoutes(s.router)
}

func (s *Server) RegisterHandlers(handlers ...Handler) {
	for _, handler := range handlers {
		s.RegisterHandler(handler)
	}
}

func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s.router)
}

func (s *Server) Router() *mux.Router {
	return s.router
}
