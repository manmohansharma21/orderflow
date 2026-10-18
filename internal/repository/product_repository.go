package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/manmohansharma21/orderflow/internal/domain"
)

type ProductRepository struct {
	pool *pgxpool.Pool
}

func NewProductRepository(pool *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{pool: pool}
}

func (r *ProductRepository) List(
	ctx context.Context,
	search string,
	limit, offset int,
) ([]domain.Product, int64, error) {
	var total int64

	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM products
		WHERE is_active = TRUE
		  AND ($1 = '' OR name ILIKE '%' || $1 || '%' OR sku ILIKE '%' || $1 || '%')
	`, search).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, sku, name, description, price_paise, is_active
		FROM products
		WHERE is_active = TRUE
		  AND ($1 = '' OR name ILIKE '%' || $1 || '%' OR sku ILIKE '%' || $1 || '%')
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, search, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]domain.Product, 0)

	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(
			&p.ID,
			&p.SKU,
			&p.Name,
			&p.Description,
			&p.PricePaise,
			&p.IsActive,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, p)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *ProductRepository) GetByID(
	ctx context.Context,
	id string,
) (domain.Product, error) {
	var p domain.Product

	err := r.pool.QueryRow(ctx, `
		SELECT id, sku, name, description, price_paise, is_active
		FROM products
		WHERE id = $1 AND is_active = TRUE
	`, id).Scan(
		&p.ID,
		&p.SKU,
		&p.Name,
		&p.Description,
		&p.PricePaise,
		&p.IsActive,
	)

	return p, err
}
