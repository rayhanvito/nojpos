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

// TransferListFilter adalah DTO filter daftar transfer stok (Unit D):
// outlet asal/tujuan dan rentang tanggal pembuatan (format YYYY-MM-DD,
// kosong berarti tanpa batas).
type TransferListFilter struct {
	FromOutletID uint
	ToOutletID   uint
	From         string
	To           string
}
