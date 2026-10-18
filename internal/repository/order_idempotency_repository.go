package repository

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/manmohansharma21/orderflow/internal/domain"
)

// LockOrderIdempotency serializes identical user/key requests before lookup.
func (r *OrderRepository) LockOrderIdempotency(ctx context.Context, tx pgx.Tx, userID, key string) error {
	_, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, userID+":"+key)
	return err
}

func (r *OrderRepository) FindOrderByIdempotencyKey(ctx context.Context, tx pgx.Tx, userID, key string) (domain.CreateOrderResult, string, error) {
	var result domain.CreateOrderResult
	var requestHash string
	err := tx.QueryRow(ctx, `
        SELECT id::text, status, total_amount_paise, COALESCE(idempotency_request_hash, '')
        FROM orders WHERE user_id = $1::uuid AND idempotency_key = $2
    `, userID, key).Scan(&result.OrderID, &result.Status, &result.TotalAmountPaise, &requestHash)
	return result, requestHash, err
}

func (r *OrderRepository) CreateOrderWithIdempotency(ctx context.Context, tx pgx.Tx, userID string, totalPaise int64, key, requestHash string) (string, error) {
	var orderID string
	err := tx.QueryRow(ctx, `
        INSERT INTO orders (user_id, customer_email, total_amount_paise, idempotency_key, idempotency_request_hash)
        SELECT u.id, u.email, $2, $3, $4
        FROM users AS u WHERE u.id = $1::uuid
        RETURNING id::text
    `, userID, totalPaise, key, requestHash).Scan(&orderID)
	return orderID, err
}
