package repository

import (
	"context"
	"github.com/manmohansharma21/orderflow/internal/domain"
)

// ProductRepositoryAdapter is the product persistence contract consumed by the service layer.
type ProductRepositoryAdapter interface {
	List(context.Context, string, int, int) ([]domain.Product, int64, error)
	GetByID(context.Context, string) (domain.Product, error)
	Create(context.Context, domain.CreateProductInput) (domain.Product, error)
	Update(context.Context, string, domain.UpdateProductInput) (domain.Product, error)
	Deactivate(context.Context, string) (domain.Product, error)
}
