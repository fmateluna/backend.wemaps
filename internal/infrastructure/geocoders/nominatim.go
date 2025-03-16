package geocoders

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"wemaps/internal/domain"
)

type NominatimGeocoder struct{}

func NewNominatimGeocoder() *NominatimGeocoder {
	return &NominatimGeocoder{}
}

func (n *NominatimGeocoder) Geocode(address string) (*domain.Geolocation, error) {
	params := url.Values{}
	params.Add("q", address+", Chile")
	params.Add("format", "json")
	params.Add("limit", "1")

	req, err := http.NewRequest("GET", "https://nominatim.openstreetmap.org/search?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "WeMaps/1.0 (contacto@wemaps.com)") // Nominatim requiere un User-Agent

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
	lat, _ := strconv.ParseFloat(result["lat"].(string), 64)
	lon, _ := strconv.ParseFloat(result["lon"].(string), 64)

	return &domain.Geolocation{
		FormattedAddress: result["display_name"].(string),
		Latitude:         lat,
		Longitude:        lon,
		Geocoder:         "nominatim",
	}, nil
}
