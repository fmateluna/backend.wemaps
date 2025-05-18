package http

import (
	"fmt"
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

func (s *Server) StartServer(port, certFile, keyFile string) error {
	mux := http.NewServeMux()

	// Angular estático
	mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./static/browser/"))))

	// Endpoints API
	mux.HandleFunc("/api/health", s.healthHandler)
	mux.HandleFunc("/api/submitcoords", s.submitCoordsHandler)
	mux.HandleFunc("/api/getcoords/", s.getCoordsHandler)

	addr := ":" + port

	if certFile != "" && keyFile != "" {
		fmt.Println("Servidor HTTPS escuchando en puerto", port)
		return http.ListenAndServeTLS(addr, certFile, keyFile, mux)
	}

	fmt.Println("Servidor HTTP escuchando en puerto", port)
	return http.ListenAndServe(addr, mux)
}
