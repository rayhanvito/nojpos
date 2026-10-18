package outlet

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
)

// fakeOutletRepo adalah Repository palsu untuk Service outlet: hanya
// perilaku yang dikonfigurasi tes yang nyata; sisanya nilai nol.
type fakeOutletRepo struct {
	listItems  []Outlet
	listErr    error
	getOut     *Outlet
	getErr     error
	createErr  error
	updateErr  error
	defaultOut *Outlet
	defaultErr error

	listStoreID      uint
	getStoreID       uint
	getID            uint
	createCalls      int
	created          *Outlet
	createdStoreID   uint
	updateCalls      int
	updated          Outlet
	updatedStoreID   uint
	updatedID        uint
	updatedIsDefault *bool
}

func (f *fakeOutletRepo) List(ctx context.Context, storeID uint) ([]Outlet, error) {
	f.listStoreID = storeID
	return f.listItems, f.listErr
}

func (f *fakeOutletRepo) GetByID(ctx context.Context, storeID, id uint) (*Outlet, error) {
	f.getStoreID, f.getID = storeID, id
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.getOut, nil
}

func (f *fakeOutletRepo) GetDefault(ctx context.Context, storeID uint) (*Outlet, error) {
	return f.defaultOut, f.defaultErr
}

func (f *fakeOutletRepo) Create(ctx context.Context, storeID uint, outlet *Outlet) (*Outlet, error) {
	f.createCalls++
	f.created, f.createdStoreID = outlet, storeID
	if f.createErr != nil {
		return nil, f.createErr
	}
	return outlet, nil
}

func (f *fakeOutletRepo) Update(ctx context.Context, storeID, id uint, outlet Outlet, isDefault *bool) (*Outlet, error) {
	f.updateCalls++
	f.updated, f.updatedStoreID, f.updatedID, f.updatedIsDefault = outlet, storeID, id, isDefault
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	out := outlet
	out.ID, out.StoreID = id, storeID
	if isDefault != nil {
		out.IsDefault = *isDefault
	}
	return &out, nil
}

// fakeSettingsRepo adalah SettingsStoreRepository palsu untuk
// SettingsService: hanya perilaku yang dikonfigurasi tes yang nyata.
type fakeSettingsRepo struct {
	getOut    *StoreSettings
	getErr    error
	updateErr error

	getStoreID     uint
	updateCalls    int
	updated        *StoreSettings
	updatedStoreID uint

	// Unit C: outlet + override pengaturan per outlet. outletErr
	// mengembalikan error GetOutlet (ErrNotFound untuk outlet asing);
	// tanpa konfigurasi, outlet dianggap ada. ovOut adalah baris
	// override tersimpan; Upsert menyimpan argumennya ke sana.
	outletErr   error
	ovOut       *OutletSettings
	upsertErr   error
	upsertCalls int
	upserted    *OutletSettings
}

func (f *fakeSettingsRepo) GetSettings(ctx context.Context, storeID uint) (*StoreSettings, error) {
	f.getStoreID = storeID
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.getOut, nil
}

func (f *fakeSettingsRepo) UpdateSettings(ctx context.Context, storeID uint, s *StoreSettings) (*StoreSettings, error) {
	f.updateCalls++
	f.updated, f.updatedStoreID = s, storeID
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	return s, nil
}

func (f *fakeSettingsRepo) GetOutlet(ctx context.Context, storeID, outletID uint) (*Outlet, error) {
	if f.outletErr != nil {
		return nil, f.outletErr
	}
	o := &Outlet{StoreID: storeID}
	o.ID = outletID
	return o, nil
}

func (f *fakeSettingsRepo) GetOutletSettings(ctx context.Context, storeID, outletID uint) (*OutletSettings, error) {
	return f.ovOut, nil
}

func (f *fakeSettingsRepo) UpsertOutletSettings(ctx context.Context, row *OutletSettings) (*OutletSettings, error) {
	f.upsertCalls++
	f.upserted = row
	if f.upsertErr != nil {
		return nil, f.upsertErr
	}
	f.ovOut = row
	return row, nil
}

func strPtr(s string) *string { return &s }

func boolPtr(b bool) *bool { return &b }

// validSettings adalah pengaturan toko minimal yang lolos semua
// validasi SettingsService.Update; tes memutasi satu field per kasus.
func validSettings() *StoreSettings {
	return &StoreSettings{
		Name:              "Toko Uji",
		Timezone:          "Asia/Jakarta",
		Paper:             "80mm",
		TaxPct:            11,
		LoyaltyEarnPer:    10000,
		LoyaltyPointValue: 100,
		CreditDueDays:     14,
		LowStockDefault:   5,
	}
}

