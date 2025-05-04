package domain

type Geolocation struct {
	OriginAddress     string          `json:"origin_address" bson:"origin_address"`
	FormattedAddress  string          `json:"formatted_address" bson:"formatted_address"`
	Latitude          float64         `json:"latitude" bson:"latitude"`
	Longitude         float64         `json:"longitude" bson:"longitude"`
	Geocoder          string          `json:"geocoder" bson:"geocoder"`
	Status            StatusGeoResult `json:"status" bson:"-"`
	ResponseCoordsApi []interface{}   `json:"-" bson:"response_coors_api"`
}

type StatusGeoResult struct {
	Count  int  `json:"count"`
	Ok     int  `json:"ok"`
	Nok    int  `json:"nok"`
	Total  int  `json:"total"`
	Result bool `json:"result"`
}
