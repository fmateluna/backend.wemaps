package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"wemaps/internal/adapters/http/dto"
)

func (s *Server) getSingleAddressCoordsHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "Missing address query parameter", http.StatusBadRequest)
		return
	}

	// Sanitize address
	address = sanitizeString(address)

	// Check local database for similar address
	var geo dto.WeMapsAddress
	geo, err := s.portalService.FindAddreessWemaps(address)

	if err == nil {
		// Assume FindAddreessWemaps includes similarity check (e.g., >0.7)
		response := struct {
			FormattedAddress string  `json:"formatted_address"`
			Latitude         float64 `json:"latitude"`
			Longitude        float64 `json:"longitude"`
		}{
			FormattedAddress: geo.FormattedAddress,
			Latitude:         geo.Latitude,
			Longitude:        geo.Longitude,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
		return // Stop execution after sending response
	}

	// Log database error but proceed to external geocoder
	if err != nil {
		fmt.Printf("Database query error: %v\n", err)
	}

	// Fallback to external geocoder
	geoFromCoords, err := s.coordService.GetCoordsFromAddress(address)
	if err != nil {
		// Handle external geocoder error
		response := dto.WeMapsAddress{
			FormattedAddress: "No se pudo geolocalizar",
			Latitude:         0,
			Longitude:        0,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
		return
	}

	// Send external geocoder response
	response := struct {
		FormattedAddress string  `json:"formatted_address"`
		Latitude         float64 `json:"latitude"`
		Longitude        float64 `json:"longitude"`
	}{
		FormattedAddress: geoFromCoords.FormattedAddress,
		Latitude:         geoFromCoords.Latitude,
		Longitude:        geoFromCoords.Longitude,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *Server) getTokenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	userID, err := s.portalService.ValidateToken(token)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid session token: %v", err), http.StatusUnauthorized)
		return
	}

	jwtToken, err := s.portalService.GenerateToken(userID.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate JWT token: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"token": jwtToken}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
