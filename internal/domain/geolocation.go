package domain

type Geolocation struct {
	FormattedAddress string  `json:"formatted_address"`
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
	Geocoder         string  `json:"geocoder"` // Indica qué geocodificador lo resolvió (google, nominatim, etc.)
}
