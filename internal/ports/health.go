package ports

import "wemaps/internal/domain"

type HealthService interface {
	Check() (domain.HealthStatus, error)
}
