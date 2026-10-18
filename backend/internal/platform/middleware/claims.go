package middleware

// Claims adalah isi token akses yang terverifikasi (dipindahkan verbatim
// dari service/auth.go; field Role kini bertipe Role di paket ini).
type Claims struct {
	UserID            uint
	StoreID           uint
	OutletID          uint
	Name              string
	Email             string
	ActingAsCashierID *uint
	Role              Role
	APIKey            bool
}

func (c *Claims) ActiveRole() Role {
	if c.ActingAsCashierID != nil {
		return RoleCashier
	}
	return RoleAdmin
}
