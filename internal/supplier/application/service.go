package application

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wyfcoding/ecommerce/internal/supplier/domain"
)

type SupplierApplicationService struct {
	repo domain.SupplierRepository
}

func NewSupplierApplicationService(repo domain.SupplierRepository) *SupplierApplicationService {
	return &SupplierApplicationService{repo: repo}
}

func (s *SupplierApplicationService) RegisterSupplier(ctx context.Context, name, contact, phone, email, addr, license string) (string, error) {
	id := fmt.Sprintf("SUP%d", time.Now().UnixNano())
	supplier := domain.NewSupplier(id, name, contact, phone, email, addr, license)

	if err := s.repo.Save(ctx, supplier); err != nil {
		return "", err
	}
	return id, nil
}

func (s *SupplierApplicationService) UpdateStatus(ctx context.Context, id string, statusStr string) error {
	supplier, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	switch statusStr {
	case "ACTIVE":
		supplier.Status = domain.StatusActive
	case "INACTIVE":
		supplier.Status = domain.StatusInactive
	case "SUSPENDED":
		supplier.Status = domain.StatusSuspended
	}

	return s.repo.Save(ctx, supplier)
}

func (s *SupplierApplicationService) AddProductSupply(ctx context.Context, supplierID, skuID string, price float64, leadTime int32) error {
	supplier, err := s.repo.Get(ctx, supplierID)
	if err != nil {
		return err
	}

	if err := supplier.AddProductSupply(skuID, decimal.NewFromFloat(price), leadTime); err != nil {
		return err
	}

	return s.repo.Save(ctx, supplier)
}
