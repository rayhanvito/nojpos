package migrations

import "embed"

// FS berisi semua file migrasi SQL di direktori ini.
//
//go:embed *.sql
var FS embed.FS
