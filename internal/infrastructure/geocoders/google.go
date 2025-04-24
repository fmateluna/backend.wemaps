package geocoders

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"wemaps/internal/domain"
)

type GoogleGeocoder struct {
	apiKey string
}

func NewGoogleGeocoder() *GoogleGeocoder {
	return &GoogleGeocoder{apiKey: os.Getenv("GOOGLE_API_KEY")}
}

func (g *GoogleGeocoder) Geocode(address string) (*domain.Geolocation, error) {
	// Configurar los parámetros de la consulta
	params := url.Values{}
	params.Add("address", address)
	params.Add("key", g.apiKey)

	// Ejecutar la solicitud HTTP
	resp, err := http.Get("https://maps.googleapis.com/maps/api/geocode/json?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Decodificar la respuesta JSON
	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	// Verificar el estado de la respuesta
	if data["status"] != "OK" {
		return nil, fmt.Errorf("error en la respuesta de Google: %s", data["status"])
	}

	// Obtener los resultados
	results, ok := data["results"].([]interface{})
	if !ok || len(results) == 0 {
		return nil, fmt.Errorf("no se encontraron resultados")
	}

	// Filtrar resultados por location_type
	for _, res := range results {
		result := res.(map[string]interface{})
		geometry, ok := result["geometry"].(map[string]interface{})
		if !ok {
			continue
		}
		locationType, ok := geometry["location_type"].(string)
		if !ok {
			continue
		}

		// Verificar si el location_type es válido GEOMETRIC_CENTER si no tiene numero
		if locationType == "ROOFTOP" || locationType == "RANGE_INTERPOLATED" || locationType == "APPROXIMATE" {
			location, ok := geometry["location"].(map[string]interface{})
			if !ok {
				continue
			}
			lat, ok := location["lat"].(float64)
			if !ok {
				continue
			}
			lng, ok := location["lng"].(float64)
			if !ok {
				continue
			}

			// Devolver el primer resultado válido
			return &domain.Geolocation{
				FormattedAddress: result["formatted_address"].(string),
				Latitude:         lat,
				Longitude:        lng,
				Geocoder:         "google",
			}, nil
		}
	}

	// Si no hay resultados con location_type válido
	return nil, fmt.Errorf("no se encontraron resultados con location_type válido (ROOFTOP o RANGE_INTERPOLATED)")
}
