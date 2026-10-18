package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/manmohansharma21/orderflow/internal/domain"
)

var ErrDuplicateSKU = errors.New("product SKU already exists")

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
		if err := rows.Scan(&p.ID, &p.SKU, &p.Name, &p.Description, &p.PricePaise, &p.IsActive); err != nil {
			return nil, 0, err
		}
		items = append(items, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id string) (domain.Product, error) {
	var p domain.Product
	err := r.pool.QueryRow(ctx, `
		SELECT id, sku, name, description, price_paise, is_active
		FROM products
		WHERE id = $1 AND is_active = TRUE
	`, id).Scan(&p.ID, &p.SKU, &p.Name, &p.Description, &p.PricePaise, &p.IsActive)
	return p, err
}

func (r *ProductRepository) Create(
	ctx context.Context,
	input domain.CreateProductInput,
) (domain.Product, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.Product{}, err
	}
	defer tx.Rollback(ctx)

	var p domain.Product
	err = tx.QueryRow(ctx, `
		INSERT INTO products (sku, name, description, price_paise, is_active)
		VALUES ($1, $2, $3, $4, TRUE)
		RETURNING id, sku, name, description, price_paise, is_active
	`, input.SKU, input.Name, input.Description, input.PricePaise).Scan(
		&p.ID, &p.SKU, &p.Name, &p.Description, &p.PricePaise, &p.IsActive,
	)
	if err != nil {
		return domain.Product{}, mapDuplicateSKU(err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO inventory (product_id, on_hand_quantity, reserved_quantity, updated_at)
		VALUES ($1, $2, 0, NOW())
	`, p.ID, input.InitialStock)
	if err != nil {
		return domain.Product{}, err
	}

	if input.InitialStock > 0 {
		_, err = tx.Exec(ctx, `
			INSERT INTO inventory_movements (
				product_id, movement_type, on_hand_delta, reserved_delta,
				on_hand_after, reserved_after, reason, created_by_user_id
			)
			VALUES ($1, 'OPENING_BALANCE', $2, 0, $2, 0,
			        'Initial stock set during product creation', NULLIF($3, '')::uuid)
		`, p.ID, input.InitialStock, input.CreatedByUserID)
		if err != nil {
			return domain.Product{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Product{}, err
	}
	return p, nil
}

func (r *ProductRepository) Update(
	ctx context.Context,
	id string,
	input domain.UpdateProductInput,
) (domain.Product, error) {
	var p domain.Product
	err := r.pool.QueryRow(ctx, `
		UPDATE products
		SET sku = COALESCE($2, sku),
		    name = COALESCE($3, name),
		    description = COALESCE($4, description),
		    price_paise = COALESCE($5, price_paise),
		    is_active = COALESCE($6, is_active),
		    updated_at = NOW()
		WHERE id = $1
		RETURNING id, sku, name, description, price_paise, is_active
	`, id, input.SKU, input.Name, input.Description, input.PricePaise, input.IsActive).Scan(
		&p.ID, &p.SKU, &p.Name, &p.Description, &p.PricePaise, &p.IsActive,
	)
	if err != nil {
		return domain.Product{}, mapDuplicateSKU(err)
	}
	return p, nil
}

func (r *ProductRepository) Deactivate(ctx context.Context, id string) (domain.Product, error) {
	inactive := false
	return r.Update(ctx, id, domain.UpdateProductInput{IsActive: &inactive})
}

func mapDuplicateSKU(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && strings.Contains(pgErr.ConstraintName, "sku") {
		return ErrDuplicateSKU
	}
	return err
}
