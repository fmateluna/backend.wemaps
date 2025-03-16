package geocoders

import "wemaps/internal/domain"

type Geocoder interface {
	Geocode(address string) (*domain.Geolocation, error)
}
