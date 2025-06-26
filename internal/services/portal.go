package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"wemaps/internal/adapters/http/dto"
	"wemaps/internal/domain"
	"wemaps/internal/ports"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// PortalService estructura sin cache manager inyectado
type PortalService struct {
	repository ports.PortalRepository
	cache      map[string]int
	cacheMu    sync.RWMutex
	jwtSecret  string
}

// NewPortalService crea un nuevo PortalService
func NewPortalService(repository ports.PortalRepository) *PortalService {
	return &PortalService{
		repository: repository,
		cache:      make(map[string]int),
	}
}

// InvalidateUserCache permite invalidar el caché para un usuario específico
func (s *PortalService) InvalidateUserCache(userID int) {
	cacheMgr := GetCacheManager()
	cacheMgr.Delete(cacheMgr.addressCacheKey(userID))
	cacheMgr.Delete(cacheMgr.reportSummaryCacheKey(userID))
}

func (s *PortalService) SaveReportInfo(idUser int, reportID int, nameReport string, infoReport map[string]string, geo domain.Geolocation, hash string, index int) (int, error) {

	var err error

	if reportID == -1 {
		reportID, err = s.repository.SaveReportByIdUser(idUser, nameReport, hash)
		if err != nil {
			return reportID, fmt.Errorf("failed to save report: %v", err)
		}
	}
	s.InvalidateUserCache(idUser)

	return reportID, s.saveReportDetails(reportID, infoReport, geo, index)
}

func (s *PortalService) SaveReportInfoCache(idUser int, nameReport string, infoReport map[string]string, geo domain.Geolocation, hash string, index int) (int, error) {
	cacheMgr := GetCacheManager()
	cacheReportKey := cacheMgr.cacheKey(nameReport, hash)
	var reportID int
	var found bool

	s.cacheMu.RLock()
	reportID, found = s.cache[cacheReportKey]
	s.cacheMu.RUnlock()

	if !found {
		var err error
		reportID, err = s.repository.SaveReportByIdUser(idUser, nameReport, hash)
		if err != nil {
			return -reportID, fmt.Errorf("failed to save report: %v", err)
		}

		s.cacheMu.Lock()
		s.cache[cacheReportKey] = reportID
		s.cacheMu.Unlock()

		s.InvalidateUserCache(idUser)
	}

	return reportID, s.saveReportDetails(reportID, infoReport, geo, index)
}

func (s *PortalService) saveReportDetails(reportID int, infoReport map[string]string, geo domain.Geolocation, index int) error {
	addressID := 0
	var err error

	if geo.Latitude != 0 && geo.Longitude != 0 {
		addressID, err = s.repository.SaveAddress(
			reportID, geo.OriginAddress, geo.Latitude, geo.Longitude, geo.FormattedAddress, geo.Geocoder,
		)
		if err != nil {
			return fmt.Errorf("failed to save address: %v", err)
		}
	}

	if addressID > 0 {
		_, err = s.repository.SaveAddressInReport(reportID, addressID, geo.Latitude, geo.Longitude, geo.FormattedAddress, geo.Geocoder)
		if err != nil {
			return fmt.Errorf("error linking new address to report: %v", err)
		}
	}
	go func() {
		s.repository.SaveReportColumnByIdReport(reportID, addressID, infoReport, index)
	}()

	return err
}

func (s *PortalService) ValidateToken(token string) (*dto.UserPortal, error) {
	repoUser, err := s.repository.FindUserByToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %v", err)
	}
	if repoUser == nil {
		return nil, nil
	}
	user := &dto.UserPortal{
		ID:       repoUser.ID,
		Alias:    repoUser.Alias,
		Email:    repoUser.Email,
		FullName: repoUser.FullName,
		Phone:    repoUser.Phone,
	}
	return user, nil
}

func (s *PortalService) CreateUser(alias string, email string, name string, phone string) (int, error) {
	id, err := s.repository.CreateUser(alias, email, name, phone)
	if err != nil {
		fmt.Println("Error al crear el usuario:", err)
		return -1, err
	}
	fmt.Println("Nuevo usuario creado:", alias)
	return id, err
}

func (s *PortalService) IdentificoTipoLogIn(ctx context.Context, request dto.RequestLogin) (*dto.UserPortal, error) {
	var user dto.UserPortal

	if request.Provider == "google.com" {
		var fmtinGoogle dto.LoginGoogle

		if responseBytes, ok := request.Response.(string); ok {
			err := json.Unmarshal([]byte(responseBytes), &fmtinGoogle)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal Google fmtin response: %w", err)
			}
		} else if responseMap, ok := request.Response.(map[string]interface{}); ok {
			bytes, err := json.Marshal(responseMap)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response map: %w", err)
			}
			err = json.Unmarshal(bytes, &fmtinGoogle)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal response to fmtinGoogle: %w", err)
			}
		} else {
			fmtinGoogle, ok = request.Response.(dto.LoginGoogle)
			if !ok {
				return nil, errors.New("invalid response type for Google fmtin")
			}
		}

		user.Alias = fmtinGoogle.User.Email
		user.Email = fmtinGoogle.User.Email
		user.FullName = fmtinGoogle.User.DisplayName
		user.Phone = ""
		if len(fmtinGoogle.User.ProviderData) > 0 && fmtinGoogle.User.ProviderData[0].PhoneNumber != nil {
			if phone, ok := fmtinGoogle.User.ProviderData[0].PhoneNumber.(string); ok {
				user.Phone = phone
			}
		}
	}

	return &user, nil
}