// --- Service outlet: Create ---

func TestCreateOutletBlankNameRejected(t *testing.T) {
	repo := &fakeOutletRepo{}
	svc := NewService(repo)
	for _, name := range []string{"", "   "} {
		_, err := svc.Create(context.Background(), 1, Outlet{Name: name})
		if err == nil || err.Error() != "nama outlet wajib diisi" {
			t.Fatalf("nama %q harus ditolak 'nama outlet wajib diisi', dapat %v", name, err)
		}
	}
	if repo.createCalls != 0 {
		t.Fatalf("repo tidak boleh dipanggil untuk nama kosong, createCalls=%d", repo.createCalls)
	}
}

func TestCreateOutletTrimsNameAndAddress(t *testing.T) {
	repo := &fakeOutletRepo{}
	svc := NewService(repo)
	got, err := svc.Create(context.Background(), 7, Outlet{Name: "  Outlet Pusat  ", Address: "  Jl. Merdeka 1  ", IsDefault: true})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if repo.createdStoreID != 7 {
		t.Fatalf("storeID harus diteruskan ke repo, dapat %d", repo.createdStoreID)
	}
	if repo.created.Name != "Outlet Pusat" || repo.created.Address != "Jl. Merdeka 1" {
		t.Fatalf("nama/alamat harus ter-trim sebelum ke repo, dapat %+v", repo.created)
	}
	if got != repo.created {
		t.Fatalf("hasil harus objek dari repo, dapat %+v", got)
	}
	if !got.IsDefault {
		t.Fatalf("IsDefault dari input harus diteruskan apa adanya, dapat %+v", got)
	}
}

func TestCreateOutletRepoErrorPropagates(t *testing.T) {
	sentinel := errors.New("db tumbang")
	svc := NewService(&fakeOutletRepo{createErr: sentinel})
	if _, err := svc.Create(context.Background(), 1, Outlet{Name: "Pusat"}); !errors.Is(err, sentinel) {
		t.Fatalf("error repo harus diteruskan apa adanya, dapat %v", err)
	}
}

// --- Service outlet: List ---

func TestListOutletsPassesThroughRepo(t *testing.T) {
	want := []Outlet{{Name: "Pusat"}, {Name: "Cabang"}}
	want[0].ID, want[1].ID = 1, 2
	repo := &fakeOutletRepo{listItems: want}
	svc := NewService(repo)
	got, err := svc.List(context.Background(), 3)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if repo.listStoreID != 3 {
		t.Fatalf("storeID harus diteruskan ke repo, dapat %d", repo.listStoreID)
	}
	if len(got) != 2 || got[0].ID != 1 || got[1].ID != 2 {
		t.Fatalf("daftar harus diteruskan dari repo, dapat %+v", got)
	}

	sentinel := errors.New("db tumbang")
	svc = NewService(&fakeOutletRepo{listErr: sentinel})
	if _, err := svc.List(context.Background(), 3); !errors.Is(err, sentinel) {
		t.Fatalf("error repo harus diteruskan apa adanya, dapat %v", err)
	}
}

// --- Service outlet: Get ---

func TestGetOutletNotFoundMapsToOutletNotFound(t *testing.T) {
	svc := NewService(&fakeOutletRepo{getErr: ErrNotFound})
	if _, err := svc.Get(context.Background(), 1, 99); !errors.Is(err, ErrOutletNotFound) {
		t.Fatalf("ErrNotFound dari repo harus menjadi ErrOutletNotFound, dapat %v", err)
	}
	// ErrNotFound yang terbungkus (wrapped) pun harus terpetakan.
	svc = NewService(&fakeOutletRepo{getErr: fmt.Errorf("query: %w", ErrNotFound)})
	if _, err := svc.Get(context.Background(), 1, 99); !errors.Is(err, ErrOutletNotFound) {
		t.Fatalf("ErrNotFound terbungkus harus menjadi ErrOutletNotFound, dapat %v", err)
	}
}

