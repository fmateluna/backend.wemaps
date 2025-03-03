package http

import (
	"net/http"

	"wemaps/internal/services"
)

type Server struct {
	healthService *services.Health
}

func NewServer() *Server {
	return &Server{
		healthService: services.NewHealthService(),
	}
}

func (s *Server) Start(port string) error {
	// Servir archivos estáticos
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./static/browser/"))))

	// Endpoint de health check
	http.HandleFunc("/health", s.healthHandler)

	// Iniciar servidor
	return http.ListenAndServe(":"+port, nil)
}
