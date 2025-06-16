package dto

type RequestLogin struct {
	Provider string      `json:"provider"`
	Token    string      `json:"token"`
	Response interface{} `json:"response"`
}

type AddressReport struct {
	Address               string          `json:"address"`
	NormalizedAddress     string          `json:"normalized_address"`
	Latitude              float64         `json:"latitude"`
	Longitude             float64         `json:"longitude"`
	AtributosRelacionados []ReportDetail  `json:"atributos_relacionados"`
	Reportes              []ReportSummary `json:"reportes"`
}

type ReportDetail struct {
	ReportID   int               `json:"report_id"`
	ReportName string            `json:"report_name"`
	Atributos  map[string]string `json:"atributos"`
}

type ReportSummary struct {
	ReportID   int    `json:"report_id"`
	ReportName string `json:"report_name"`
}