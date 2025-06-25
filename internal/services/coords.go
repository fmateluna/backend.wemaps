package services

import (
	"context"
	"errors"
	"strings"
	"wemaps/internal/domain"
	"wemaps/internal/infrastructure/geocoders"
	"wemaps/internal/ports"
)

type CoordsReportRequest struct {
	ReportName string              `json:"report_name"`
	Columns    []string            `json:"columns"`
	Values     map[string][]string `json:"values"`
}

type CoordsResponse struct {
	Address string  `json:"address"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	Origin  string  `json:"origin"`
	Format  string  `json:"format"`
}

type GeolocationService struct {
	geocoders  []geocoders.Geocoder
	repository ports.GeolocationRepository
}

func NewGeolocationService(repo ports.GeolocationRepository, portalRepo ports.PortalRepository) *GeolocationService {
	return &GeolocationService{
		geocoders: []geocoders.Geocoder{
			geocoders.NewWemapsGeocoder(portalRepo),
			geocoders.NewNominatimGeocoder(),
			geocoders.NewGoogleGeocoder(),
		},
		repository: repo,
	}
}

func (s *GeolocationService) GetCoordsFromAddress(address string) (domain.Geolocation, error) {
	formattedAddress := formatAddress(address)

	// Consultar en MongoDB primero
	result, exists, err := s.repository.Get(context.Background(), formattedAddress)
	if err != nil {
		return domain.Geolocation{}, err
	}
	if exists {
		return result, nil
	}

	// Si no está en MongoDB, consultar los geocodificadores
	for _, geocoder := range s.geocoders {
		addressCoords, err := geocoder.Geocode(formattedAddress)
		if err == nil && addressCoords != nil {
			addressCoords.OriginAddress = address
			// Guardar en MongoDB
			if err := s.repository.Save(context.Background(), formattedAddress, *addressCoords); err != nil {
				return domain.Geolocation{}, err
			}
			return *addressCoords, nil
		}
	}

	return domain.Geolocation{}, errors.New("no se pudo geolocalizar la dirección")
}

func formatAddress(address string) string {
	return strings.TrimSpace(strings.ToUpper(address))
}
