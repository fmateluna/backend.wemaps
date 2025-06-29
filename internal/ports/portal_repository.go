package ports

import (
	"time"
	"wemaps/internal/adapters/http/dto"
	"wemaps/internal/infrastructure/repository"
)

type PortalRepository interface {
	GetUserID(alias string) (int, error)
	CreateUser(email, alias, fullName, phone, provider string) (int, error)
	LogSession(sessionID string, userID int, tokenString string, ipAddress string, expiresAt time.Time, active bool)
	FindUserByToken(token string) (*repository.User, error)
	FindUserByID(userID int) (*repository.User, error)
	SaveAddress(reportID int, address string, latitude float64, longitude float64, param5 string, geocoder string) (int, error)
	SaveReportColumnByIdReport(reportID int, addressID int, infoReport map[string]string, index int) (int, error)
	SaveReportByIdUser(idUser int, nameReport string, instance string) (int, error)
	SaveAddressInReport(reportID int, addressID int, latitude float64, longitude float64, formatAddress string, geocoder string) (int, error)
	GetAddressInfoByUserId(userID int) ([]dto.AddressReport, error)
	GetReportSummaryByUserId(userID int) ([]dto.ReportResume, error)
	GetReportByReportUserID(userID, reportID int) (dto.ReportResume, error)
	GetReportRowsByReportID(reportID int, page int, pageSize int) ([]dto.ReportRow, int, error)
	GetTotalReportsAndAddress(userID int) ([]dto.CategoryCount, error)
	GetAddressInfoByUserIdPeerPage(userID int, query string, limit, offset int) ([]dto.AddressReport, int, error)
	FindAddress(address string) (dto.WeMapsAddress, error)
	SetStatusReport(userID, reportID, status int) (dto.ReportResume, error)
}
