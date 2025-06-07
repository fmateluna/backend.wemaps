package http

import (
	"encoding/json"
	"net/http"
	"wemaps/internal/adapters/http/dto"
)

func (s *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	request := dto.RequestLogin{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	/*
		user, err := s.portalService.SeteaUserOnline(r.Context(), request)
		if err != nil {
			http.Error(w, fmt.Sprintf("Authentication failed: %v", err), http.StatusUnauthorized)
			return
		}

		s.portalService.GetUserID(user.Alias)

		ipAddress := r.RemoteAddr
		session, err := s.portalService.CreateSession(user.ID, ipAddress)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create session: %v", err), http.StatusInternalServerError)
			return
		}*/

	response := struct {
		Token string `json:"token"`
	}{
		Token: request.Token,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
