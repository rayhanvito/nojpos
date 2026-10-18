package apikey

import apierr "github.com/rayhanvito/nojpos/backend/internal/platform/errors"

var (
	ErrNotFound    = apierr.New(404, "NOT_FOUND", "Kunci API tidak ditemukan.")
	ErrInvalidName = apierr.New(400, "API_KEY_NAME_REQUIRED", "Nama kunci API wajib diisi.")
)
