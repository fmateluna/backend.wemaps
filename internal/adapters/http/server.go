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

func fileServerWithHeaders(fs http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ⚠️ No usar esto si quieres que funcione `signInWithPopup`
		//w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		//w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")

		// Otras cabeceras que sí puedes mantener
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Cache-Control", "no-cache")

		fs.ServeHTTP(w, r)
	})
}

func (s *Server) StartServer(port, certFile, keyFile string) error {
	mux := http.NewServeMux()

	// Angular estático
	fs := http.FileServer(http.Dir("./static/browser/"))
	mux.Handle("/", fileServerWithHeaders(fs))

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
