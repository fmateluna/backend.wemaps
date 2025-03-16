package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"wemaps/internal/services"
)

func (s *Server) getCoordsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var report services.CoordsRequest
	err := json.NewDecoder(r.Body).Decode(&report)
	if err != nil {
		http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
		return
	}

	geolocationService := services.NewGeolocationService()

	if len(report.Columns) > 0 {
		key := report.Columns[0]
		values := report.Values[key]

		for i, address := range values {

			fmt.Printf("%d: %s\n", i+1, address)

			geo, err := geolocationService.GetCoordsFromAddress(address)
			if err != nil {
				fmt.Printf("Error geolocalizando %s: %v\n", address, err)
				continue
			}

			fmt.Printf("Resultado: %s, Lat: %f, Lon: %f, Geocoder: %s\n",
				geo.FormattedAddress, geo.Latitude, geo.Longitude, geo.Geocoder)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Reporte recibido"})
}