func TestGetOutletSuccessAndOtherErrorUnmapped(t *testing.T) {
	want := &Outlet{Name: "Pusat", StoreID: 1}
	want.ID = 9
	repo := &fakeOutletRepo{getOut: want}
	svc := NewService(repo)
	got, err := svc.Get(context.Background(), 1, 9)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != want || repo.getStoreID != 1 || repo.getID != 9 {
		t.Fatalf("outlet/storeID/id harus diteruskan dari repo, dapat %+v (%d,%d)", got, repo.getStoreID, repo.getID)
	}

	sentinel := errors.New("db tumbang")
	svc = NewService(&fakeOutletRepo{getErr: sentinel})
	_, err = svc.Get(context.Background(), 1, 9)
	if !errors.Is(err, sentinel) || errors.Is(err, ErrOutletNotFound) {
		t.Fatalf("error non-NotFound tidak boleh dipetakan, dapat %v", err)
	}
}

// --- Service outlet: Update ---

func TestUpdateOutletBlankNameRejected(t *testing.T) {
	repo := &fakeOutletRepo{}
	svc := NewService(repo)
	_, err := svc.Update(context.Background(), 1, 9, Outlet{Name: "   "}, nil)
	if err == nil || err.Error() != "nama outlet wajib diisi" {
		t.Fatalf("nama kosong harus ditolak 'nama outlet wajib diisi', dapat %v", err)
	}
	if repo.updateCalls != 0 {
		t.Fatalf("repo tidak boleh dipanggil untuk nama kosong, updateCalls=%d", repo.updateCalls)
	}
}

func TestUpdateOutletTrimsAndForwardsIsDefault(t *testing.T) {
	repo := &fakeOutletRepo{}
	svc := NewService(repo)
	isDefault := boolPtr(true)
	got, err := svc.Update(context.Background(), 4, 9, Outlet{Name: "  Cabang Baru  ", Address: "  Jl. Sudirman  "}, isDefault)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if repo.updatedStoreID != 4 || repo.updatedID != 9 {
		t.Fatalf("storeID/id harus diteruskan ke repo, dapat (%d,%d)", repo.updatedStoreID, repo.updatedID)
	}
	if repo.updated.Name != "Cabang Baru" || repo.updated.Address != "Jl. Sudirman" {
		t.Fatalf("nama/alamat harus ter-trim sebelum ke repo, dapat %+v", repo.updated)
	}
	if repo.updatedIsDefault != isDefault {
		t.Fatalf("pointer isDefault harus diteruskan apa adanya ke repo")
	}
	if got.ID != 9 || got.StoreID != 4 || !got.IsDefault {
		t.Fatalf("hasil harus dari repo, dapat %+v", got)
	}

	// isDefault nil (tidak diubah) juga harus diteruskan sebagai nil.
	repo = &fakeOutletRepo{}
	svc = NewService(repo)
	if _, err := svc.Update(context.Background(), 4, 9, Outlet{Name: "Cabang"}, nil); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if repo.updatedIsDefault != nil {
		t.Fatalf("isDefault nil harus diteruskan sebagai nil, dapat %v", repo.updatedIsDefault)
	}
}

func TestUpdateOutletNotFoundMapsToOutletNotFound(t *testing.T) {
	svc := NewService(&fakeOutletRepo{updateErr: ErrNotFound})
	if _, err := svc.Update(context.Background(), 1, 99, Outlet{Name: "Pusat"}, nil); !errors.Is(err, ErrOutletNotFound) {
		t.Fatalf("ErrNotFound dari repo harus menjadi ErrOutletNotFound, dapat %v", err)
	}
	svc = NewService(&fakeOutletRepo{updateErr: fmt.Errorf("query: %w", ErrNotFound)})
	if _, err := svc.Update(context.Background(), 1, 99, Outlet{Name: "Pusat"}, nil); !errors.Is(err, ErrOutletNotFound) {
		t.Fatalf("ErrNotFound terbungkus harus menjadi ErrOutletNotFound, dapat %v", err)
	}
}

func TestUpdateOutletOtherErrorsPassThrough(t *testing.T) {
	// ErrOutletDefault (repo menolak menonaktifkan outlet default)
	// tidak dipetakan service — diteruskan apa adanya.
	svc := NewService(&fakeOutletRepo{updateErr: ErrOutletDefault})
	_, err := svc.Update(context.Background(), 1, 9, Outlet{Name: "Pusat"}, boolPtr(false))
	if !errors.Is(err, ErrOutletDefault) || errors.Is(err, ErrOutletNotFound) {
		t.Fatalf("ErrOutletDefault harus diteruskan apa adanya, dapat %v", err)
	}
	sentinel := errors.New("db tumbang")
	svc = NewService(&fakeOutletRepo{updateErr: sentinel})
	if _, err := svc.Update(context.Background(), 1, 9, Outlet{Name: "Pusat"}, nil); !errors.Is(err, sentinel) {
		t.Fatalf("error repo lain harus diteruskan apa adanya, dapat %v", err)
	}
}

