package notification

// NotificationPage adalah halaman hasil List (dipindah verbatim dari
// service/notification.go, BE-F9a).
type NotificationPage struct {
	Items []*Notification `json:"items"`
	Total int             `json:"total"`
	Page  int             `json:"page"`
	Limit int             `json:"limit"`
}
