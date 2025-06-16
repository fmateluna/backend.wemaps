package ports

import (
	"time"
	"wemaps/internal/adapters/http/dto"
	"wemaps/internal/infrastructure/repository"
)

type PortalRepository interface {
	GetUserID(alias string) (int, error)
	CreateUser(email, alias, fullName, phone string) (int, error)
	LogSession(sessionID string, userID int, tokenString string, ipAddress string, expiresAt time.Time, active bool)
	FindUserByToken(token string) (*repository.User, error)
	SaveAddress(idReport int, address string, latitude float64, longitude float64, param5 string, geocoder string) (int, error)
	SaveReportColumnByIdReport(idReport int, addressID int, infoReport map[string]string, index int) (int, error)
	SaveReportByIdUser(idUser int, nameReport string, instance string) (int, error)
	SaveAddressInReport(idReport int, addressID int, latitude float64, longitude float64, formatAddress string, geocoder string) (int, error)
	GetAddressInfoByUserId(userID int) ([]dto.AddressReport, error)
	GetReportSummaryByUserId(userID int) ([]dto.ReportResume, error)
}
