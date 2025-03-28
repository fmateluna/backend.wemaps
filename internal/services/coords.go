package services

import (
	"errors"
	"strings"
	"wemaps/internal/domain"
	"wemaps/internal/infrastructure/geocoders"
)

// CoordsRequest ya debe existir en tu código, lo reutilizamos
type CoordsRequest struct {
	Columns []string            `json:"columns"`
	Values  map[string][]string `json:"values"`
}

type CoordsResponse struct {
	Address string  `json:"address"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	Origin  string  `json:"origin"`
	Format  string  `json:"format"`
}

type GeolocationService struct {
	geocoders []geocoders.Geocoder
}

func NewGeolocationService() *GeolocationService {
	return &GeolocationService{
		//Aca el orden de los geocoders indica la prioridad
		geocoders: []geocoders.Geocoder{
			geocoders.NewNominatimGeocoder(),
			geocoders.NewGoogleGeocoder(),
			// aquí (Here, TomTom, etc.)
		},
	}
}

func (s *GeolocationService) GetCoordsFromAddress(address string) (domain.Geolocation, error) {
	formattedAddress := formatAddress(address)

	for _, geocoder := range s.geocoders {
		result, err := geocoder.Geocode(formattedAddress)
		if err == nil && result != nil {
			return *result, nil
		}
	}
	return domain.Geolocation{}, errors.New("no se pudo geolocalizar la dirección")
}

func formatAddress(address string) string {
	return strings.TrimSpace(strings.ToUpper(address))
}
