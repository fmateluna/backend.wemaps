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
	params := url.Values{}
	params.Add("address", address+", Chile")
	params.Add("key", g.apiKey)

	resp, err := http.Get("https://maps.googleapis.com/maps/api/geocode/json?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	if data["status"] != "OK" {
		return nil, fmt.Errorf("error en la respuesta de Google: %s", data["status"])
	}

	results, ok := data["results"].([]interface{})
	if !ok || len(results) == 0 {
		return nil, fmt.Errorf("no se encontraron resultados")
	}

	result := results[0].(map[string]interface{})
	geometry := result["geometry"].(map[string]interface{})
	location := geometry["location"].(map[string]interface{})

	return &domain.Geolocation{
		FormattedAddress: result["formatted_address"].(string),
		Latitude:         location["lat"].(float64),
		Longitude:        location["lng"].(float64),
		Geocoder:         "google",
	}, nil
}
