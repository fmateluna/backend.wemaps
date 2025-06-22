package dto

import "time"

type RequestLogin struct {
	Provider string      `json:"provider"`
	Token    string      `json:"token"`
	Response interface{} `json:"response"`
}

type AddressReport struct {
	ID                    string          `json:"id"`
	Address               string          `json:"address"`
	NormalizedAddress     string          `json:"normalized_address"`
	Latitude              float64         `json:"latitude"`
	Longitude             float64         `json:"longitude"`
	AtributosRelacionados []ReportDetail  `json:"atributos_relacionados"`
	Reportes              []ReportSummary `json:"reportes"`
}

type ReportResume struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"created_at"`
	Direcciones int       `json:"direcciones"`
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

type ReportRow struct {
	IndexColumn     int               `json:"index_column"`
	FilaTranspuesta map[string]string `json:"fila_transpuesta"`
}

type CategoryCount struct {
	Category string `json:"category"`
	Total    int    `json:"total"`
}

type UserPortal struct {
	ID       int
	Email    string
	Alias    string
	FullName string
	Phone    string
}
