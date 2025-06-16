package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
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
	status, err := s.healthService.Check(user.Alias)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