// --- SettingsService: Get ---

func TestSettingsGetPassesThroughRepo(t *testing.T) {
	want := validSettings()
	repo := &fakeSettingsRepo{getOut: want}
	svc := NewSettingsService(repo)
	got, err := svc.Get(context.Background(), 5)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != want || repo.getStoreID != 5 {
		t.Fatalf("pengaturan/storeID harus diteruskan dari repo, dapat %+v (%d)", got, repo.getStoreID)
	}

	sentinel := errors.New("db tumbang")
	svc = NewSettingsService(&fakeSettingsRepo{getErr: sentinel})
	if _, err := svc.Get(context.Background(), 5); !errors.Is(err, sentinel) {
		t.Fatalf("error repo harus diteruskan apa adanya, dapat %v", err)
	}
}

// --- SettingsService: Update, normalisasi & default ---

func TestSettingsUpdateDefaultsNormalized(t *testing.T) {
	repo := &fakeSettingsRepo{}
	svc := NewSettingsService(repo)
	in := validSettings()
	in.Name = "  Toko A  "
	in.Timezone = "   " // kosong -> Asia/Makassar
	in.Paper = "A4"     // tak dikenal -> 58mm
	in.TaxPct = -5      // negatif -> 0
	in.BusinessType = " retail "
	in.Email = "  toko@mail.id "
	in.City = "  Makassar "
	in.Province = "  Sulawesi Selatan "
	in.Currency = " idr " // -> IDR
	in.TaxName = "  PB1  "
	// Hours nil, TaxRounding "", TaxApplyTo "" -> dinormalisasi service.
	got, err := svc.Update(context.Background(), 2, in)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if repo.updateCalls != 1 || repo.updatedStoreID != 2 || repo.updated != in {
		t.Fatalf("repo harus dipanggil sekali dengan input yang sama, calls=%d storeID=%d", repo.updateCalls, repo.updatedStoreID)
	}
	if got.Name != "Toko A" || got.Timezone != "Asia/Makassar" || got.Paper != "58mm" || got.TaxPct != 0 {
		t.Fatalf("nama/timezone/paper/taxPct harus dinormalisasi, dapat %+v", got)
	}
	if got.BusinessType != "retail" || got.Email != "toko@mail.id" || got.City != "Makassar" || got.Province != "Sulawesi Selatan" {
		t.Fatalf("field extended harus ter-trim, dapat %+v", got)
	}
	if got.Currency != "IDR" || got.TaxName != "PB1" || got.TaxRounding != "none" || got.TaxApplyTo != "all" {
		t.Fatalf("currency/taxName/taxRounding/taxApplyTo harus dinormalisasi, dapat %+v", got)
	}
	if got.Hours == nil || len(got.Hours) != 0 {
		t.Fatalf("Hours nil harus menjadi slice kosong non-nil, dapat %+v", got.Hours)
	}
}

func TestSettingsUpdateKeepsValidValues(t *testing.T) {
	repo := &fakeSettingsRepo{}
	svc := NewSettingsService(repo)
	in := validSettings()
	in.Currency = "usd" // huruf kecil -> USD
	in.TaxRounding = "up"
	in.TaxApplyTo = "all"
	got, err := svc.Update(context.Background(), 1, in)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if got.Timezone != "Asia/Jakarta" || got.Paper != "80mm" || got.TaxPct != 11 {
		t.Fatalf("nilai valid timezone/paper/taxPct harus dipertahankan, dapat %+v", got)
	}
	if got.Currency != "USD" || got.TaxRounding != "up" || got.TaxApplyTo != "all" {
		t.Fatalf("nilai valid currency/taxRounding/taxApplyTo harus dipertahankan, dapat %+v", got)
	}
	if got.LoyaltyEarnPer != 10000 || got.LoyaltyPointValue != 100 || got.CreditDueDays != 14 || got.LowStockDefault != 5 {
		t.Fatalf("nilai loyalty/credit/lowStock valid harus dipertahankan, dapat %+v", got)
	}
}