func (s *PortalService) RecordSession(userID int, ipAddress string, active bool) (*Session, error) {
	sessionID := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Minute * 5)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    userID,
		"session_id": sessionID,
		"ip_address": ipAddress,
		"exp":        expiresAt.Unix(),
	})

	secretKey := "PONTUPIN"

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return nil, errors.New("failed to sign token")
	}

	s.repository.LogSession(sessionID, userID, tokenString, ipAddress, expiresAt, active)

	return &Session{
		ID:        sessionID,
		UserID:    userID,
		Token:     tokenString,
		IPAddress: ipAddress,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *PortalService) GetUserID(Alias string) (int, error) {
	userID, err := s.repository.GetUserID(Alias)
	return userID, err
}

func (s *PortalService) GetAddressInfoByUserId(userID int) ([]dto.AddressReport, error) {

	cacheMgr := GetCacheManager()
	cacheKey := cacheMgr.addressCacheKey(userID)

	if cachedData, found := cacheMgr.Get(cacheKey); found {
		if addressInfo, ok := cachedData.([]dto.AddressReport); ok {
			go func() {
				addressInfoUpdate, err := s.repository.GetAddressInfoByUserId(userID)
				if err == nil {

					if !cacheMgr.Update(cacheKey, addressInfoUpdate, time.Minute*5) {
						cacheMgr.Set(cacheKey, addressInfoUpdate, time.Minute*5)
					}
				}
			}()
			return addressInfo, nil
		}
	}

	addressInfo, err := s.repository.GetAddressInfoByUserId(userID)
	if err != nil {
		return nil, err
	}

	if !cacheMgr.Update(cacheKey, addressInfo, time.Minute*5) {
		cacheMgr.Set(cacheKey, addressInfo, time.Minute*5)
	}

	return addressInfo, nil
}

func (s *PortalService) GetReportSummaryByUserId(userID int) ([]dto.ReportResume, error) {
	cacheMgr := GetCacheManager()
	cacheKey := cacheMgr.reportSummaryCacheKey(userID)

	if cachedData, found := cacheMgr.Get(cacheKey); found {
		if summaries, ok := cachedData.([]dto.ReportResume); ok {
			go func() {
				summariesUpdate, err := s.repository.GetReportSummaryByUserId(userID)
				if err == nil {

					if !cacheMgr.Update(cacheKey, summariesUpdate, time.Minute*5) {
						cacheMgr.Set(cacheKey, summariesUpdate, time.Minute*5)
					}
				}
			}()
			return summaries, nil
		}
	}

	summaries, err := s.repository.GetReportSummaryByUserId(userID)
	if err != nil {
		return nil, err
	}

	if !cacheMgr.Update(cacheKey, summaries, time.Minute*5) {
		cacheMgr.Set(cacheKey, summaries, time.Minute*5)
	}

	return summaries, nil
}

func (s *PortalService) GetReportRowsByReportID(reportID int, page int, pageSize int) ([]dto.ReportRow, int, error) {
	rows, totalRows, err := s.repository.GetReportRowsByReportID(reportID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return rows, totalRows, nil
}
func (s *PortalService) GetTotalReportsAndAddress(userID int) ([]dto.CategoryCount, error) {
	totales, err := s.repository.GetTotalReportsAndAddress(userID)
	return totales, err
}

func (s *PortalService) SetStatusReport(userID, reportID, status int) (dto.ReportResume, error) {
	report, err := s.repository.SetStatusReport(userID, reportID, status)
	return report, err
}

func (s *PortalService) GetAddressInfoByUserIdPeerPage(userID int, query string, page, pageSize int) ([]dto.AddressReport, int, error) {
	offset := (page - 1) * pageSize
	addresses, total, err := s.repository.GetAddressInfoByUserIdPeerPage(userID, query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	return addresses, total, nil
}

func (s *PortalService) GetReportByReportUserID(userID, reportID int) (dto.ReportResume, error) {
	report, err := s.repository.GetReportByReportUserID(userID, reportID)
	return report, err
}

func (s *PortalService) FindAddreessWemaps(address string) (dto.WeMapsAddress, error) {
	report, err := s.repository.FindAddress(address)
	return report, err
}

func (s *PortalService) GenerateToken(userID int) (string, error) {

	user, err := s.repository.FindUserByID(userID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch user alias: %v", err)
	}
	if user.Alias == "" {
		return "", fmt.Errorf("user alias is empty")
	}

	claims := &dto.Claims{
		UserAlias: user.Alias,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Minute * 5)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}

	return tokenString, nil
}

func (s *PortalService) ValidateTokenAPI(tokenString string) (*dto.Claims, error) {
	claims := &dto.Claims{}

	s.jwtSecret = "PONTUPINWEMAPS"

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}
	return claims, nil
}
