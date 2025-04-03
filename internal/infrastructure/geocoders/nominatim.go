package geocoders

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"wemaps/internal/domain"
)

// ErrNotExact se usa cuando el resultado no cumple con los criterios de exactitud
var ErrNotExact = fmt.Errorf("resultado no exacto")

type NominatimGeocoder struct{}

func NewNominatimGeocoder() *NominatimGeocoder {
	return &NominatimGeocoder{}
}

func (n *NominatimGeocoder) Geocode(address string) (*domain.Geolocation, error) {
	params := url.Values{}
	params.Add("q", address)
	params.Add("format", "json")
	params.Add("limit", "1")

	req, err := http.NewRequest("GET", "https://nominatim.openstreetmap.org/search?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "WeMaps/1.0 (contacto@wemaps.com)") // Requerido por Nominatim

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("no se encontraron resultados")
	}

	result := data[0]

	category, ok := result["type"].(string)
	if !ok {
		return nil, fmt.Errorf("no se pudo determinar la categoría del resultado")
	}

	exactCategories := map[string]bool{
		"building": true,
		"place":    true,
	}

	if !exactCategories[category] {
		return nil, ErrNotExact
	}

	lat, err := strconv.ParseFloat(result["lat"].(string), 64)
	if err != nil {
		return nil, fmt.Errorf("error al parsear latitud: %v", err)
	}
	lon, err := strconv.ParseFloat(result["lon"].(string), 64)
	if err != nil {
		return nil, fmt.Errorf("error al parsear longitud: %v", err)
	}

	// Devolver el resultado exacto
	return &domain.Geolocation{
		FormattedAddress: result["display_name"].(string),
		Latitude:         lat,
		Longitude:        lon,
		Geocoder:         "nominatim",
	}, nil
}
