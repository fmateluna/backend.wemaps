package ports

type PortalRepository interface {
	GetUserID(alias string) (int, error)
	CreateUser(email, alias, fullName, phone string) (int, error)
}
