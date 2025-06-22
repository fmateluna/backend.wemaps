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
)

func (s *Server) getSingleAddressCoordsHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Handle preflight OPTIONS request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Validate HTTP method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	_, err := s.portalService.ValidateToken(token)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
		return
	}

	// Get address from query parameter
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "Missing address query parameter", http.StatusBadRequest)
		return
	}

	// Sanitize address
	address = sanitizeString(address)

	// Fetch coordinates
	geo, err := s.coordService.GetCoordsFromAddress(address)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to geocode address: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := struct {
		FormattedAddress string  `json:"formatted_address"`
		Latitude         float64 `json:"latitude"`
		Longitude        float64 `json:"longitude"`
	}{
		FormattedAddress: geo.FormattedAddress,
		Latitude:         geo.Latitude,
		Longitude:        geo.Longitude,
	}

	// If geocoding failed, return default values
	if geo.FormattedAddress == address {
		response.FormattedAddress = "No se pudo geolocalizar"
		response.Latitude = 0
		response.Longitude = 0
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

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

	if len(s.reports.Columns) > 0 {
		keyAddresToGeoCoding := s.reports.Columns[0]
		addressToGeoCoding := s.reports.Values[keyAddresToGeoCoding]
		nok := 0
		ok := 0

		//s.portalService.SaveReportInfo()

		for index, address := range addressToGeoCoding {

			geo, err := geolocationService.GetCoordsFromAddress(address)

			status := domain.StatusGeoResult{}
			status.Count = index + 1
			status.Total = len(addressToGeoCoding)
			status.Ok = ok
			status.Nok = nok
			status.Result = (err == nil)

			infoReport := make(map[string]string)
			for _, col := range s.reports.Columns {
				if col == keyAddresToGeoCoding {
					infoReport[col] = geo.OriginAddress
				} else {
					infoReport[col] = s.reports.Values[col][index]
				}
			}

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

				infoReport[keyAddresToGeoCoding] = address
				infoReport["Dirección Normalizada"] = "-"
				infoReport["Latitud"] = fmt.Sprintf("%f", geo.Latitude)
				infoReport["Longitud"] = fmt.Sprintf("%f", geo.Longitude)
				//s.portalService.SaveReportInfo(user.ID, s.reports.ReportName, infoReport, geo, token, index)

				data, _ := json.Marshal(geo)
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()

				continue
			} else {
				geo.OriginAddress = address
				ok = ok + 1
				geo.Status.Ok = ok
				geo.Status = status
				infoReport["Dirección Normalizada"] = geo.FormattedAddress
				infoReport["Latitud"] = fmt.Sprintf("%f", geo.Latitude)
				infoReport["Longitud"] = fmt.Sprintf("%f", geo.Longitude)
				//s.portalService.SaveReportInfo(user.ID, s.reports.ReportName, infoReport, geo, token, index)

				data, _ := json.Marshal(geo)
				_, err = fmt.Fprintf(w, "data: %s\n\n", data)
				if err != nil {
					return
				}

			}

			s.saveToPortal(user.ID, geo, infoReport, token, index)
			flusher.Flush()
		}
	}

	_, errLoad := fmt.Fprintf(w, "data: {\"status\": \"done\"}\n\n")
	if errLoad == nil {
		flusher.Flush()
	}
}

func (s *Server) saveToPortal(userID int, geo domain.Geolocation, infoReport map[string]string, token string, index int) {
	found := false
	for _, addr := range s.addressUnique {
		if addr == geo.FormattedAddress {
			found = true
			break
		}
	}
	if !found {
		s.addressUnique = append(s.addressUnique, geo.FormattedAddress)
		go s.portalService.SaveReportInfo(userID, s.reports.ReportName, infoReport, geo, token, index)
	} else {
		s.portalService.SaveReportInfo(userID, s.reports.ReportName, infoReport, geo, token, index)
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