func TestSettingsUpdatePaperCoercion(t *testing.T) {
	for paper, want := range map[string]string{"": "58mm", "A4": "58mm", "58mm": "58mm", "80mm": "80mm"} {
		svc := NewSettingsService(&fakeSettingsRepo{})
		in := validSettings()
		in.Paper = paper
		got, err := svc.Update(context.Background(), 1, in)
		if err != nil {
			t.Fatalf("paper %q: %v", paper, err)
		}
		if got.Paper != want {
			t.Fatalf("paper %q harus menjadi %q, dapat %q", paper, want, got.Paper)
		}
	}
}

func TestSettingsUpdateRepoErrorPropagates(t *testing.T) {
	sentinel := errors.New("db tumbang")
	repo := &fakeSettingsRepo{updateErr: sentinel}
	svc := NewSettingsService(repo)
	if _, err := svc.Update(context.Background(), 1, validSettings()); !errors.Is(err, sentinel) {
		t.Fatalf("error repo harus diteruskan apa adanya, dapat %v", err)
	}
	if repo.updateCalls != 1 {
		t.Fatalf("repo harus tetap dipanggil untuk input valid, calls=%d", repo.updateCalls)
	}
}

// --- SettingsService: Update, penolakan numerik & nama ---

func TestSettingsUpdateRejectsNegativeLowStockDefault(t *testing.T) {
	repo := &fakeSettingsRepo{}
	svc := NewSettingsService(repo)
	in := validSettings()
	in.LowStockDefault = -1
	if _, err := svc.Update(context.Background(), 1, in); !errors.Is(err, ErrBadLowStockDefault) {
		t.Fatalf("LowStockDefault negatif harus ErrBadLowStockDefault, dapat %v", err)
	}
	if repo.updateCalls != 0 {
		t.Fatalf("repo tidak boleh dipanggil untuk input tak valid, calls=%d", repo.updateCalls)
	}
}

func TestSettingsUpdateRejectsNonPositiveLoyaltyEarnPer(t *testing.T) {
	for _, v := range []int64{0, -1000} {
		repo := &fakeSettingsRepo{}
		svc := NewSettingsService(repo)
		in := validSettings()
		in.LoyaltyEarnPer = v
		if _, err := svc.Update(context.Background(), 1, in); !errors.Is(err, ErrBadLoyaltyEarnPer) {
			t.Fatalf("LoyaltyEarnPer %d harus ErrBadLoyaltyEarnPer, dapat %v", v, err)
		}
		if repo.updateCalls != 0 {
			t.Fatalf("repo tidak boleh dipanggil untuk input tak valid, calls=%d", repo.updateCalls)
		}
	}
}

func TestSettingsUpdateRejectsNegativeLoyaltyPointValue(t *testing.T) {
	repo := &fakeSettingsRepo{}
	svc := NewSettingsService(repo)
	in := validSettings()
	in.LoyaltyPointValue = -1
	if _, err := svc.Update(context.Background(), 1, in); !errors.Is(err, ErrBadLoyaltyPointValue) {
		t.Fatalf("LoyaltyPointValue negatif harus ErrBadLoyaltyPointValue, dapat %v", err)
	}
	if repo.updateCalls != 0 {
		t.Fatalf("repo tidak boleh dipanggil untuk input tak valid, calls=%d", repo.updateCalls)
	}
	// Nol diperbolehkan.
	in = validSettings()
	in.LoyaltyPointValue = 0
	if _, err := svc.Update(context.Background(), 1, in); err != nil {
		t.Fatalf("LoyaltyPointValue 0 harus diterima, dapat %v", err)
	}
}

func TestSettingsUpdateRejectsNegativeCreditDueDays(t *testing.T) {
	repo := &fakeSettingsRepo{}
	svc := NewSettingsService(repo)
	in := validSettings()
	in.CreditDueDays = -1
	if _, err := svc.Update(context.Background(), 1, in); !errors.Is(err, ErrBadCreditDueDays) {
		t.Fatalf("CreditDueDays negatif harus ErrBadCreditDueDays, dapat %v", err)
	}
	if repo.updateCalls != 0 {
		t.Fatalf("repo tidak boleh dipanggil untuk input tak valid, calls=%d", repo.updateCalls)
	}
}

func TestSettingsUpdateRejectsBlankStoreName(t *testing.T) {
	repo := &fakeSettingsRepo{}
	svc := NewSettingsService(repo)
	in := validSettings()
	in.Name = "   "
	_, err := svc.Update(context.Background(), 1, in)
	if err == nil || err.Error() != "nama toko wajib diisi" {
		t.Fatalf("nama toko kosong harus ditolak 'nama toko wajib diisi', dapat %v", err)
	}
	if repo.updateCalls != 0 {
		t.Fatalf("repo tidak boleh dipanggil untuk nama kosong, calls=%d", repo.updateCalls)
	}
}

