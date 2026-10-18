package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/manmohansharma21/orderflow/internal/domain"
)

type InventoryRepository struct {
	pool *pgxpool.Pool
}

func NewInventoryRepository(pool *pgxpool.Pool) *InventoryRepository {
	return &InventoryRepository{pool: pool}
}

func (r *InventoryRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.BeginTx(ctx, pgx.TxOptions{})
}

func (r *InventoryRepository) GetAvailability(
	ctx context.Context,
	productID string,
) (domain.InventorySummary, error) {
	var item domain.InventorySummary
	err := r.pool.QueryRow(ctx, `
		SELECT p.id::text, p.sku, p.name,
		       i.on_hand_quantity, i.reserved_quantity,
		       i.on_hand_quantity - i.reserved_quantity
		FROM products AS p
		JOIN inventory AS i ON i.product_id = p.id
		WHERE p.id = $1 AND p.is_active = TRUE
	`, productID).Scan(
		&item.ProductID,
		&item.SKU,
		&item.ProductName,
		&item.OnHandQuantity,
		&item.ReservedQuantity,
		&item.AvailableQuantity,
	)
	return item, err
}

func (r *InventoryRepository) GetActiveInventoryForUpdate(
	ctx context.Context,
	tx pgx.Tx,
	productID string,
) (domain.InventorySummary, error) {
	var item domain.InventorySummary
	err := tx.QueryRow(ctx, `
		SELECT p.id::text, p.sku, p.name,
		       i.on_hand_quantity, i.reserved_quantity,
		       i.on_hand_quantity - i.reserved_quantity
		FROM products AS p
		JOIN inventory AS i ON i.product_id = p.id
		WHERE p.id = $1 AND p.is_active = TRUE
		FOR UPDATE OF i
	`, productID).Scan(
		&item.ProductID,
		&item.SKU,
		&item.ProductName,
		&item.OnHandQuantity,
		&item.ReservedQuantity,
		&item.AvailableQuantity,
	)
	return item, err
}

func (r *InventoryRepository) SetOnHandQuantity(
	ctx context.Context,
	tx pgx.Tx,
	productID string,
	newQuantity int,
) (int, int, error) {
	var onHand int
	var reserved int
	err := tx.QueryRow(ctx, `
		UPDATE inventory
		SET on_hand_quantity = $2,
		    updated_at = NOW()
		WHERE product_id = $1
		  AND $2 >= reserved_quantity
		RETURNING on_hand_quantity, reserved_quantity
	`, productID, newQuantity).Scan(&onHand, &reserved)
	return onHand, reserved, err
}

func (r *InventoryRepository) CreateMovement(
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

func (r *InventoryRepository) ListMovements(
	ctx context.Context,
	productID string,
	movementType string,
	limit int,
	offset int64,
) ([]domain.InventoryMovement, int64, error) {
	var total int64
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM inventory_movements AS m
		WHERE ($1::text = '' OR m.product_id::text = $1)
		  AND ($2::text = '' OR m.movement_type = $2)
	`, productID, movementType).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT m.id::text, m.product_id::text, p.sku, p.name,
		       m.movement_type, m.on_hand_delta, m.reserved_delta,
		       m.on_hand_after, m.reserved_after, m.reason,
		       COALESCE(m.created_by_user_id::text, ''),
		       COALESCE(m.reference_type, ''),
		       COALESCE(m.reference_id::text, ''), m.created_at
		FROM inventory_movements AS m
		JOIN products AS p ON p.id = m.product_id
		WHERE ($1::text = '' OR m.product_id::text = $1)
		  AND ($2::text = '' OR m.movement_type = $2)
		ORDER BY m.created_at DESC, m.id DESC
		LIMIT $3 OFFSET $4
	`, productID, movementType, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]domain.InventoryMovement, 0)
	for rows.Next() {
		var item domain.InventoryMovement
		if err := rows.Scan(
			&item.ID, &item.ProductID, &item.SKU, &item.ProductName,
			&item.MovementType, &item.OnHandDelta, &item.ReservedDelta,
			&item.OnHandAfter, &item.ReservedAfter, &item.Reason,
			&item.CreatedByUserID, &item.ReferenceType, &item.ReferenceID,
			&item.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
