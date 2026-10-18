package service

import (
	"context"
	"errors"
	"strings"

	"github.com/manmohansharma21/orderflow/internal/domain"
	"github.com/manmohansharma21/orderflow/internal/repository"
)

var ErrInvalidProduct = errors.New("invalid product")

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
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Product{}, err
	}
	return product, nil
}

func (s *ProductService) Create(
	ctx context.Context,
	input domain.CreateProductInput,
) (domain.Product, error) {
	input.SKU = strings.TrimSpace(input.SKU)
	input.Name = strings.TrimSpace(input.Name)
	input.CreatedByUserID = strings.TrimSpace(input.CreatedByUserID)
	if input.Description != nil {
		description := strings.TrimSpace(*input.Description)
		input.Description = &description
	}
	if len(input.SKU) == 0 || len(input.SKU) > 100 ||
		len(input.Name) == 0 || len(input.Name) > 255 ||
		input.PricePaise < 0 || input.InitialStock < 0 || input.InitialStock > 1_000_000 ||
		input.CreatedByUserID == "" {
		return domain.Product{}, ErrInvalidProduct
	}
	return s.repo.Create(ctx, input)
}

func (s *ProductService) Update(
	ctx context.Context,
	id string,
	input domain.UpdateProductInput,
) (domain.Product, error) {
	id = strings.TrimSpace(id)
	if id == "" || (input.SKU == nil && input.Name == nil && input.Description == nil && input.PricePaise == nil && input.IsActive == nil) {
		return domain.Product{}, ErrInvalidProduct
	}
	if input.SKU != nil {
		value := strings.TrimSpace(*input.SKU)
		if len(value) == 0 || len(value) > 100 {
			return domain.Product{}, ErrInvalidProduct
		}
		input.SKU = &value
	}
	if input.Name != nil {
		value := strings.TrimSpace(*input.Name)
		if len(value) == 0 || len(value) > 255 {
			return domain.Product{}, ErrInvalidProduct
		}
		input.Name = &value
	}
	if input.Description != nil {
		value := strings.TrimSpace(*input.Description)
		input.Description = &value
	}
	if input.PricePaise != nil && *input.PricePaise < 0 {
		return domain.Product{}, ErrInvalidProduct
	}
	return s.repo.Update(ctx, id, input)
}

func (s *ProductService) Deactivate(ctx context.Context, id string) (domain.Product, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return domain.Product{}, ErrInvalidProduct
	}
	return s.repo.Deactivate(ctx, id)
}
