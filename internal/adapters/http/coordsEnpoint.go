package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"wemaps/internal/infrastructure/geocoders"
	"wemaps/internal/services"
)

func (s *Server) submitCoordsHandler(w http.ResponseWriter, r *http.Request) {
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
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	if len(s.reports.Columns) > 0 {
		key := s.reports.Columns[0]
		values := s.reports.Values[key]
		for _, address := range values {
			geolocationService := geocoders.NewNominatimGeocoder()
			geo, err := geolocationService.Geocode(address)
			if err != nil {
				fmt.Fprintf(w, "data: {\"error\": \"Error geolocalizando %s: %v\"}\n\n", address, err)
				flusher.Flush()
				continue
			}
			geo.OriginAddress = address

			// Intentar serializar a JSON
			data, err := json.Marshal(geo)
			if err != nil {
				// Si falla, construir una representación alternativa
				fallbackData := fmt.Sprintf(
					`{"origin_address": %q, "formatted_address": %q, "latitude": %f, "longitude": %f, "geocoder": %q, "marshal_error": %q}`,
					sanitizeString(geo.OriginAddress),
					sanitizeString(geo.FormattedAddress),
					geo.Latitude,
					geo.Longitude,
					sanitizeString(geo.Geocoder),
					err.Error(),
				)
				fmt.Fprintf(w, "data: %s\n\n", fallbackData)
				flusher.Flush()
				continue
			}

			// Enviar el evento SSE normalmente si la serialización funciona
			_, err = fmt.Fprintf(w, "data: %s\n\n", data)
			if err != nil {
				return
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

func (s *Server) getCoordsHandlerX(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	if len(s.reports.Columns) > 0 {
		key := s.reports.Columns[0]
		values := s.reports.Values[key]
		for _, address := range values {
			//geolocationService := services.NewGeolocationService()
			//geo, err := geolocationService.GetCoordsFromAddress(address)

			geolocationService := geocoders.NewNominatimGeocoder()
			geo, err := geolocationService.Geocode(address)
			geo.OriginAddress = address
			if err != nil {
				fmt.Printf("Error geolocalizando %s: %v\n", address, err)
				continue
			}
			data, _ := json.Marshal(geo)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
	fmt.Fprintf(w, "data: {\"status\": \"done\"}\n\n")
	flusher.Flush()
}
