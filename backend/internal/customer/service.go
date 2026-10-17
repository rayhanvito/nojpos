package customer

import (
	"context"
	"errors"
	"strings"
)

// Repository adalah kontrak penyimpanan pelanggan yang dipakai
// Service (CustomerRepository lama di service/interfaces.go, bertipe
// slice).
type Repository interface {
	List(ctx context.Context, storeID uint, f CustomerFilter) (*CustomerPage, error)
	GetByID(ctx context.Context, storeID, id uint) (*Customer, error)
	Create(ctx context.Context, storeID uint, customer *Customer) (*Customer, error)
	Update(ctx context.Context, storeID, id uint, customer *Customer) (*Customer, error)
	Delete(ctx context.Context, storeID, id uint) error
	ListTransactions(ctx context.Context, storeID, customerID uint) ([]*trx, error)
}

// Service adalah CustomerService lama di service/customer.go, dipindah
// verbatim.
type Service struct {
	customers Repository
}

func NewService(customers Repository) *Service {
	return &Service{customers: customers}
}

func normalizeCustomer(in *CustomerInput) error {
	in.Name = strings.TrimSpace(in.Name)
	in.Phone = strings.TrimSpace(in.Phone)
	in.Email = strings.TrimSpace(in.Email)
	in.Address = strings.TrimSpace(in.Address)
	if in.Name == "" {
		return ErrCustomerNameRequired
	}
	return nil
}

func (s *Service) List(ctx context.Context, storeID uint, f CustomerFilter) (*CustomerPage, error) {
	return s.customers.List(ctx, storeID, f)
}

func (s *Service) Get(ctx context.Context, storeID, id uint) (*Customer, error) {
	customer, err := s.customers.GetByID(ctx, storeID, id)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrCustomerNotFound
	}
	return customer, err
}

func (s *Service) Create(ctx context.Context, storeID uint, in CustomerInput) (*Customer, error) {
	if err := normalizeCustomer(&in); err != nil {
		return nil, err
	}
	customer, err := s.customers.Create(ctx, storeID, &Customer{
		Name: in.Name, Phone: in.Phone, Email: in.Email, Address: in.Address,
	})
	if errors.Is(err, ErrDuplicate) {
		return nil, ErrCustomerPhoneTaken
	}
	return customer, err
}

func (s *Service) Update(ctx context.Context, storeID, id uint, in CustomerInput) (*Customer, error) {
	if err := normalizeCustomer(&in); err != nil {
		return nil, err
	}
	customer, err := s.customers.Update(ctx, storeID, id, &Customer{
		Name: in.Name, Phone: in.Phone, Email: in.Email, Address: in.Address,
	})
	if errors.Is(err, ErrNotFound) {
		return nil, ErrCustomerNotFound
	}
	if errors.Is(err, ErrDuplicate) {
		return nil, ErrCustomerPhoneTaken
	}
	return customer, err
}

func (s *Service) Delete(ctx context.Context, storeID, id uint) error {
	err := s.customers.Delete(ctx, storeID, id)
	if errors.Is(err, ErrNotFound) {
		return ErrCustomerNotFound
	}
	return err
}

func (s *Service) Transactions(ctx context.Context, storeID, id uint) ([]*trx, error) {
	if _, err := s.Get(ctx, storeID, id); err != nil {
		return nil, err
	}
	return s.customers.ListTransactions(ctx, storeID, id)
}
