package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/manmohansharma21/orderflow/internal/domain"
)

// OrderRepository handles persistence operations related to orders.
type OrderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

func (r *OrderRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.BeginTx(ctx, pgx.TxOptions{})
}

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
		WHERE id = $1 AND is_active = TRUE
	`, productID).Scan(&name, &pricePaise)
	return name, pricePaise, err
}

// ReserveInventory atomically reserves stock and returns the post-update quantities.
func (r *OrderRepository) ReserveInventory(
	ctx context.Context,
	tx pgx.Tx,
	productID string,
	quantity int,
) (bool, int, int, error) {
	var onHand int
	var reserved int
	err := tx.QueryRow(ctx, `
		UPDATE inventory
		SET reserved_quantity = reserved_quantity + $2,
		    updated_at = NOW()
		WHERE product_id = $1
		  AND on_hand_quantity - reserved_quantity >= $2
		RETURNING on_hand_quantity, reserved_quantity
	`, productID, quantity).Scan(&onHand, &reserved)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, 0, 0, nil
		}
		return false, 0, 0, err
	}
	return true, onHand, reserved, nil
}

func (r *OrderRepository) CreateOrder(
	ctx context.Context,
	tx pgx.Tx,
	userID string,
	totalAmountPaise int64,
) (string, error) {
	var orderID string
	err := tx.QueryRow(ctx, `
		INSERT INTO orders (user_id, customer_email, total_amount_paise)
		SELECT u.id, u.email, $2
		FROM users AS u
		WHERE u.id = $1
		RETURNING id
	`, userID, totalAmountPaise).Scan(&orderID)
	return orderID, err
}

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
			order_id, product_id, product_name, quantity,
			unit_price_paise, line_total_paise
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, orderID, productID, productName, quantity, unitPricePaise, lineTotalPaise)
	return err
}

func (r *OrderRepository) RecordInventoryMovement(
	ctx context.Context,
	tx pgx.Tx,
	input domain.InventoryMovementInput,
) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO inventory_movements (
			product_id, movement_type, on_hand_delta, reserved_delta,
			on_hand_after, reserved_after, reason, created_by_user_id,
			reference_type, reference_id
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			NULLIF($8, '')::uuid, NULLIF($9, ''), NULLIF($10, '')::uuid
		)
	`, input.ProductID, input.MovementType, input.OnHandDelta,
		input.ReservedDelta, input.OnHandAfter, input.ReservedAfter,
		input.Reason, input.CreatedByUserID, input.ReferenceType, input.ReferenceID)
	return err
}