func TestSettingsUpdateRejectsInvalidTimezone(t *testing.T) {
	repo := &fakeSettingsRepo{}
	svc := NewSettingsService(repo)
	in := validSettings()
	in.Timezone = "Mars/Olympus"
	if _, err := svc.Update(context.Background(), 1, in); !errors.Is(err, ErrBadTimezone) {
		t.Fatalf("timezone tak dikenal harus ErrBadTimezone, dapat %v", err)
	}
	if repo.updateCalls != 0 {
		t.Fatalf("repo tidak boleh dipanggil untuk timezone tak valid, calls=%d", repo.updateCalls)
	}
}

func TestSettingsUpdateTaxPctBounds(t *testing.T) {
	repo := &fakeSettingsRepo{}
	svc := NewSettingsService(repo)
	in := validSettings()
	in.TaxPct = 100.5
	if _, err := svc.Update(context.Background(), 1, in); !errors.Is(err, ErrBadTaxPct) {
		t.Fatalf("TaxPct >100 harus ErrBadTaxPct, dapat %v", err)
	}
	if repo.updateCalls != 0 {
		t.Fatalf("repo tidak boleh dipanggil untuk TaxPct >100, calls=%d", repo.updateCalls)
	}
	// Batas 100 diterima; negatif dijepit ke 0 (bukan error).
	in = validSettings()
	in.TaxPct = 100
	if _, err := svc.Update(context.Background(), 1, in); err != nil {
		t.Fatalf("TaxPct 100 harus diterima, dapat %v", err)
	}
	in = validSettings()
	in.TaxPct = -0.5
	got, err := svc.Update(context.Background(), 1, in)
	if err != nil || got.TaxPct != 0 {
		t.Fatalf("TaxPct negatif harus dijepit ke 0, dapat %v %+v", err, got)
	}
}

// --- SettingsService: validateExtendedSettings ---

func TestSettingsUpdateRejectsInvalidBusinessType(t *testing.T) {
	for _, bt := range []string{"warung", "Retail", "RETAIL"} {
		repo := &fakeSettingsRepo{}
		svc := NewSettingsService(repo)
		in := validSettings()
		in.BusinessType = bt
		if _, err := svc.Update(context.Background(), 1, in); !errors.Is(err, ErrBadBusinessType) {
			t.Fatalf("BusinessType %q harus ErrBadBusinessType, dapat %v", bt, err)
		}
		if repo.updateCalls != 0 {
			t.Fatalf("repo tidak boleh dipanggil untuk BusinessType tak valid, calls=%d", repo.updateCalls)
		}
	}
}

func TestSettingsUpdateValidBusinessTypesAccepted(t *testing.T) {
	for _, bt := range []string{"", "retail", "fnb", "fashion", "jasa", "lainnya"} {
		svc := NewSettingsService(&fakeSettingsRepo{})
		in := validSettings()
		in.BusinessType = bt
		if _, err := svc.Update(context.Background(), 1, in); err != nil {
			t.Fatalf("BusinessType %q harus diterima, dapat %v", bt, err)
		}
	}
}

func TestSettingsUpdateRejectsInvalidStoreEmail(t *testing.T) {
	for _, email := range []string{"bukan-email", "@mail.id", "toko@", "toko@mail", "toko @mail.id"} {
		repo := &fakeSettingsRepo{}
		svc := NewSettingsService(repo)
		in := validSettings()
		in.Email = email
		if _, err := svc.Update(context.Background(), 1, in); !errors.Is(err, ErrBadStoreEmail) {
			t.Fatalf("Email %q harus ErrBadStoreEmail, dapat %v", email, err)
		}
		if repo.updateCalls != 0 {
			t.Fatalf("repo tidak boleh dipanggil untuk email tak valid, calls=%d", repo.updateCalls)
		}
	}
	// Email kosong diperbolehkan.
	svc := NewSettingsService(&fakeSettingsRepo{})
	if _, err := svc.Update(context.Background(), 1, validSettings()); err != nil {
		t.Fatalf("email kosong harus diterima, dapat %v", err)
	}
}

