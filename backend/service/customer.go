package service

import (
	"context"
	"errors"
	"strings"

	"github.com/adrr-dev/openPOS/backend/model"
	"github.com/adrr-dev/openPOS/backend/repo"
)

var (
	ErrCustomerNotFound     = errors.New("pelanggan tidak ditemukan")
	ErrCustomerPhoneTaken   = errors.New("nomor telepon sudah digunakan di toko ini")
	ErrCustomerNameRequired = errors.New("nama pelanggan wajib diisi")
)

type CustomerInput struct {
	Name    string
	Phone   string
	Email   string
	Address string
}

type CustomerService struct {
	customers CustomerRepository
}

func NewCustomerService(customers CustomerRepository) *CustomerService {
	return &CustomerService{customers: customers}
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

func (s *CustomerService) List(ctx context.Context, storeID uint, f repo.CustomerFilter) (*repo.CustomerPage, error) {
	return s.customers.List(ctx, storeID, f)
}

func (s *CustomerService) Get(ctx context.Context, storeID, id uint) (*model.Customer, error) {
	customer, err := s.customers.GetByID(ctx, storeID, id)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, ErrCustomerNotFound
	}
	return customer, err
}

func (s *CustomerService) Create(ctx context.Context, storeID uint, in CustomerInput) (*model.Customer, error) {
	if err := normalizeCustomer(&in); err != nil {
		return nil, err
	}
	customer, err := s.customers.Create(ctx, storeID, &model.Customer{
		Name: in.Name, Phone: in.Phone, Email: in.Email, Address: in.Address,
	})
	if errors.Is(err, repo.ErrDuplicate) {
		return nil, ErrCustomerPhoneTaken
	}
	return customer, err
}

func (s *CustomerService) Update(ctx context.Context, storeID, id uint, in CustomerInput) (*model.Customer, error) {
	if err := normalizeCustomer(&in); err != nil {
		return nil, err
	}
	customer, err := s.customers.Update(ctx, storeID, id, &model.Customer{
		Name: in.Name, Phone: in.Phone, Email: in.Email, Address: in.Address,
	})
	if errors.Is(err, repo.ErrNotFound) {
		return nil, ErrCustomerNotFound
	}
	if errors.Is(err, repo.ErrDuplicate) {
		return nil, ErrCustomerPhoneTaken
	}
	return customer, err
}

func (s *CustomerService) Delete(ctx context.Context, storeID, id uint) error {
	err := s.customers.Delete(ctx, storeID, id)
	if errors.Is(err, repo.ErrNotFound) {
		return ErrCustomerNotFound
	}
	return err
}

func (s *CustomerService) Transactions(ctx context.Context, storeID, id uint) ([]*model.Trx, error) {
	if _, err := s.Get(ctx, storeID, id); err != nil {
		return nil, err
	}
	return s.customers.ListTransactions(ctx, storeID, id)
}
