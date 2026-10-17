package activity

import (
	"context"
)

// Repository adalah kontrak penyimpanan aktivitas yang dipakai
// Service (ActivityRepository lama di service/activity.go, bertipe
// slice).
type Repository interface {
	Create(ctx context.Context, storeID uint, actorID, actorName, action, detail, refType, refID string) error
	List(ctx context.Context, storeID uint, f ActivityFilter) ([]ActivityLog, int64, error)
}

// Konstanta aksi aktivitas, dipindahkan verbatim dari
// service/activity.go (BE-F9b).
const (
	ActivityLogin           = "LOGIN"
	ActivityLogout          = "LOGOUT"
	ActivitySwitch          = "SWITCH"
	ActivityUserCreated     = "USER_CREATED"
	ActivityProfileUpdated  = "PROFILE_UPDATED"
	ActivityAccountDisabled = "ACCOUNT_DISABLED"
	ActivityAccountEnabled  = "ACCOUNT_ENABLED"
	ActivityPasscodeChanged = "PASSCODE_CHANGED"
)

// Service adalah ActivityService lama di service/activity.go, dipindah
// verbatim.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Log(ctx context.Context, storeID uint, actorID, actorName, action, detail, refType, refID string) {
	if s.repo == nil {
		return
	}
	_ = s.repo.Create(ctx, storeID, actorID, actorName, action, detail, refType, refID)
}

func (s *Service) List(ctx context.Context, storeID uint, f ActivityFilter) ([]ActivityLog, int64, error) {
	if s.repo == nil {
		return []ActivityLog{}, 0, nil
	}
	return s.repo.List(ctx, storeID, f)
}
