package http

import (
	"net/http"

	"wemaps/internal/ports"
	"wemaps/internal/services"
)

type Server struct {
	healthService *services.Health
	coordService  *services.GeolocationService
	reports       services.CoordsRequest
}

func NewServer(repo ports.GeolocationRepository) *Server {
	s := &Server{
		healthService: services.NewHealthService(),
		coordService:  services.NewGeolocationService(repo),
		reports:       services.CoordsRequest{}, // Buffer para reportes
	}
	return s
}

func (s *Server) Start(port string) error {
	// ANGULAR.JS
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./static/browser/"))))

	// Endpoint de health check
	http.HandleFunc("/api/health", s.healthHandler)

	http.HandleFunc("/api/submitcoords", s.submitCoordsHandler) // POST para enviar el reporte
	http.HandleFunc("/api/getcoords/", s.getCoordsHandler)      // SSE para resultados

	return http.ListenAndServe(":"+port, nil)
}
