package middleware

// Role adalah peran akun (dipindahkan verbatim dari model/models.go).
type Role string

const (
	RoleAdmin   Role = "admin"
	RoleCashier Role = "cashier"
)
