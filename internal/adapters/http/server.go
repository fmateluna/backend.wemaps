package http

import (
	"fmt"
	"net/http"
	"sync"

	"wemaps/internal/ports"
	"wemaps/internal/services"
)

type Server struct {
	healthService *services.Health
	coordService  *services.GeolocationService
	portalService *services.PortalService
	reports       services.CoordsReportRequest
	addressUnique []string
	mu            sync.Mutex
	sessions      map[string]*ReportSession //CEREBRO DE MULTISESION!!
	sessionsMutex sync.RWMutex
}

func NewServer(repoAddress ports.GeolocationRepository, portalRepo ports.PortalRepository) *Server {
	s := &Server{
		healthService: services.NewHealthService(),
		coordService:  services.NewGeolocationService(repoAddress, portalRepo),
		portalService: services.NewPortalService(portalRepo),
		reports:       services.CoordsReportRequest{},
		sessions:      make(map[string]*ReportSession),
	}
	return s
}

func fileServerWithHeaders(fs http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

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
	mux.HandleFunc("/api/health", s.AuthMiddleware(s.healthHandler))
	mux.HandleFunc("/api/submitcoords", s.submitCoordsHandler)
	mux.HandleFunc("/api/getcoords/", s.getCoordsHandler)
	mux.HandleFunc("/api/coordinates", s.getSingleAddressCoordsHandler)
	mux.HandleFunc("/api/token", s.getTokenHandler)

	//login
	mux.HandleFunc("/api/login", s.logInHandler)
	mux.HandleFunc("/api/logout", s.logOutHandler)

	//porta
	mux.HandleFunc("/portal/addressInfo", s.AuthMiddleware(s.addressInfoHandler))
	mux.HandleFunc("/portal/addressInfoPeerPage", s.AuthMiddleware(s.addressInfoHandlerPeerPage))
	mux.HandleFunc("/portal/reports", s.AuthMiddleware(s.reportSummaryHandler))
	mux.HandleFunc("/portal/report", s.AuthMiddleware(s.reportRowsHandler))
	mux.HandleFunc("/portal/countInfo", s.AuthMiddleware(s.countInfo))

	addr := ":" + port

	if certFile != "" && keyFile != "" {
		fmt.Println("Servidor HTTPS escuchando en puerto", port)
		return http.ListenAndServeTLS(addr, certFile, keyFile, mux)
	}

	fmt.Println("Servidor HTTP escuchando en puerto", port)
	return http.ListenAndServe(addr, mux)
}
