package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"wemaps/internal/adapters/http/dto"
)

type UserKey struct{}

func (s *Server) GetUserFromContext(r *http.Request) (*dto.UserPortal, error) {
	user, ok := r.Context().Value(UserKey{}).(*dto.UserPortal)
	if !ok || user == nil {
		return nil, fmt.Errorf("user not found in context")
	}
	return user, nil
}

func (s *Server) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Validate the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}

		// Extract the token
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate the token and get the user
		user, err := s.portalService.ValidateToken(token)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
			return
		}

		// Add the user to the request context
		ctx := context.WithValue(r.Context(), UserKey{}, user)
		r = r.WithContext(ctx)

		// Call the next handler
		next(w, r)
	}
}

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

func (s *Server) countInfo(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, err := s.GetUserFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Validar el método HTTP
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	totales, err := s.portalService.GetTotalReportsAndAddress(user.ID)
	if err != nil {
		log.Printf("Error fetching report rows: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Retornar un array vacío si no hay resultados
	if len(totales) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]dto.CategoryCount{})
		return
	}

	// Enviar la respuesta JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(totales); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (s *Server) addressInfoHandlerPeerPage(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	user, err := s.portalService.ValidateToken(token)
	if err != nil {
		http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Obtener parámetros de la query
	query := r.URL.Query().Get("query")
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	// Obtener direcciones con filtrado y paginación
	addressInfo, total, err := s.portalService.GetAddressInfoByUserIdPeerPage(user.ID, query, page, pageSize)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Configurar streaming de la respuesta
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)

	// Iniciar streaming de la respuesta
	enc := json.NewEncoder(w)
	flusher, _ := w.(http.Flusher)

	// Enviar total de resultados
	response := struct {
		Total int                 `json:"total"`
		Data  []dto.AddressReport `json:"data"`
	}{Total: total, Data: addressInfo}

	if err := enc.Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		return
	}

	if flusher != nil {
		flusher.Flush()
	}
}
