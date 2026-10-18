package repository

import (
	"context"
	"errors"
	"time"

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

// ReserveInventory atomically reserves stock and returns post-update quantities.
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

// ListForUser returns a page of a user's orders and their items using one joined
// result query (plus a count query), avoiding an N+1 query per order.
func (r *OrderRepository) ListForUser(
	ctx context.Context,
	userID string,
	limit int,
	offset int64,
) ([]domain.Order, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM orders WHERE user_id = $1
	`, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT o.id::text, o.status, o.total_amount_paise, o.created_at, o.updated_at,
		       COALESCE(oi.product_id::text, ''), COALESCE(oi.product_name, ''),
		       COALESCE(oi.quantity, 0), COALESCE(oi.unit_price_paise, 0),
		       COALESCE(oi.line_total_paise, 0)
		FROM (
			SELECT id, status, total_amount_paise, created_at, updated_at
			FROM orders
			WHERE user_id = $1
			ORDER BY created_at DESC, id DESC
			LIMIT $2 OFFSET $3
		) AS o
		LEFT JOIN order_items AS oi ON oi.order_id = o.id
		ORDER BY o.created_at DESC, o.id DESC, oi.created_at, oi.id
	`, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	orders, err := collectOrders(rows)
	if err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

// GetForUser hides orders belonging to other users by filtering by both IDs.
func (r *OrderRepository) GetForUser(
	ctx context.Context,
	userID string,
	orderID string,
) (domain.Order, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT o.id::text, o.status, o.total_amount_paise, o.created_at, o.updated_at,
		       COALESCE(oi.product_id::text, ''), COALESCE(oi.product_name, ''),
		       COALESCE(oi.quantity, 0), COALESCE(oi.unit_price_paise, 0),
		       COALESCE(oi.line_total_paise, 0)
		FROM orders AS o
		LEFT JOIN order_items AS oi ON oi.order_id = o.id
		WHERE o.id = $1 AND o.user_id = $2
		ORDER BY oi.created_at, oi.id
	`, orderID, userID)
	if err != nil {
		return domain.Order{}, err
	}
	orders, err := collectOrders(rows)
	if err != nil {
		return domain.Order{}, err
	}
	if len(orders) == 0 {
		return domain.Order{}, pgx.ErrNoRows
	}
	return orders[0], nil
}

func collectOrders(rows pgx.Rows) ([]domain.Order, error) {
	defer rows.Close()
	orders := make([]domain.Order, 0)
	lastID := ""

	for rows.Next() {
		var orderID, status, productID, productName string
		var totalAmount, unitPrice, lineTotal int64
		var createdAt, updatedAt time.Time
		var quantity int
		if err := rows.Scan(
			&orderID, &status, &totalAmount, &createdAt, &updatedAt,
			&productID, &productName, &quantity, &unitPrice, &lineTotal,
		); err != nil {
			return nil, err
		}
		if orderID != lastID {
			orders = append(orders, domain.Order{
				ID: orderID, Status: status, TotalAmountPaise: totalAmount,
				CreatedAt: createdAt, UpdatedAt: updatedAt,
				Items: make([]domain.OrderItem, 0),
			})
			lastID = orderID
		}
		if productID != "" {
			orders[len(orders)-1].Items = append(orders[len(orders)-1].Items, domain.OrderItem{
				ProductID: productID, ProductName: productName, Quantity: quantity,
				UnitPricePaise: unitPrice, LineTotalPaise: lineTotal,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

type orderItemForCancellation struct {
	ProductID string
	Quantity  int
}

func (r *OrderRepository) LockOrderForCancellation(
	ctx context.Context,
	tx pgx.Tx,
	userID string,
	orderID string,
) (string, error) {
	var status string
	err := tx.QueryRow(ctx, `
		SELECT status FROM orders
		WHERE id = $1 AND user_id = $2
		FOR UPDATE
	`, orderID, userID).Scan(&status)
	return status, err
}

func (r *OrderRepository) GetItemsForCancellation(
	ctx context.Context,
	tx pgx.Tx,
	orderID string,
) ([]orderItemForCancellation, error) {
	rows, err := tx.Query(ctx, `
		SELECT product_id::text, quantity
		FROM order_items
		WHERE order_id = $1
		ORDER BY product_id
		FOR UPDATE
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]orderItemForCancellation, 0)
	for rows.Next() {
		var item orderItemForCancellation
		if err := rows.Scan(&item.ProductID, &item.Quantity); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *OrderRepository) ReleaseReservedInventory(
	ctx context.Context,
	tx pgx.Tx,
	productID string,
	quantity int,
) (int, int, error) {
	var onHand, reserved int
	err := tx.QueryRow(ctx, `
		UPDATE inventory
		SET reserved_quantity = reserved_quantity - $2,
		    updated_at = NOW()
		WHERE product_id = $1 AND reserved_quantity >= $2
		RETURNING on_hand_quantity, reserved_quantity
	`, productID, quantity).Scan(&onHand, &reserved)
	return onHand, reserved, err
}

func (r *OrderRepository) MarkCancelled(
	ctx context.Context,
	tx pgx.Tx,
	userID string,
	orderID string,
) error {
	_, err := tx.Exec(ctx, `
		UPDATE orders
		SET status = 'CANCELLED', updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND status = 'PENDING'
	`, orderID, userID)
	return err
}
