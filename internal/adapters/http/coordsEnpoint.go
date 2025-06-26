package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"wemaps/internal/domain"
	"wemaps/internal/services"

	"github.com/google/uuid"
)

type GeoReport struct {
	Geo      domain.Geolocation `json:"geo"`
	Index    int                `json:"index"`
	IdReport int                `json:"-"` //nolint
	IdUser   int                `json:"-"` //nolint
}

const (
	LOAD_INIT     = 1
	LOAD_FINISH   = 3
	LOAD_ERROR    = 4
	STILL_WORKING = 2
)

// ReportSession almacena el reporte y su canal de resultados por sesión
type ReportSession struct {
	Report    services.CoordsReportRequest
	ResultCh  chan GeoReport
	DoneCh    chan struct{}
	CreatedAt time.Time
}

func (s *Server) submitCoordsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	user, err := s.portalService.ValidateToken(token)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
		return
	}

	fmt.Println("Usuario autenticado:", user.Alias)

	var report services.CoordsReportRequest
	if err := json.Unmarshal(bodyBytes, &report); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if report.ReportName == "" {
		now := time.Now()
		formattedTime := now.Format("2006-01-02_150405")
		report.ReportName = "Nuevo Reporte " + formattedTime
	}

	// Generar un ID único para la sesión
	sessionID := uuid.New().String()

	// Crear canales para la sesión
	resultCh := make(chan GeoReport, 1)
	doneCh := make(chan struct{})

	// Guardar la sesión en el mapa
	s.sessionsMutex.Lock()
	s.sessions[sessionID] = &ReportSession{
		Report:    report,
		ResultCh:  resultCh,
		DoneCh:    doneCh,
		CreatedAt: time.Now(),
	}
	s.sessionsMutex.Unlock()

	// Devolver el sessionID al cliente
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"sessionID": sessionID})
}

func (s *Server) getCoordsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
	w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	cookieSessionID, err := r.Cookie("sessionID")
	if err != nil {
		http.Error(w, "Missing auth token cookie", http.StatusUnauthorized)
		return
	}
	sessionID := cookieSessionID.Value
	if sessionID == "" {
		http.Error(w, "Missing sessionID", http.StatusBadRequest)
		return
	}

	// Obtener la sesión del mapa
	s.sessionsMutex.RLock()
	session, exists := s.sessions[sessionID]
	s.sessionsMutex.RUnlock()
	if !exists {
		http.Error(w, "Invalid sessionID", http.StatusBadRequest)
		return
	}

	// Validar token de autenticación
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		http.Error(w, "Missing auth token cookie", http.StatusUnauthorized)
		return
	}
	token := cookie.Value
	user, err := s.portalService.ValidateToken(token)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
		return
	}

	fmt.Println("Usuario autenticado:", user.Alias)

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	geolocationService := s.coordService
	report := session.Report

	if len(report.Columns) == 0 {
		http.Error(w, "No columns in report", http.StatusBadRequest)
		return
	}

	keyAddresToGeoCoding := report.Columns[0]
	addressToGeoCoding := report.Values[keyAddresToGeoCoding]

	// Goroutine para procesar direcciones
	go func() {
		defer func() {
			close(session.DoneCh)
			// Limpiar la sesión al terminar
			s.sessionsMutex.Lock()
			delete(s.sessions, sessionID)
			s.sessionsMutex.Unlock()
		}()

		nok := 0
		ok := 0
		idReport := -1

		for index, address := range addressToGeoCoding {
			geo, err := geolocationService.GetCoordsFromAddress(address)

			status := domain.StatusGeoResult{
				Count:  index + 1,
				Total:  len(addressToGeoCoding),
				Ok:     ok,
				Nok:    nok,
				Result: err == nil,
			}

			infoReport := make(map[string]string)
			for _, col := range report.Columns {
				if col == keyAddresToGeoCoding {
					infoReport[col] = address
				} else {
					infoReport[col] = report.Values[col][index]
				}
			}

			if err != nil {
				nok++
				geo = domain.Geolocation{
					Status:           status,
					OriginAddress:    address,
					FormattedAddress: address,
					Latitude:         0,
					Longitude:        0,
					Geocoder:         "Sin Información: " + err.Error(),
				}
				infoReport["Dirección Normalizada"] = "-"
				infoReport["Latitud"] = fmt.Sprintf("%f", geo.Latitude)
				infoReport["Longitud"] = fmt.Sprintf("%f", geo.Longitude)
			} else {
				ok++
				geo.Status = status
				geo.OriginAddress = address
				infoReport["Dirección Normalizada"] = geo.FormattedAddress
				infoReport["Latitud"] = fmt.Sprintf("%f", geo.Latitude)
				infoReport["Longitud"] = fmt.Sprintf("%f", geo.Longitude)
			}

			// Guardar en el portal

			idReport, _ = s.saveToPortal(user.ID, geo, report.ReportName, infoReport, token, index)

			fmt.Println("Reporte:", report.ReportName, " Origen : ["+geo.Geocoder+"] Dirección:", geo.FormattedAddress)
			// Enviar resultado al canal
			gr := GeoReport{
				Geo:      geo,
				Index:    index,
				IdReport: idReport,
				IdUser:   user.ID,
			}

			//s.portalService.SetStatusReport(gr.IdUser, gr.IdReport, STILL_WORKING)

			select {
			case session.ResultCh <- gr:
			case <-r.Context().Done():
				// Cliente desconectado, continuar procesando
				//s.portalService.SetStatusReport(gr.IdUser, gr.IdReport, LOAD_ERROR)
			}
		}
		s.portalService.SetStatusReport(user.ID, idReport, LOAD_FINISH)
	}()

	// Bucle principal para enviar datos al cliente
	for {
		select {
		case geo, ok := <-session.ResultCh:
			if !ok {
				// Procesamiento completado
				_, err := fmt.Fprintf(w, "data: {\"status\": \"done\"}\n\n")
				if err == nil {
					s.portalService.SetStatusReport(geo.IdUser, geo.IdReport, LOAD_FINISH)
					time.Sleep(2 * time.Second)
					flusher.Flush()
				}
				return
			}
			// Enviar datos al cliente
			data, _ := json.Marshal(geo)
			_, err := fmt.Fprintf(w, "data: %s\n\n", data)
			time.Sleep(2 * time.Second)
			if err != nil {
				s.portalService.SetStatusReport(geo.IdUser, geo.IdReport, LOAD_ERROR)
				return
			}
			flusher.Flush()

		case <-r.Context().Done():
			// Cliente canceló el request, pero la goroutine sigue procesando
			return
		}
	}
}

func (s *Server) saveToPortal(userID int, geo domain.Geolocation, reportName string, infoReport map[string]string, token string, index int) (int, error) {
	/*
		found := false
		for _, addr := range s.addressUnique {
			if addr == geo.FormattedAddress {
				found = true
				break
			}
		}
		if !found {
			s.addressUnique = append(s.addressUnique, geo.FormattedAddress)
		}
	*/
	return s.portalService.SaveReportInfo(userID, reportName, infoReport, geo, token, index)
}

// sanitizeString limpia una cadena para que sea válida en JSON
func sanitizeString(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if r < 32 || r > 126 || r == '"' || r == '\\' {
			sb.WriteString(fmt.Sprintf("\\u%04x", r))
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
