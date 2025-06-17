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

type PortalService struct {
	repository ports.PortalRepository
	cache      map[string]int
	cacheMu    sync.RWMutex
}

func NewPortalService(repository ports.PortalRepository) *PortalService {
	return &PortalService{
		repository: repository,
		cache:      make(map[string]int),
	}
}

func (s *PortalService) cacheKey(nameReport, hash string) string {
	return fmt.Sprintf("%s:%s", nameReport, hash)
}

func (s *PortalService) SaveReportInfo(idUser int, nameReport string, infoReport map[string]string, geo domain.Geolocation, hash string, index int) error {
	cacheReportKey := s.cacheKey(nameReport, hash)
	var idReport int
	var found bool

	s.cacheMu.RLock()
	idReport, found = s.cache[cacheReportKey]
	s.cacheMu.RUnlock()

	if !found {
		var err error
		idReport, err = s.repository.SaveReportByIdUser(idUser, nameReport, hash)
		if err != nil {
			return fmt.Errorf("failed to save report: %v", err)
		}

		s.cacheMu.Lock()
		s.cache[cacheReportKey] = idReport
		s.cacheMu.Unlock()
	}

	// Guardar detalles
	return s.saveReportDetails(idReport, infoReport, geo, index)
}

func (s *PortalService) saveReportDetails(idReport int, infoReport map[string]string, geo domain.Geolocation, index int) error {
	addressID := 0
	cacheAddressKey := s.cacheKey(geo.OriginAddress, geo.FormattedAddress)
	var err error

	s.cacheMu.RLock()
	var found bool
	addressID, found = s.cache[cacheAddressKey]
	s.cacheMu.RUnlock()

	if !found {

		addressID, err = s.repository.SaveAddress(
			idReport, geo.OriginAddress, geo.Latitude, geo.Longitude, geo.FormattedAddress, geo.Geocoder,
		)
		if err != nil {
			return fmt.Errorf("failed to save address: %v", err)
		}

		s.cacheMu.Lock()
		s.cache[cacheAddressKey] = addressID
		s.cacheMu.Unlock()
	}

	if addressID > 0 {
		_, err = s.repository.SaveAddressInReport(idReport, addressID, geo.Latitude, geo.Longitude, geo.FormattedAddress, geo.Geocoder)
		if err != nil {
			return fmt.Errorf("error linking new address to report: %v", err)
		}
	}

	_, err = s.repository.SaveReportColumnByIdReport(idReport, addressID, infoReport, index)
	if err != nil {
		return fmt.Errorf("failed to save report columns: %v", err)
	}

	return nil
}

func (s *PortalService) ValidateToken(token string) (*User, error) {
	repoUser, err := s.repository.FindUserByToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %v", err)
	}
	if repoUser == nil {
		return nil, nil
	}
	user := &User{
		ID:       repoUser.ID,
		Alias:    repoUser.Alias,
		Email:    repoUser.Email,
		FullName: repoUser.FullName,
		Phone:    repoUser.Phone,
	}
	return user, nil
}

func (s *PortalService) CreateUser(alias string, email string, name string, phone string) (int, error) {
	id, error := s.repository.CreateUser(alias, email, name, phone)
	if error != nil {
		fmt.Println("Error al crear el usuario:", error)
		return -1, error
	} else {
		fmt.Println("Nuevo usuario creado:", alias)
		return id, error
	}
}

func (s *PortalService) IdentificoTipoLogIn(ctx context.Context, request dto.RequestLogin) (*User, error) {
	var user User

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
			// Intenta castear directamente a dto.fmtinGoogle
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
	//Sesion dura 24 horas
	expiresAt := time.Now().Add(24 * time.Hour)

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

	userID, _ := s.repository.GetUserID(Alias)
	return userID, nil
}

func (s *PortalService) GetAddressInfoByUserId(userID int) ([]dto.AddressReport, error) {
	addressInfo, err := s.repository.GetAddressInfoByUserId(userID)
	return addressInfo, err
}

func (s *PortalService) GetReportSummaryByUserId(userID int) ([]dto.ReportResume, error) {
	summaries, err := s.repository.GetReportSummaryByUserId(userID)
	return summaries, err
}

func (s *PortalService) GetReportRowsByReportID(reportID int) ([]dto.ReportRow, error) {
	rows, err := s.repository.GetReportRowsByReportID(reportID)
	return rows, err
}
