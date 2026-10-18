package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/manmohansharma21/orderflow/internal/domain"
	"github.com/manmohansharma21/orderflow/internal/repository"
)

var ErrInvalidProduct = errors.New("invalid product")

type ProductService struct {
	repo     repository.ProductRepositoryAdapter
	cache    ProductCache
	cacheTTL time.Duration
}

func NewProductService(repo *repository.ProductRepository) *ProductService {
	return NewProductServiceWithCache(repo, nil, time.Minute)
}

// NewProductServiceWithCache enables best-effort Redis caching. Database reads
// remain the source of truth and cache failures transparently fall back to SQL.
func NewProductServiceWithCache(repo repository.ProductRepositoryAdapter, productCache ProductCache, ttl time.Duration) *ProductService {
	if ttl <= 0 {
		ttl = time.Minute
	}
	return &ProductService{repo: repo, cache: productCache, cacheTTL: ttl}
}

func (s *ProductService) List(ctx context.Context, search string, page, limit int) (domain.ProductList, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	search = strings.TrimSpace(search)
	offset := (page - 1) * limit
	key := s.cacheKey(ctx, fmt.Sprintf("list:%x", sha256.Sum256([]byte(fmt.Sprintf("%s|%d|%d", search, page, limit)))))
	if key != "" {
		var cached domain.ProductList
		if data, err := s.cache.Get(ctx, key); err == nil && json.Unmarshal(data, &cached) == nil {
			return cached, nil
		}
	}
	items, total, err := s.repo.List(ctx, search, limit, offset)
	if err != nil {
		return domain.ProductList{}, err
	}
	result := domain.ProductList{Items: items, Page: page, Limit: limit, Total: total}
	s.cacheSet(ctx, key, result)
	return result, nil
}

func (s *ProductService) GetByID(ctx context.Context, id string) (domain.Product, error) {
	id = strings.TrimSpace(id)
	key := s.cacheKey(ctx, "item:"+id)
	if key != "" {
		var cached domain.Product
		if data, err := s.cache.Get(ctx, key); err == nil && json.Unmarshal(data, &cached) == nil {
			return cached, nil
		}
	}
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Product{}, err
	}
	s.cacheSet(ctx, key, product)
	return product, nil
}

func (s *ProductService) Create(ctx context.Context, input domain.CreateProductInput) (domain.Product, error) {
	input.SKU = strings.TrimSpace(input.SKU)
	input.Name = strings.TrimSpace(input.Name)
	input.CreatedByUserID = strings.TrimSpace(input.CreatedByUserID)
	if input.Description != nil {
		description := strings.TrimSpace(*input.Description)
		input.Description = &description
	}
	if len(input.SKU) == 0 || len(input.SKU) > 100 || len(input.Name) == 0 || len(input.Name) > 255 || input.PricePaise < 0 || input.InitialStock < 0 || input.InitialStock > 1_000_000 || input.CreatedByUserID == "" {
		return domain.Product{}, ErrInvalidProduct
	}
	product, err := s.repo.Create(ctx, input)
	if err != nil {
		return domain.Product{}, err
	}
	s.invalidate(ctx)
	return product, nil
}

func (s *ProductService) Update(ctx context.Context, id string, input domain.UpdateProductInput) (domain.Product, error) {
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
	product, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return domain.Product{}, err
	}
	s.invalidate(ctx)
	return product, nil
}

func (s *ProductService) Deactivate(ctx context.Context, id string) (domain.Product, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return domain.Product{}, ErrInvalidProduct
	}
	product, err := s.repo.Deactivate(ctx, id)
	if err != nil {
		return domain.Product{}, err
	}
	s.invalidate(ctx)
	return product, nil
}

func (s *ProductService) cacheKey(ctx context.Context, suffix string) string {
	if s.cache == nil {
		return ""
	}
	cacheCtx, cancel := context.WithTimeout(ctx, 250*time.Millisecond)
	defer cancel()
	version, err := s.cache.Version(cacheCtx)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("orderflow:products:v%d:%s", version, suffix)
}

func (s *ProductService) cacheSet(ctx context.Context, key string, value any) {
	if key == "" || s.cache == nil {
		return
	}
	data, err := json.Marshal(value)
	if err != nil {
		return
	}
	cacheCtx, cancel := context.WithTimeout(ctx, 250*time.Millisecond)
	defer cancel()
	_ = s.cache.Set(cacheCtx, key, data, s.cacheTTL)
}

func (s *ProductService) invalidate(ctx context.Context) {
	if s.cache == nil {
		return
	}
	cacheCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second)
	defer cancel()
	if err := s.cache.InvalidateProducts(cacheCtx); err != nil {
		log.Printf("WARN: product cache invalidation failed; entries expire by TTL: %v", err)
	}
}
