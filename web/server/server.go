package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Handler interface {
	RegisterRoutes(router *mux.Router)
}

type Server struct {
	router   *mux.Router
	handlers []Handler
}

func New() *Server {
	return &Server{
		router:   mux.NewRouter(),
		handlers: make([]Handler, 0),
	}
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
