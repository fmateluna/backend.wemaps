package geocoders

import (
	"fmt"
	"wemaps/internal/domain"
	"wemaps/internal/ports"
)

type WemapsGeocoder struct {
	repo ports.PortalRepository
}

func NewWemapsGeocoder(repo ports.PortalRepository) *WemapsGeocoder {
	return &WemapsGeocoder{repo: repo}
}

func (w *WemapsGeocoder) Geocode(address string) (*domain.Geolocation, error) {
	// Consultar la dirección en la base de datos local
	geo, err := w.repo.FindAddress(address)
	if err != nil {
		return nil, fmt.Errorf("error al consultar la dirección en Wemaps: %v", err)
	}

	// Verificar si se obtuvo un resultado válido
	if geo.FormattedAddress == "" {
		return nil, fmt.Errorf("no se encontraron resultados para la dirección: %s", address)
	}

	// Mapear el resultado a domain.Geolocation
	return &domain.Geolocation{
		FormattedAddress: geo.FormattedAddress,
		Latitude:         geo.Latitude,
		Longitude:        geo.Longitude,
		Geocoder:         "wemaps",
		ResponseCoordsApi: []interface{}{map[string]interface{}{
			"formatted_address": geo.FormattedAddress,
			"latitude":          geo.Latitude,
			"longitude":         geo.Longitude,
		}},
	}, nil
}
