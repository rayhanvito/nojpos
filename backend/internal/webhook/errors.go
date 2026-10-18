package webhook

import apierr "github.com/rayhanvito/nojpos/backend/internal/platform/errors"

var (
	ErrNotFound     = apierr.New(404, "WEBHOOK_NOT_FOUND", "Webhook tidak ditemukan.")
	ErrInvalidURL   = apierr.New(400, "WEBHOOK_URL_INVALID", "URL webhook tidak valid.")
	ErrPrivateURL   = apierr.New(400, "WEBHOOK_URL_PRIVATE", "URL webhook tidak boleh menunjuk alamat privat, loopback, atau link-local.")
	ErrEventsEmpty  = apierr.New(400, "WEBHOOK_EVENTS_INVALID", "Events webhook wajib berisi minimal satu event.")
	ErrEventUnknown = apierr.New(400, "WEBHOOK_EVENTS_INVALID", "Ada nama event webhook yang tidak dikenal.")
)
