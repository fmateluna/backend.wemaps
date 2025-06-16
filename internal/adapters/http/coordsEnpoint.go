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

				infoReport["Dirección Normalizada"] = "-"
				infoReport["Latitud"] = fmt.Sprintf("%f", geo.Latitude)
				infoReport["Longitud"] = fmt.Sprintf("%f", geo.Longitude)	
				go s.portalService.SaveReportInfo(user.ID,s.reports.ReportName,infoReport,geo , token ,index)

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
				go s.portalService.SaveReportInfo(user.ID,s.reports.ReportName,infoReport,geo , token ,index)

				data, _ := json.Marshal(geo)
				_, err = fmt.Fprintf(w, "data: %s\n\n", data)
				if err != nil {
					return
				}
				
			}
			


			
			flusher.Flush()
		}
	}

	_, errLoad := fmt.Fprintf(w, "data: {\"status\": \"done\"}\n\n")
	if errLoad == nil {
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