func TestSettingsUpdateRejectsInvalidCurrency(t *testing.T) {
	for _, cur := range []string{"ID", "IDRX", "I1R", "ID$"} {
		repo := &fakeSettingsRepo{}
		svc := NewSettingsService(repo)
		in := validSettings()
		in.Currency = cur
		if _, err := svc.Update(context.Background(), 1, in); !errors.Is(err, ErrBadCurrency) {
			t.Fatalf("Currency %q harus ErrBadCurrency, dapat %v", cur, err)
		}
		if repo.updateCalls != 0 {
			t.Fatalf("repo tidak boleh dipanggil untuk currency tak valid, calls=%d", repo.updateCalls)
		}
	}
}

func TestSettingsUpdateRejectsInvalidHours(t *testing.T) {
	open, close := strPtr("08:00"), strPtr("17:00")
	validDay := func(days string) StoreHours { return StoreHours{Days: days, Open: open, Close: close} }
	cases := map[string][]StoreHours{
		"satu entri":  {validDay("Senin – Jumat")},
		"dua entri":   {validDay("Senin – Jumat"), validDay("Sabtu")},
		"empat entri": {validDay("Senin – Jumat"), validDay("Sabtu"), validDay("Minggu"), validDay("Minggu")},
		"label hari salah": {
			validDay("Senin-Jumat"), validDay("Sabtu"), validDay("Minggu"),
		},
		"hanya jam buka": {
			{Days: "Senin – Jumat", Open: open}, validDay("Sabtu"), validDay("Minggu"),
		},
		"hanya jam tutup": {
			{Days: "Senin – Jumat", Close: close}, validDay("Sabtu"), validDay("Minggu"),
		},
		"buka sama dengan tutup": {
			{Days: "Senin – Jumat", Open: strPtr("09:00"), Close: strPtr("09:00")}, validDay("Sabtu"), validDay("Minggu"),
		},
		"buka setelah tutup": {
			{Days: "Senin – Jumat", Open: strPtr("17:00"), Close: strPtr("08:00")}, validDay("Sabtu"), validDay("Minggu"),
		},
		"format jam buruk": {
			{Days: "Senin – Jumat", Open: strPtr("8:00"), Close: close}, validDay("Sabtu"), validDay("Minggu"),
		},
		"jam 24 tidak valid": {
			{Days: "Senin – Jumat", Open: strPtr("24:00"), Close: strPtr("23:59")}, validDay("Sabtu"), validDay("Minggu"),
		},
		"menit 60 tidak valid": {
			{Days: "Senin – Jumat", Open: strPtr("08:60"), Close: close}, validDay("Sabtu"), validDay("Minggu"),
		},
	}
	for name, hours := range cases {
		repo := &fakeSettingsRepo{}
		svc := NewSettingsService(repo)
		in := validSettings()
		in.Hours = hours
		if _, err := svc.Update(context.Background(), 1, in); !errors.Is(err, ErrBadHours) {
			t.Fatalf("kasus %q harus ErrBadHours, dapat %v", name, err)
		}
		if repo.updateCalls != 0 {
			t.Fatalf("kasus %q: repo tidak boleh dipanggil, calls=%d", name, repo.updateCalls)
		}
	}
}

func TestSettingsUpdateValidHoursAccepted(t *testing.T) {
	svc := NewSettingsService(&fakeSettingsRepo{})
	in := validSettings()
	in.Hours = []StoreHours{
		{Days: "Senin – Jumat", Open: strPtr(" 08:00 "), Close: strPtr(" 17:00 ")},
		{Days: "Sabtu", Open: strPtr("09:00"), Close: strPtr("15:00")},
		{Days: "Minggu"}, // tutup: Open dan Close sama-sama nil
	}
	got, err := svc.Update(context.Background(), 1, in)
	if err != nil {
		t.Fatalf("jam operasional valid harus diterima, dapat %v", err)
	}
	if len(got.Hours) != 3 || *got.Hours[0].Open != "08:00" || *got.Hours[0].Close != "17:00" {
		t.Fatalf("jam harus ter-trim dan dipertahankan, dapat %+v", got.Hours)
	}
	if got.Hours[2].Open != nil || got.Hours[2].Close != nil {
		t.Fatalf("hari tutup harus tetap nil/nil, dapat %+v", got.Hours[2])
	}
}

