package service

import (
	"context"
	"errors"
	"strings"

	"github.com/manmohansharma21/orderflow/internal/domain"
	"github.com/manmohansharma21/orderflow/internal/repository"
)

var ErrInvalidProduct = errors.New("invalid product")

// ErrProductNotFound is declared in order_service.go and shared by services.
// ErrDuplicateSKU is declared in the repository and translated to HTTP 409.

type ProductService struct {
	repo *repository.ProductRepository
}

func NewProductService(repo *repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) List(
	ctx context.Context,
	search string,
	page, limit int,
) (domain.ProductList, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit
	items, total, err := s.repo.List(ctx, strings.TrimSpace(search), limit, offset)
	if err != nil {
		return domain.ProductList{}, err
	}
	return domain.ProductList{Items: items, Page: page, Limit: limit, Total: total}, nil
}

func (s *ProductService) GetByID(ctx context.Context, id string) (domain.Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProductService) Create(
	ctx context.Context,
	input domain.CreateProductInput,
) (domain.Product, error) {
	input.SKU = strings.ToUpper(strings.TrimSpace(input.SKU))
	input.Name = strings.TrimSpace(input.Name)
	if input.Description != nil {
		description := strings.TrimSpace(*input.Description)
		input.Description = &description
	}
	if input.SKU == "" || len(input.SKU) > 100 ||
		input.Name == "" || len(input.Name) > 255 ||
		input.PricePaise < 0 || input.InitialStock < 0 || input.InitialStock > 2147483647 {
		return domain.Product{}, ErrInvalidProduct
	}
	if input.Description != nil && len(*input.Description) > 4000 {
		return domain.Product{}, ErrInvalidProduct
	}
	return s.repo.Create(ctx, input)
}

func (s *ProductService) Update(
	ctx context.Context,
	id string,
	input domain.UpdateProductInput,
) (domain.Product, error) {
	if input.SKU == nil && input.Name == nil && input.Description == nil &&
		input.PricePaise == nil && input.IsActive == nil {
		return domain.Product{}, ErrInvalidProduct
	}
	if input.SKU != nil {
		v := strings.ToUpper(strings.TrimSpace(*input.SKU))
		if v == "" || len(v) > 100 {
			return domain.Product{}, ErrInvalidProduct
		}
		input.SKU = &v
	}
	if input.Name != nil {
		v := strings.TrimSpace(*input.Name)
		if v == "" || len(v) > 255 {
			return domain.Product{}, ErrInvalidProduct
		}
		input.Name = &v
	}
	if input.Description != nil {
		v := strings.TrimSpace(*input.Description)
		if len(v) > 4000 {
			return domain.Product{}, ErrInvalidProduct
		}
		input.Description = &v
	}
	if input.PricePaise != nil && *input.PricePaise < 0 {
		return domain.Product{}, ErrInvalidProduct
	}
	return s.repo.Update(ctx, id, input)
}

func (s *ProductService) Deactivate(ctx context.Context, id string) (domain.Product, error) {
	return s.repo.Deactivate(ctx, id)
}
