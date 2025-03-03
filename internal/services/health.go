package services

import "wemaps/internal/domain"

type Health struct{}

func NewHealthService() *Health {
	return &Health{}
}

func (h *Health) Check() (domain.HealthStatus, error) {
	return domain.HealthStatus{Status: "ok"}, nil
}
