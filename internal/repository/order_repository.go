package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OrderRepository handles persistence operations related to orders.
type OrderRepository struct {
	pool *pgxpool.Pool
}

// NewOrderRepository creates an order repository.
func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{
		pool: pool,
	}
}

// BeginTx starts a database transaction.
func (r *OrderRepository) BeginTx(
	ctx context.Context,
) (pgx.Tx, error) {
	return r.pool.BeginTx(ctx, pgx.TxOptions{})
}

// GetActiveProduct retrieves an active product's name and price.
func (r *OrderRepository) GetActiveProduct(
	ctx context.Context,
	tx pgx.Tx,
	productID string,
) (string, int64, error) {
	var name string
	var pricePaise int64

	err := tx.QueryRow(ctx, `
		SELECT name, price_paise
		FROM products
		WHERE id = $1
		  AND is_active = TRUE
	`, productID).Scan(&name, &pricePaise)

	return name, pricePaise, err
}

// ReserveInventory reserves stock if enough inventory is available.
// func (r *OrderRepository) ReserveInventory(
// 	ctx context.Context,
// 	tx pgx.Tx,
// 	productID string,
// 	quantity int,
// ) error {
// 	_, err := tx.Exec(ctx, `
// 		UPDATE inventory
// 		SET reserved_quantity = reserved_quantity + $2,
// 		    updated_at = NOW()
// 		WHERE product_id = $1
// 		  AND on_hand_quantity - reserved_quantity >= $2
// 	`, productID, quantity)

// 	return err
// }

// ReserveInventory reserves stock if enough inventory is available.
func (r *OrderRepository) ReserveInventory(
	ctx context.Context,
	tx pgx.Tx,
	productID string,
	quantity int,
) (bool, error) {
	result, err := tx.Exec(ctx, `
		UPDATE inventory
		SET reserved_quantity = reserved_quantity + $2,
		    updated_at = NOW()
		WHERE product_id = $1
		  AND on_hand_quantity - reserved_quantity >= $2
	`, productID, quantity)
	if err != nil {
		return false, err
	}

	return result.RowsAffected() == 1, nil
}

// CreateOrder inserts an order and returns its ID.
func (r *OrderRepository) CreateOrder(
	ctx context.Context,
	tx pgx.Tx,
	customerEmail string,
	totalAmountPaise int64,
) (string, error) {
	var orderID string

	err := tx.QueryRow(ctx, `
		INSERT INTO orders (
			customer_email,
			total_amount_paise
		)
		VALUES ($1, $2)
		RETURNING id
	`, customerEmail, totalAmountPaise).Scan(&orderID)

	return orderID, err
}

// CreateOrderItem inserts an order item.
func (r *OrderRepository) CreateOrderItem(
	ctx context.Context,
	tx pgx.Tx,
	orderID string,
	productID string,
	productName string,
	quantity int,
	unitPricePaise int64,
	lineTotalPaise int64,
) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO order_items (
			order_id,
			product_id,
			product_name,
			quantity,
			unit_price_paise,
			line_total_paise
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`,
		orderID,
		productID,
		productName,
		quantity,
		unitPricePaise,
		lineTotalPaise,
	)

	return err
}
