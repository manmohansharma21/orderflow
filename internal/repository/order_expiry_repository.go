package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/manmohansharma21/orderflow/internal/domain"
)

// ExpireUnpaidPendingOrders expires only PENDING orders with no unresolved or successful payment.
// Orders with a PENDING payment are intentionally excluded until the provider result is known.
func (r *OrderRepository) ExpireUnpaidPendingOrders(ctx context.Context, olderThanMinutes, limit int) (int, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
        SELECT o.id::text
        FROM orders AS o
        WHERE o.status = 'PENDING'
          AND o.created_at < NOW() - ($1 * INTERVAL '1 minute')
        ORDER BY o.created_at, o.id
        LIMIT $2
        FOR UPDATE OF o SKIP LOCKED
    `, olderThanMinutes, limit)
	if err != nil {
		return 0, err
	}
	var orderIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return 0, err
		}
		orderIDs = append(orderIDs, id)
	}
	rowsErr := rows.Err()
	rows.Close()
	if rowsErr != nil {
		return 0, rowsErr
	}

	expired := 0
	for _, orderID := range orderIDs {
		// Re-check status/payment after acquiring the row locks. This second
		// statement gets a fresh READ COMMITTED snapshot, avoiding expiry based
		// on a stale candidate snapshot racing with payment initiation.
		var status string
		if err := tx.QueryRow(ctx, `SELECT status FROM orders WHERE id=$1::uuid FOR UPDATE`, orderID).Scan(&status); err != nil {
			return 0, err
		}
		if status != "PENDING" {
			continue
		}
		var hasUnresolvedOrSuccessfulPayment bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM payments WHERE order_id=$1::uuid AND status IN ('PENDING','SUCCESS'))`, orderID).Scan(&hasUnresolvedOrSuccessfulPayment); err != nil {
			return 0, err
		}
		if hasUnresolvedOrSuccessfulPayment {
			continue
		}
		items, err := r.GetItemsForCancellation(ctx, tx, orderID)
		if err != nil {
			return 0, err
		}
		for _, item := range items {
			onHand, reserved, err := r.ReleaseReservedInventory(ctx, tx, item.ProductID, item.Quantity)
			if errors.Is(err, pgx.ErrNoRows) {
				return 0, ErrOrderExpiryInventoryConflict
			}
			if err != nil {
				return 0, err
			}
			if err := r.RecordInventoryMovement(ctx, tx, domain.InventoryMovementInput{ProductID: item.ProductID, MovementType: "ORDER_RELEASED", OnHandDelta: 0, ReservedDelta: -item.Quantity, OnHandAfter: onHand, ReservedAfter: reserved, Reason: "Reserved stock released after unpaid order expiry", ReferenceType: "ORDER", ReferenceID: orderID}); err != nil {
				return 0, err
			}
		}
		result, err := tx.Exec(ctx, `UPDATE orders SET status='EXPIRED', updated_at=NOW() WHERE id=$1::uuid AND status='PENDING'`, orderID)
		if err != nil {
			return 0, err
		}
		if result.RowsAffected() == 1 {
			expired++
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return expired, nil
}
