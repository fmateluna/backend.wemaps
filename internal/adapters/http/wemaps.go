package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"wemaps/internal/adapters/http/dto"
)

func (s *Server) addressInfoHandler(w http.ResponseWriter, r *http.Request) {
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

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	addressInfo, err := s.portalService.GetAddressInfoByUserId(user.ID)

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(addressInfo) == 0 {
		// Return empty array instead of error for no results
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]dto.AddressReport{})
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(addressInfo); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (s *Server) reportSummaryHandler(w http.ResponseWriter, r *http.Request) {
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

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	summaries, err := s.portalService.GetReportSummaryByUserId(user.ID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(summaries) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]dto.ReportSummary{})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(summaries); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (s *Server) reportRowsHandler(w http.ResponseWriter, r *http.Request) {
	// Validar el token de autenticación
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

	// Validar el método HTTP
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Obtener el parámetro report_id de la query string
	reportIDStr := r.URL.Query().Get("report_id")
	if reportIDStr == "" {
		http.Error(w, "Missing report_id parameter", http.StatusBadRequest)
		return
	}
	reportID, err := strconv.Atoi(reportIDStr)
	if err != nil {
		http.Error(w, "Invalid report_id parameter", http.StatusBadRequest)
		return
	}
	/*
		authorized, err := s.portalService.IsUserAuthorizedForReport(user.ID, reportID)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if !authorized {
			http.Error(w, "Unauthorized access to report", http.StatusForbidden)
			return
		}
	*/
	// Obtener los datos del reporte
	rows, err := s.portalService.GetReportRowsByReportID(reportID)
	if err != nil {
		log.Printf("Error fetching report rows: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Retornar un array vacío si no hay resultados
	if len(rows) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]dto.ReportRow{})
		return
	}

	// Enviar la respuesta JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(rows); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
