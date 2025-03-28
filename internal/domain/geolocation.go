package domain

type Geolocation struct {
	OriginAddress    string  `json:"origin_address"`
	FormattedAddress string  `json:"formatted_address"`
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
	Geocoder         string  `json:"geocoder"`
}
