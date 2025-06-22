package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"wemaps/internal/adapters/http/dto"
)

func (s *Server) logInHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	//VALIDO REQUEST
	request := dto.RequestLogin{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	//SETEO USUARIO ONLINE
	user, err := s.portalService.IdentificoTipoLogIn(r.Context(), request)
	if err != nil {
		http.Error(w, fmt.Sprintf("No se pudo identificar el provedor de login: %v", err), http.StatusUnauthorized)
		return
	}

	id, errorNoID := s.portalService.GetUserID(user.Alias)

	if errorNoID != nil || id == -1 {
		//Usuario no existe, lo creo
		id, err = s.portalService.CreateUser(user.Alias, user.Email, user.FullName, user.Phone)
	}

	ipAddress := r.RemoteAddr
	session, err := s.portalService.RecordSession(id, ipAddress, true)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create session: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

func (s *Server) logOutHandler(w http.ResponseWriter, r *http.Request) {
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

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ipAddress := r.RemoteAddr
	// Invalidar la sesión del usuario
	_, err = s.portalService.RecordSession(user.ID, ipAddress, false)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create session: %v", err), http.StatusInternalServerError)
		return
	}

	// Respuesta exitosa
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logout successful",
	})
}
