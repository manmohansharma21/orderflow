package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/manmohansharma21/orderflow/internal/domain"
)

var (
	ErrPaymentIdempotencyConflict = errors.New("payment idempotency key conflicts with an existing payment")
	ErrPaymentStateConflict       = errors.New("payment or order state does not allow this operation")
)

type PaymentRepository struct {
	pool *pgxpool.Pool
}

func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

func scanPayment(row pgx.Row) (domain.Payment, error) {
	var p domain.Payment
	err := row.Scan(
		&p.ID, &p.OrderID, &p.Provider, &p.ProviderPaymentID,
		&p.Status, &p.AmountPaise, &p.CreatedAt, &p.UpdatedAt,
	)
	return p, err
}

const paymentColumns = `
	id::text, order_id::text, provider,
	COALESCE(provider_payment_id, ''), status, amount_paise,
	created_at, updated_at
`

func (r *PaymentRepository) CreateOrGetPending(
	ctx context.Context,
	userID, orderID, idempotencyKey string,
) (domain.Payment, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.Payment{}, err
	}
	defer tx.Rollback(ctx)

	// Lock the owned order first. This serializes attempts to start a payment
	// for the same order and prevents concurrent duplicate pending payments.
	var orderStatus string
	var amount int64
	var ownerID string
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(user_id::text, ''), status, total_amount_paise
		FROM orders
		WHERE id = $1
		FOR UPDATE
	`, orderID).Scan(&ownerID, &orderStatus, &amount)
	if err != nil {
		return domain.Payment{}, err
	}
	if ownerID != userID {
		return domain.Payment{}, pgx.ErrNoRows
	}

	// A key is scoped to this customer's request. Reusing it for another
	// order is rejected instead of silently returning an unrelated payment.
	existing, err := scanPayment(tx.QueryRow(ctx, `
		SELECT `+paymentColumns+`
		FROM payments
		WHERE idempotency_key = $1
	`, idempotencyKey))
	if err == nil {
		if existing.OrderID != orderID {
			return domain.Payment{}, ErrPaymentIdempotencyConflict
		}
		return existing, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Payment{}, err
	}

	if orderStatus != "PENDING" {
		return domain.Payment{}, ErrPaymentStateConflict
	}

	// Do not create multiple simultaneously pending payment attempts for one
	// order. A retry with a different key still returns the active attempt.
	existing, err = scanPayment(tx.QueryRow(ctx, `
		SELECT `+paymentColumns+`
		FROM payments
		WHERE order_id = $1 AND status = 'PENDING'
		ORDER BY created_at DESC, id DESC
		LIMIT 1
		FOR UPDATE
	`, orderID))
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Payment{}, err
	}

	p, err := scanPayment(tx.QueryRow(ctx, `
		INSERT INTO payments (order_id, provider, status, amount_paise, idempotency_key)
		VALUES ($1, 'MOCK', 'PENDING', $2, $3)
		RETURNING `+paymentColumns,
		orderID, amount, idempotencyKey,
	))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.Payment{}, ErrPaymentIdempotencyConflict
		}
		return domain.Payment{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Payment{}, err
	}
	return p, nil
}

func (r *PaymentRepository) ProcessMockCallback(
	ctx context.Context,
	paymentID, outcome, eventID string,
) (domain.Payment, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.Payment{}, err
	}
	defer tx.Rollback(ctx)

	// Read the order ID without locking first, then lock the order before the
	// payment. This matches the initiation flow's lock order and reduces deadlocks.
	var orderID string
	err = tx.QueryRow(ctx, `SELECT order_id::text FROM payments WHERE id = $1`, paymentID).Scan(&orderID)
	if err != nil {
		return domain.Payment{}, err
	}

	var orderStatus string
	err = tx.QueryRow(ctx, `
		SELECT status
		FROM orders
		WHERE id = $1
		FOR UPDATE
	`, orderID).Scan(&orderStatus)
	if err != nil {
		return domain.Payment{}, err
	}

	p, err := scanPayment(tx.QueryRow(ctx, `
		SELECT `+paymentColumns+`
		FROM payments
		WHERE id = $1
		FOR UPDATE
	`, paymentID))
	if err != nil {
		return domain.Payment{}, err
	}

	// Replayed callbacks with the same terminal result are safe and idempotent.
	if p.Status == outcome {
		return p, nil
	}
	if p.Status != "PENDING" || orderStatus != "PENDING" {
		return domain.Payment{}, ErrPaymentStateConflict
	}

	var usedBy string
	err = tx.QueryRow(ctx, `
		SELECT id::text FROM payments
		WHERE callback_event_id = $1 AND id <> $2
	`, eventID, paymentID).Scan(&usedBy)
	if err == nil {
		return domain.Payment{}, ErrPaymentIdempotencyConflict
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Payment{}, err
	}

	// On success, confirm the order but keep its stock reserved until a later
	// fulfillment step. On failure, release all order reservations atomically.
	if outcome == "FAILED" {
		items, err := r.getOrderItemsForUpdate(ctx, tx, p.OrderID)
		if err != nil {
			return domain.Payment{}, err
		}
		if len(items) == 0 {
			return domain.Payment{}, ErrPaymentStateConflict
		}
		for _, item := range items {
			var onHandAfter, reservedAfter int
			err := tx.QueryRow(ctx, `
				UPDATE inventory
				SET reserved_quantity = reserved_quantity - $2,
				    updated_at = NOW()
				WHERE product_id = $1 AND reserved_quantity >= $2
				RETURNING on_hand_quantity, reserved_quantity
			`, item.ProductID, item.Quantity).Scan(&onHandAfter, &reservedAfter)
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.Payment{}, ErrPaymentStateConflict
			}
			if err != nil {
				return domain.Payment{}, err
			}
			_, err = tx.Exec(ctx, `
				INSERT INTO inventory_movements (
					product_id, movement_type, on_hand_delta, reserved_delta,
					on_hand_after, reserved_after, reason, reference_type, reference_id
				)
				VALUES ($1, 'ORDER_RELEASED', 0, $2, $3, $4,
				        'Reserved stock released after payment failure', 'ORDER', $5)
			`, item.ProductID, -item.Quantity, onHandAfter, reservedAfter, p.OrderID)
			if err != nil {
				return domain.Payment{}, err
			}
		}
	}

	providerID := fmt.Sprintf("mock_%s", paymentID)
	_, err = tx.Exec(ctx, `
		UPDATE payments
		SET status = $2,
		    provider_payment_id = COALESCE(provider_payment_id, $3),
		    callback_event_id = $4,
		    updated_at = NOW()
		WHERE id = $1 AND status = 'PENDING'
	`, paymentID, outcome, providerID, eventID)
	if err != nil {
		return domain.Payment{}, err
	}

	nextOrderStatus := "CONFIRMED"
	if outcome == "FAILED" {
		nextOrderStatus = "FAILED"
	}
	result, err := tx.Exec(ctx, `
		UPDATE orders
		SET status = $2, updated_at = NOW()
		WHERE id = $1 AND status = 'PENDING'
	`, p.OrderID, nextOrderStatus)
	if err != nil {
		return domain.Payment{}, err
	}
	if result.RowsAffected() != 1 {
		return domain.Payment{}, ErrPaymentStateConflict
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Payment{}, err
	}
	return r.GetByID(ctx, paymentID)
}

type cancellableOrderItem struct {
	ProductID string
	Quantity  int
}

func (r *PaymentRepository) getOrderItemsForUpdate(
	ctx context.Context,
	tx pgx.Tx,
	orderID string,
) ([]cancellableOrderItem, error) {
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

	items := make([]cancellableOrderItem, 0)
	for rows.Next() {
		var item cancellableOrderItem
		if err := rows.Scan(&item.ProductID, &item.Quantity); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PaymentRepository) GetByID(ctx context.Context, paymentID string) (domain.Payment, error) {
	return scanPayment(r.pool.QueryRow(ctx, `
		SELECT `+paymentColumns+` FROM payments WHERE id = $1
	`, paymentID))
}

func (r *PaymentRepository) ReconciliationReport(
	ctx context.Context,
	olderThanMinutes int,
) (domain.PaymentReconciliationReport, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT issue_type, payment_id, order_id, payment_status, order_status,
		       amount_paise, created_at, details
		FROM (
			SELECT
				'STALE_PENDING_PAYMENT'::text AS issue_type,
				p.id::text AS payment_id,
				o.id::text AS order_id,
				p.status::text AS payment_status,
				o.status::text AS order_status,
				p.amount_paise AS amount_paise,
				p.created_at AS created_at,
				'Pending payment is older than the configured threshold'::text AS details
			FROM payments p
			JOIN orders o ON o.id = p.order_id
			WHERE p.status = 'PENDING'
			  AND p.created_at < NOW() - ($1 * INTERVAL '1 minute')

			UNION ALL

			SELECT
				'CONFIRMED_ORDER_WITHOUT_SUCCESSFUL_PAYMENT',
				''::text,
				o.id::text,
				''::text,
				o.status::text,
				o.total_amount_paise,
				o.created_at,
				'Order is CONFIRMED but has no SUCCESS payment'
			FROM orders o
			WHERE o.status = 'CONFIRMED'
			  AND NOT EXISTS (
				SELECT 1 FROM payments p
				WHERE p.order_id = o.id AND p.status = 'SUCCESS'
			  )

			UNION ALL

			SELECT
				'FAILED_PAYMENT_ORDER_PENDING',
				p.id::text,
				o.id::text,
				p.status::text,
				o.status::text,
				p.amount_paise,
				p.created_at,
				'Payment is FAILED while the order is still PENDING'
			FROM payments p
			JOIN orders o ON o.id = p.order_id
			WHERE p.status = 'FAILED' AND o.status = 'PENDING'
		) AS issues
		ORDER BY created_at ASC
	`, olderThanMinutes)
	if err != nil {
		return domain.PaymentReconciliationReport{}, err
	}
	defer rows.Close()

	issues := make([]domain.PaymentReconciliationIssue, 0)
	for rows.Next() {
		var item domain.PaymentReconciliationIssue
		if err := rows.Scan(
			&item.IssueType, &item.PaymentID, &item.OrderID,
			&item.PaymentStatus, &item.OrderStatus, &item.AmountPaise,
			&item.CreatedAt, &item.Details,
		); err != nil {
			return domain.PaymentReconciliationReport{}, err
		}
		issues = append(issues, item)
	}
	if err := rows.Err(); err != nil {
		return domain.PaymentReconciliationReport{}, err
	}
	return domain.PaymentReconciliationReport{
		GeneratedAt: time.Now().UTC(), OlderThanMinutes: olderThanMinutes,
		IssueCount: len(issues), Issues: issues,
	}, nil
}
