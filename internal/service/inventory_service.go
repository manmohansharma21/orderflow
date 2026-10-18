package service

import (
	"context"
	"errors"
	"math"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/manmohansharma21/orderflow/internal/domain"
	"github.com/manmohansharma21/orderflow/internal/repository"
)

var (
	ErrInvalidInventoryRequest = errors.New("invalid inventory request")
	ErrInventoryConflict       = errors.New("inventory quantity conflicts with reserved stock")
)

const maxInventoryChange = 1_000_000

var supportedMovementTypes = map[string]bool{
	"OPENING_BALANCE": true,
	"RECEIPT":         true,
	"ADJUSTMENT":      true,
	"ORDER_RESERVED":  true,
	"ORDER_RELEASED":  true,
	"ORDER_FULFILLED": true,
	"REFUND_RESTOCK":  true,
}

type InventoryService struct {
	repo *repository.InventoryRepository
}

func NewInventoryService(repo *repository.InventoryRepository) *InventoryService {
	return &InventoryService{repo: repo}
}

func (s *InventoryService) GetAvailability(
	ctx context.Context,
	productID string,
) (domain.InventorySummary, error) {
	return s.repo.GetAvailability(ctx, productID)
}

func (s *InventoryService) ReceiveStock(
	ctx context.Context,
	input domain.InventoryMovementInput,
) (domain.InventorySummary, error) {
	input.MovementType = "RECEIPT"
	if input.OnHandDelta < 1 || input.OnHandDelta > maxInventoryChange {
		return domain.InventorySummary{}, ErrInvalidInventoryRequest
	}
	return s.changeOnHand(ctx, input)
}

func (s *InventoryService) AdjustStock(
	ctx context.Context,
	input domain.InventoryMovementInput,
) (domain.InventorySummary, error) {
	input.MovementType = "ADJUSTMENT"
	if input.OnHandDelta == 0 || input.OnHandDelta < -maxInventoryChange || input.OnHandDelta > maxInventoryChange {
		return domain.InventorySummary{}, ErrInvalidInventoryRequest
	}
	return s.changeOnHand(ctx, input)
}

func (s *InventoryService) changeOnHand(
	ctx context.Context,
	input domain.InventoryMovementInput,
) (domain.InventorySummary, error) {
	input.ProductID = strings.TrimSpace(input.ProductID)
	input.CreatedByUserID = strings.TrimSpace(input.CreatedByUserID)
	input.Reason = strings.TrimSpace(input.Reason)
	if input.ProductID == "" || input.CreatedByUserID == "" || len(input.Reason) < 3 || len(input.Reason) > 500 {
		return domain.InventorySummary{}, ErrInvalidInventoryRequest
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return domain.InventorySummary{}, err
	}
	defer tx.Rollback(ctx)

	current, err := s.repo.GetActiveInventoryForUpdate(ctx, tx, input.ProductID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.InventorySummary{}, ErrProductNotFound
		}
		return domain.InventorySummary{}, err
	}

	newOnHand64 := int64(current.OnHandQuantity) + int64(input.OnHandDelta)
	if newOnHand64 < int64(current.ReservedQuantity) || newOnHand64 < 0 || newOnHand64 > math.MaxInt32 {
		return domain.InventorySummary{}, ErrInventoryConflict
	}
	newOnHand := int(newOnHand64)

	onHandAfter, reservedAfter, err := s.repo.SetOnHandQuantity(ctx, tx, input.ProductID, newOnHand)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.InventorySummary{}, ErrInventoryConflict
		}
		return domain.InventorySummary{}, err
	}

	input.OnHandAfter = onHandAfter
	input.ReservedAfter = reservedAfter
	input.ReservedDelta = 0
	if err := s.repo.CreateMovement(ctx, tx, input); err != nil {
		return domain.InventorySummary{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.InventorySummary{}, err
	}

	current.OnHandQuantity = onHandAfter
	current.ReservedQuantity = reservedAfter
	current.AvailableQuantity = onHandAfter - reservedAfter
	return current, nil
}

func (s *InventoryService) ListMovements(
	ctx context.Context,
	productID string,
	movementType string,
	page int,
	limit int,
) (domain.InventoryMovementList, error) {
	productID = strings.TrimSpace(productID)
	movementType = strings.ToUpper(strings.TrimSpace(movementType))
	if movementType != "" && !supportedMovementTypes[movementType] {
		return domain.InventoryMovementList{}, ErrInvalidInventoryRequest
	}
	if page < 1 {
		page = 1
	}
	if page > 1_000_000 {
		return domain.InventoryMovementList{}, ErrInvalidInventoryRequest
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := int64(page-1) * int64(limit)
	items, total, err := s.repo.ListMovements(ctx, productID, movementType, limit, offset)
	if err != nil {
		return domain.InventoryMovementList{}, err
	}
	return domain.InventoryMovementList{Items: items, Page: page, Limit: limit, Total: total}, nil
}
