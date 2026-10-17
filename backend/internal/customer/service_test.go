package customer

import (
	"context"
	"errors"
	"testing"
)

// fakeRepo adalah Repository palsu: hanya perilaku yang dikonfigurasi
// tes yang nyata; sisanya mengembalikan nilai nol.
type fakeRepo struct {
	getErr    error
	createErr error
	updateErr error
	deleteErr error
	listTrx   []*trx

	created   *Customer
	updated   *Customer
	trxCalled bool
}

func (f *fakeRepo) List(ctx context.Context, storeID uint, filter CustomerFilter) (*CustomerPage, error) {
	return &CustomerPage{Items: []*Customer{}}, nil
}

func (f *fakeRepo) GetByID(ctx context.Context, storeID, id uint) (*Customer, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return &Customer{}, nil
}

func (f *fakeRepo) Create(ctx context.Context, storeID uint, customer *Customer) (*Customer, error) {
	f.created = customer
	if f.createErr != nil {
		return nil, f.createErr
	}
	return customer, nil
}

func (f *fakeRepo) Update(ctx context.Context, storeID, id uint, customer *Customer) (*Customer, error) {
	f.updated = customer
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	return customer, nil
}

func (f *fakeRepo) Delete(ctx context.Context, storeID, id uint) error {
	return f.deleteErr
}

func (f *fakeRepo) ListTransactions(ctx context.Context, storeID, customerID uint) ([]*trx, error) {
	f.trxCalled = true
	return f.listTrx, nil
}

func TestCreateEmptyNameRequired(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)
	if _, err := svc.Create(context.Background(), 1, CustomerInput{Name: "   "}); !errors.Is(err, ErrCustomerNameRequired) {
		t.Fatalf("nama kosong harus ErrCustomerNameRequired, dapat %v", err)
	}
	if repo.created != nil {
		t.Fatalf("repo tidak boleh terpanggil saat nama kosong")
	}
}

func TestCreateTrimsFields(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)
	created, err := svc.Create(context.Background(), 1, CustomerInput{
		Name: "  Budi  ", Phone: " 0812 ", Email: " b@mail.id ", Address: " Jl. A ",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.Name != "Budi" || created.Phone != "0812" || created.Email != "b@mail.id" || created.Address != "Jl. A" {
		t.Fatalf("field harus ter-trim, dapat %+v", created)
	}
}

func TestCreateDuplicateMapsToPhoneTaken(t *testing.T) {
	svc := NewService(&fakeRepo{createErr: ErrDuplicate})
	if _, err := svc.Create(context.Background(), 1, CustomerInput{Name: "Budi"}); !errors.Is(err, ErrCustomerPhoneTaken) {
		t.Fatalf("ErrDuplicate dari repo harus menjadi ErrCustomerPhoneTaken, dapat %v", err)
	}
}

func TestGetNotFoundMapsToCustomerNotFound(t *testing.T) {
	svc := NewService(&fakeRepo{getErr: ErrNotFound})
	if _, err := svc.Get(context.Background(), 1, 9); !errors.Is(err, ErrCustomerNotFound) {
		t.Fatalf("ErrNotFound dari repo harus menjadi ErrCustomerNotFound, dapat %v", err)
	}
}

func TestUpdateNotFoundAndDuplicateMapping(t *testing.T) {
	svc := NewService(&fakeRepo{updateErr: ErrNotFound})
	if _, err := svc.Update(context.Background(), 1, 9, CustomerInput{Name: "Budi"}); !errors.Is(err, ErrCustomerNotFound) {
		t.Fatalf("ErrNotFound dari repo harus menjadi ErrCustomerNotFound, dapat %v", err)
	}
	svc = NewService(&fakeRepo{updateErr: ErrDuplicate})
	if _, err := svc.Update(context.Background(), 1, 9, CustomerInput{Name: "Budi"}); !errors.Is(err, ErrCustomerPhoneTaken) {
		t.Fatalf("ErrDuplicate dari repo harus menjadi ErrCustomerPhoneTaken, dapat %v", err)
	}
}

func TestDeleteNotFoundMapsToCustomerNotFound(t *testing.T) {
	svc := NewService(&fakeRepo{deleteErr: ErrNotFound})
	if err := svc.Delete(context.Background(), 1, 9); !errors.Is(err, ErrCustomerNotFound) {
		t.Fatalf("ErrNotFound dari repo harus menjadi ErrCustomerNotFound, dapat %v", err)
	}
}

func TestTransactionsRequiresExistingCustomer(t *testing.T) {
	repo := &fakeRepo{getErr: ErrNotFound}
	svc := NewService(repo)
	if _, err := svc.Transactions(context.Background(), 1, 9); !errors.Is(err, ErrCustomerNotFound) {
		t.Fatalf("pelanggan tak ditemukan harus ErrCustomerNotFound, dapat %v", err)
	}
	if repo.trxCalled {
		t.Fatalf("ListTransactions tidak boleh terpanggil untuk pelanggan yang tidak ada")
	}

	want := []*trx{{}}
	repo = &fakeRepo{listTrx: want}
	svc = NewService(repo)
	got, err := svc.Transactions(context.Background(), 1, 9)
	if err != nil {
		t.Fatalf("Transactions: %v", err)
	}
	if len(got) != 1 || got[0] != want[0] {
		t.Fatalf("transaksi harus diteruskan dari repo, dapat %+v", got)
	}
}
