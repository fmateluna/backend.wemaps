package services

import (
	"errors"
	"strings"
	"sync"
	"time"
	"wemaps/internal/domain"
	"wemaps/internal/infrastructure/geocoders"

	lru "github.com/hashicorp/golang-lru"
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

type cacheEntry struct {
	geolocation domain.Geolocation
	timestamp   time.Time
}

type GeolocationService struct {
	geocoders []geocoders.Geocoder
	cache     *lru.Cache
	ttl       time.Duration
	mutex     sync.RWMutex
}

func NewGeolocationService() *GeolocationService {
	cache, _ := lru.New(1000) // Máximo de 1000 entradas

	return &GeolocationService{
		geocoders: []geocoders.Geocoder{
			geocoders.NewNominatimGeocoder(),
			geocoders.NewGoogleGeocoder(),
		},
		cache: cache,
		ttl:   30 * 24 * time.Hour, // 30 días
	}
}

func (s *GeolocationService) GetCoordsFromAddress(address string) (domain.Geolocation, error) {
	formattedAddress := formatAddress(address)

	// Primero verificamos la caché
	if result, exist := s.getFromCache(formattedAddress); exist {
		return result, nil
	}

	// Si no está en caché, intentamos con los geocodificadores
	for _, geocoder := range s.geocoders {
		result, err := geocoder.Geocode(formattedAddress)

		if err == nil && result != nil {
			s.saveToCache(formattedAddress, *result)
			return *result, nil
		}
	}

	return domain.Geolocation{}, errors.New("no se pudo geolocalizar la dirección")
}

func (s *GeolocationService) saveToCache(formattedAddress string, result domain.Geolocation) {
	entry := cacheEntry{
		geolocation: result,
		timestamp:   time.Now(),
	}
	s.cache.Add(formattedAddress, entry)
}

func (s *GeolocationService) getFromCache(formattedAddress string) (domain.Geolocation, bool) {
	value, exists := s.cache.Get(formattedAddress)
	if !exists {
		return domain.Geolocation{}, false
	}

	entry := value.(cacheEntry)

	// Si la entrada ha expirado, no la devolvemos
	if time.Since(entry.timestamp) >= s.ttl {
		s.cache.Remove(formattedAddress) // Eliminamos la entrada
		return domain.Geolocation{}, false
	}

	return entry.geolocation, true
}

func formatAddress(address string) string {
	return strings.TrimSpace(strings.ToUpper(address))
}
