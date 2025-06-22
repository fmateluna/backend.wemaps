package services

import "time"

type Session struct {
	ID        string    `json:"id"`
	UserID    int       `json:"userId"`
	Token     string    `json:"token"`
	IPAddress string    `json:"ipAddress"`
	ExpiresAt time.Time `json:"expiresAt"`
}
