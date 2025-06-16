package services

import "wemaps/internal/domain"

type Health struct{}

func NewHealthService() *Health {
	return &Health{}
}

func (h *Health) Check(alias string ) (domain.HealthStatus, error) {
	return domain.HealthStatus{Status: "ok " + alias}, nil
}
