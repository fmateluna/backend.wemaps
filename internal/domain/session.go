package domain

import "time"

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
