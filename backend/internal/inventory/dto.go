package inventory

// MovementListFilter adalah DTO filter daftar pergerakan stok di layer
// service; handler tidak lagi membangun repo.MovementFilter secara langsung.
type MovementListFilter struct {
	Type      string
	ProductID uint
	Q         string
	Page      int
	Limit     int
}
