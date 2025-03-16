package ports

import "wemaps/internal/domain"

type GeolocationService interface {
	GetCoordsFromAddress(address string) (domain.Geolocation, error)
}
