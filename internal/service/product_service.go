package service

import (
	"context"
	"strings"

	"github.com/manmohansharma21/orderflow/internal/domain"
	"github.com/manmohansharma21/orderflow/internal/repository"
)

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

	items, total, err := s.repo.List(
		ctx,
		strings.TrimSpace(search),
		limit,
		offset,
	)
	if err != nil {
		return domain.ProductList{}, err
	}

	return domain.ProductList{
		Items: items,
		Page:  page,
		Limit: limit,
		Total: total,
	}, nil
}

func (s *ProductService) GetByID(
	ctx context.Context,
	id string,
) (domain.Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Product{}, err
	}
	return product, nil
}
