package staff

import (
	"net/http"

	"github.com/rayhanvito/nojpos/backend/internal/outlet"
	apierr "github.com/rayhanvito/nojpos/backend/internal/platform/errors"
)

var (
	// ErrCashierOpenShift dan ErrCashierOutletChanged adalah sentinel
	// apierr yang sama persis dengan yang dulu didefinisikan di
	// repo/base.go (status, code, dan pesan identik). repo/base.go
	// kini meng-alias kedua var ini (transisional) supaya produsen
	// yang menetap (repo/shift.go) dan seluruh rantai errors.Is —
	// termasuk pemetaan service.ErrCashierOpenShift di
	// service/user.go dan service.ErrCashierOutletChanged di
	// service/shift.go — tetap menunjuk instans yang sama.
	ErrCashierOpenShift     = apierr.New(http.StatusConflict, "CASHIER_OPEN_SHIFT", "kasir memiliki shift terbuka")
	ErrCashierOutletChanged = apierr.New(http.StatusConflict, "CASHIER_OUTLET_CHANGED", "outlet kasir sudah berubah")

	// ErrNotFound dan ErrDuplicate adalah alias ke instans kanonis
	// slice outlet (BE-F1a): konsumen menetap memeriksa
	// repo.ErrNotFound pada hasil repo kasir (mis. service/user.go
	// pada GetByIDUnscoped), dan repo.ErrNotFound sendiri adalah
	// alias ke instans yang sama — jadi errors.Is tetap cocok.
	ErrNotFound  = outlet.ErrNotFound
	ErrDuplicate = outlet.ErrDuplicate
)
