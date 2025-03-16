package http

import (
	"net/http"

	"wemaps/internal/services"
)

type Server struct {
	healthService *services.Health
	coordService  *services.GeolocationService
}

func NewServer() *Server {
	return &Server{
		healthService: services.NewHealthService(),
	}
}

func (s *Server) Start(port string) error {
	// ANGULAR.JS
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./static/browser/"))))

	// Endpoint de health check
	http.HandleFunc("/health", s.healthHandler)

	// Endpoint de CORE GEO CODER!!
	http.HandleFunc("/getcoords/", s.getCoordsHandler)

	return http.ListenAndServe(":"+port, nil)
}
