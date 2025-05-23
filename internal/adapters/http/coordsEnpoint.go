package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"wemaps/internal/domain"
	"wemaps/internal/services"
)

func (s *Server) submitCoordsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")                            // Permitir todos los orígenes
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")          // Métodos permitidos
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization") // Headers permitidos
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var report services.CoordsRequest
	if err := json.Unmarshal(bodyBytes, &report); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	s.reports = report
	w.Write([]byte("Report submitted"))
}

func (s *Server) getCoordsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")                            // Permitir todos los orígenes
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")          // Métodos permitidos
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization") // Headers permitidos
	w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
	w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	geolocationService := s.coordService

	if len(s.reports.Columns) > 0 {
		key := s.reports.Columns[0]
		values := s.reports.Values[key]
		nok := 0
		ok := 0
		for index, address := range values {

			geo, err := geolocationService.GetCoordsFromAddress(address)

			status := domain.StatusGeoResult{}
			status.Count = index + 1
			status.Total = len(values)
			status.Ok = ok
			status.Nok = nok
			status.Result = (err == nil)

			if err != nil {
				geo := domain.Geolocation{}
				nok++
				status.Nok = nok
				geo.Status = status
				geo.OriginAddress = address
				geo.FormattedAddress = (address)
				geo.Latitude = 0
				geo.Longitude = 0
				geo.Geocoder = "Sin Información : " + err.Error()
				data, _ := json.Marshal(geo)
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
				continue
			} else {
				geo.OriginAddress = address
				ok = ok + 1
				geo.Status.Ok = ok
				geo.Status = status
				data, _ := json.Marshal(geo)
				_, err = fmt.Fprintf(w, "data: %s\n\n", data)
				if err != nil {
					return
				}
			}
			flusher.Flush()
		}
	}

	_, err := fmt.Fprintf(w, "data: {\"status\": \"done\"}\n\n")
	if err == nil {
		flusher.Flush()
	}
}

// sanitizeString limpia una cadena para que sea válida en JSON
func sanitizeString(s string) string {
	// Reemplazar caracteres no válidos por su representación escapada o un placeholder
	var sb strings.Builder
	for _, r := range s {
		if r < 32 || r > 126 || r == '"' || r == '\\' { // Caracteres de control o especiales
			sb.WriteString(fmt.Sprintf("\\u%04x", r))
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
