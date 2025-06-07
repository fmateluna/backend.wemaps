package services

import (
	"context"
	"errors"
	"time"
	"wemaps/internal/adapters/http/dto"
	"wemaps/internal/ports"

	"github.com/google/uuid"
)

type PortalService struct {
	repository ports.PortalRepository
}

type User struct {
	ID       int
	Email    string
	Alias    string
	FullName string
	Phone    string
}

type Session struct {
	ID        string
	UserID    int
	Token     string
	IPAddress string
	ExpiresAt time.Time
}

func NewPortalService(repositoryPortal ports.PortalRepository) *PortalService {
	return &PortalService{
		repository: repositoryPortal,
	}
}

func (s *PortalService) SeteaUserOnline(ctx context.Context, request dto.RequestLogin) (*User, error) {
	var user User

	if request.Provider == "google.com" {
		loginGoogle, ok := request.Response.(dto.LoginGoogle)
		if !ok {
			return nil, errors.New("invalid response type for Google login")
		}
		user.Alias = loginGoogle.User.Email
		user.Email = loginGoogle.User.Email
		user.FullName = loginGoogle.User.DisplayName
		if len(loginGoogle.User.ProviderData) > 0 && loginGoogle.User.ProviderData[0].PhoneNumber != nil {
			phone, ok := loginGoogle.User.ProviderData[0].PhoneNumber.(string)
			if !ok {
				user.Phone = ""
			}
			user.Phone = phone
		} else {
			user.Phone = ""
		}

	}

	return &user, nil
}

func (s *PortalService) CreateSession(userID int, ipAddress string) (*Session, error) {
	sessionID := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)
	/*
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id":    userID,
			"session_id": sessionID,
			"ip_address": ipAddress,
			"exp":        expiresAt.Unix(),
		})
	*/

	tokenString := ""
	/*
		query := `
			INSERT INTO sessions (user_id, session_id, token, ip_address, expires_at)
			VALUES ($1, $2, $3, $4, $5)`
		_, err = s.repository.Exec(query, userID, sessionID, tokenString, ipAddress, expiresAt)
		if err != nil {
			return nil, fmt.Errorf("error saving session: %v", err)
		}
	*/
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

func (s *PortalService) ValidateSession(tokenString, ipAddress string) (*Session, error) {

	/*
			token , err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(s.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			return nil, fmt.Errorf("invalid token: %v", err)
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, fmt.Errorf("invalid token claims")
		}

		sessionID, _ := claims["session_id"].(string)
		//userID, _ := claims["user_id"].(float64)
		storedIPAddress, _ := claims["ip_address"].(string)
		exp, _ := claims["exp"].(float64)

		if storedIPAddress != ipAddress {
			return nil, fmt.Errorf("IP address mismatch")
		}

		if time.Now().Unix() > int64(exp) {
			return nil, fmt.Errorf("token expired")
		}

		var session Session
		query := `
			SELECT user_id, session_id, token, ip_address, expires_at
			FROM sessions
			WHERE session_id = $1 AND is_active = TRUE`
		err = s.repository.QueryRow(query, sessionID).Scan(&session.UserID, &session.ID, &session.Token, &session.IPAddress, &session.ExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("session not found or inactive: %v", err)
		}
	*/
	var session Session
	return &session, nil
}