func TestSettingsUpdateRejectsLongReceiptFooter(t *testing.T) {
	repo := &fakeSettingsRepo{}
	svc := NewSettingsService(repo)
	in := validSettings()
	in.ReceiptFooter = strings.Repeat("a", 201)
	if _, err := svc.Update(context.Background(), 1, in); !errors.Is(err, ErrBadReceiptFooter) {
		t.Fatalf("footer 201 karakter harus ErrBadReceiptFooter, dapat %v", err)
	}
	if repo.updateCalls != 0 {
		t.Fatalf("repo tidak boleh dipanggil untuk footer terlalu panjang, calls=%d", repo.updateCalls)
	}
	in = validSettings()
	in.ReceiptFooter = strings.Repeat("a", 200)
	if _, err := svc.Update(context.Background(), 1, in); err != nil {
		t.Fatalf("footer 200 karakter harus diterima, dapat %v", err)
	}
}

func TestSettingsUpdateRejectsLongTaxName(t *testing.T) {
	repo := &fakeSettingsRepo{}
	svc := NewSettingsService(repo)
	in := validSettings()
	in.TaxName = strings.Repeat("a", 21)
	if _, err := svc.Update(context.Background(), 1, in); !errors.Is(err, ErrBadTaxName) {
		t.Fatalf("nama pajak 21 karakter harus ErrBadTaxName, dapat %v", err)
	}
	if repo.updateCalls != 0 {
		t.Fatalf("repo tidak boleh dipanggil untuk nama pajak terlalu panjang, calls=%d", repo.updateCalls)
	}
	in = validSettings()
	in.TaxName = strings.Repeat("a", 20)
	if _, err := svc.Update(context.Background(), 1, in); err != nil {
		t.Fatalf("nama pajak 20 karakter harus diterima, dapat %v", err)
	}
}

func TestSettingsUpdateTaxRoundingNormalization(t *testing.T) {
	for rounding, want := range map[string]string{"": "none", "none": "none", "down": "down", "up": "up"} {
		svc := NewSettingsService(&fakeSettingsRepo{})
		in := validSettings()
		in.TaxRounding = rounding
		got, err := svc.Update(context.Background(), 1, in)
		if err != nil {
			t.Fatalf("TaxRounding %q: %v", rounding, err)
		}
		if got.TaxRounding != want {
			t.Fatalf("TaxRounding %q harus menjadi %q, dapat %q", rounding, want, got.TaxRounding)
		}
	}
	repo := &fakeSettingsRepo{}
	svc := NewSettingsService(repo)
	in := validSettings()
	in.TaxRounding = "nearest"
	if _, err := svc.Update(context.Background(), 1, in); !errors.Is(err, ErrBadTaxRounding) {
		t.Fatalf("TaxRounding tak dikenal harus ErrBadTaxRounding, dapat %v", err)
	}
	if repo.updateCalls != 0 {
		t.Fatalf("repo tidak boleh dipanggil untuk TaxRounding tak valid, calls=%d", repo.updateCalls)
	}
}

func TestSettingsUpdateRejectsInvalidTaxApplyTo(t *testing.T) {
	repo := &fakeSettingsRepo{}
	svc := NewSettingsService(repo)
	in := validSettings()
	in.TaxApplyTo = "some"
	if _, err := svc.Update(context.Background(), 1, in); !errors.Is(err, ErrBadTaxApplyTo) {
		t.Fatalf("TaxApplyTo selain 'all' harus ErrBadTaxApplyTo, dapat %v", err)
	}
	if repo.updateCalls != 0 {
		t.Fatalf("repo tidak boleh dipanggil untuk TaxApplyTo tak valid, calls=%d", repo.updateCalls)
	}
}

// --- Helper privat di service_settings.go ---

func TestValidHourMinute(t *testing.T) {
	for _, s := range []string{"00:00", "08:30", "23:59"} {
		if !validHourMinute(s) {
			t.Fatalf("%q harus valid", s)
		}
	}
	for _, s := range []string{"", "8:00", "08:0", "0800", "24:00", "08:60", "ab:cd", "08-00"} {
		if validHourMinute(s) {
			t.Fatalf("%q harus tidak valid", s)
		}
	}
}

func TestIsEmail(t *testing.T) {
	for _, s := range []string{"toko@mail.id", "a@b.co"} {
		if !isEmail(s) {
			t.Fatalf("%q harus valid", s)
		}
	}
	for _, s := range []string{"", "bukan-email", "@mail.id", "toko@", "toko@mail", "toko @mail.id", "toko@mail .id"} {
		if isEmail(s) {
			t.Fatalf("%q harus tidak valid", s)
		}
	}
}
