package ports

import (
	"context"
	"wemaps/internal/domain"
)

type GeolocationRepository interface {
	Save(ctx context.Context, address string, geolocation domain.Geolocation) error
	Get(ctx context.Context, address string) (domain.Geolocation, bool, error)
}
