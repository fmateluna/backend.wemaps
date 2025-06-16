package services

import "time"

type User struct {
	ID       int
	Email    string
	Alias    string
	FullName string
	Phone    string
}

type Session struct {
	ID        string    `json:"id"`
	UserID    int       `json:"userId"`
	Token     string    `json:"token"`
	IPAddress string    `json:"ipAddress"`
	ExpiresAt time.Time `json:"expiresAt"`
}


