package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/manmohansharma21/orderflow/internal/domain"
	"github.com/manmohansharma21/orderflow/internal/repository"
)

var (
	ErrProductNotFound          = errors.New("product not found")
	ErrUserNotFound             = errors.New("user not found")
	ErrInsufficientStock        = errors.New("insufficient stock")
	ErrInvalidOrder             = errors.New("invalid order")
	ErrOrderNotFound            = errors.New("order not found")
	ErrOrderNotCancellable      = errors.New("order is not cancellable")
	ErrOrderInventoryConflict   = errors.New("reserved inventory cannot be released")
	ErrInvalidOrderQuery        = errors.New("invalid order query")
	ErrOrderIdempotencyConflict = errors.New("order idempotency key conflicts with a different request")
)

const maxOrderPage = 1_000_000

type OrderService struct {
	repo *repository.OrderRepository
}

func NewOrderService(repo *repository.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) CreateOrder(
	ctx context.Context,
	req domain.CreateOrderRequest,
) (domain.CreateOrderResult, error) {
	req.UserID = strings.TrimSpace(req.UserID)
	req.IdempotencyKey = strings.TrimSpace(req.IdempotencyKey)
	if req.UserID == "" || len(req.IdempotencyKey) > 200 || (req.IdempotencyKey != "" && len(req.IdempotencyKey) < 8) {
		return domain.CreateOrderResult{}, ErrInvalidOrder
	}

	items := append([]domain.CreateOrderLineRequest(nil), req.Items...)
	if len(items) == 0 && strings.TrimSpace(req.ProductID) != "" {
		items = append(items, domain.CreateOrderLineRequest{ProductID: req.ProductID, Quantity: req.Quantity})
	} else if len(items) > 0 && (strings.TrimSpace(req.ProductID) != "" || req.Quantity != 0) {
		return domain.CreateOrderResult{}, ErrInvalidOrder
	}
	if len(items) == 0 || len(items) > 50 {
		return domain.CreateOrderResult{}, ErrInvalidOrder
	}

	seen := make(map[string]struct{}, len(items))
	for i := range items {
		items[i].ProductID = strings.TrimSpace(items[i].ProductID)
		if items[i].ProductID == "" || items[i].Quantity < 1 || items[i].Quantity > 10000 {
			return domain.CreateOrderResult{}, ErrInvalidOrder
		}
		if _, duplicate := seen[items[i].ProductID]; duplicate {
			return domain.CreateOrderResult{}, ErrInvalidOrder
		}
		seen[items[i].ProductID] = struct{}{}
	}
	// Stable lock ordering reduces deadlocks when multi-product orders overlap.
	sort.Slice(items, func(i, j int) bool { return items[i].ProductID < items[j].ProductID })
	canonical, err := json.Marshal(items)
	if err != nil {
		return domain.CreateOrderResult{}, err
	}
	hash := sha256.Sum256(canonical)
	requestHash := hex.EncodeToString(hash[:])

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return domain.CreateOrderResult{}, err
	}
	defer tx.Rollback(ctx)

	if req.IdempotencyKey != "" {
		if err := s.repo.LockOrderIdempotency(ctx, tx, req.UserID, req.IdempotencyKey); err != nil {
			return domain.CreateOrderResult{}, err
		}
		existing, existingHash, err := s.repo.FindOrderByIdempotencyKey(ctx, tx, req.UserID, req.IdempotencyKey)
		if err == nil {
			if existingHash != requestHash {
				return domain.CreateOrderResult{}, ErrOrderIdempotencyConflict
			}
			existing.Replayed = true
			return existing, nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return domain.CreateOrderResult{}, err
		}
	}

	type resolvedLine struct {
		productID string
		name      string
		quantity  int
		unitPrice int64
		lineTotal int64
		onHand    int
		reserved  int
	}
	resolved := make([]resolvedLine, 0, len(items))
	var totalPaise int64
	for _, item := range items {
		productName, pricePaise, err := s.repo.GetActiveProduct(ctx, tx, item.ProductID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.CreateOrderResult{}, ErrProductNotFound
			}
			return domain.CreateOrderResult{}, err
		}
		if pricePaise > 0 && int64(item.Quantity) > math.MaxInt64/pricePaise {
			return domain.CreateOrderResult{}, ErrInvalidOrder
		}
		lineTotal := pricePaise * int64(item.Quantity)
		if lineTotal > math.MaxInt64-totalPaise {
			return domain.CreateOrderResult{}, ErrInvalidOrder
		}
		reserved, onHandAfter, reservedAfter, err := s.repo.ReserveInventory(ctx, tx, item.ProductID, item.Quantity)
		if err != nil {
			return domain.CreateOrderResult{}, err
		}
		if !reserved {
			return domain.CreateOrderResult{}, ErrInsufficientStock
		}
		totalPaise += lineTotal
		resolved = append(resolved, resolvedLine{
			productID: item.ProductID, name: productName, quantity: item.Quantity,
			unitPrice: pricePaise, lineTotal: lineTotal,
			onHand: onHandAfter, reserved: reservedAfter,
		})
	}

	var orderID string
	if req.IdempotencyKey != "" {
		orderID, err = s.repo.CreateOrderWithIdempotency(ctx, tx, req.UserID, totalPaise, req.IdempotencyKey, requestHash)
	} else {
		orderID, err = s.repo.CreateOrder(ctx, tx, req.UserID, totalPaise)
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.CreateOrderResult{}, ErrUserNotFound
		}
		return domain.CreateOrderResult{}, err
	}

	for _, item := range resolved {
		if err := s.repo.RecordInventoryMovement(ctx, tx, domain.InventoryMovementInput{
			ProductID: item.productID, MovementType: "ORDER_RESERVED",
			OnHandDelta: 0, ReservedDelta: item.quantity,
			OnHandAfter: item.onHand, ReservedAfter: item.reserved,
			Reason: "Stock reserved for order creation", CreatedByUserID: req.UserID,
			ReferenceType: "ORDER", ReferenceID: orderID,
		}); err != nil {
			return domain.CreateOrderResult{}, err
		}
		if err := s.repo.CreateOrderItem(ctx, tx, orderID, item.productID, item.name, item.quantity, item.unitPrice, item.lineTotal); err != nil {
			return domain.CreateOrderResult{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.CreateOrderResult{}, err
	}
	return domain.CreateOrderResult{OrderID: orderID, Status: "PENDING", TotalAmountPaise: totalPaise}, nil
}

func (s *OrderService) ListForUser(
	ctx context.Context,
	userID string,
	page int,
	limit int,
) (domain.OrderList, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return domain.OrderList{}, ErrInvalidOrderQuery
	}
	if page < 1 {
		page = 1
	}
	if page > maxOrderPage {
		return domain.OrderList{}, ErrInvalidOrderQuery
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := int64(page-1) * int64(limit)
	items, total, err := s.repo.ListForUser(ctx, userID, limit, offset)
	if err != nil {
		return domain.OrderList{}, err
	}
	return domain.OrderList{Items: items, Page: page, Limit: limit, Total: total}, nil
}

func (s *OrderService) GetForUser(
	ctx context.Context,
	userID string,
	orderID string,
) (domain.Order, error) {
	userID = strings.TrimSpace(userID)
	orderID = strings.TrimSpace(orderID)
	if userID == "" || orderID == "" {
		return domain.Order{}, ErrInvalidOrderQuery
	}
	order, err := s.repo.GetForUser(ctx, userID, orderID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Order{}, ErrOrderNotFound
	}
	return order, err
}

func (s *OrderService) CancelForUser(
	ctx context.Context,
	userID string,
	orderID string,
) (domain.Order, error) {
	userID = strings.TrimSpace(userID)
	orderID = strings.TrimSpace(orderID)
	if userID == "" || orderID == "" {
		return domain.Order{}, ErrInvalidOrderQuery
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return domain.Order{}, err
	}
	defer tx.Rollback(ctx)

	status, err := s.repo.LockOrderForCancellation(ctx, tx, userID, orderID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Order{}, ErrOrderNotFound
	}
	if err != nil {
		return domain.Order{}, err
	}
	if status != "PENDING" {
		return domain.Order{}, ErrOrderNotCancellable
	}

	items, err := s.repo.GetItemsForCancellation(ctx, tx, orderID)
	if err != nil {
		return domain.Order{}, err
	}
	if len(items) == 0 {
		return domain.Order{}, ErrOrderNotCancellable
	}

	for _, item := range items {
		onHandAfter, reservedAfter, err := s.repo.ReleaseReservedInventory(ctx, tx, item.ProductID, item.Quantity)
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Order{}, ErrOrderInventoryConflict
		}
		if err != nil {
			return domain.Order{}, err
		}
		if err := s.repo.RecordInventoryMovement(ctx, tx, domain.InventoryMovementInput{
			ProductID: item.ProductID, MovementType: "ORDER_RELEASED",
			OnHandDelta: 0, ReservedDelta: -item.Quantity,
			OnHandAfter: onHandAfter, ReservedAfter: reservedAfter,
			Reason:          "Reserved stock released after order cancellation",
			CreatedByUserID: userID, ReferenceType: "ORDER", ReferenceID: orderID,
		}); err != nil {
			return domain.Order{}, err
		}
	}

	if err := s.repo.MarkCancelled(ctx, tx, userID, orderID); err != nil {
		return domain.Order{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Order{}, err
	}

	return s.repo.GetForUser(ctx, userID, orderID)
}

// ExpireUnpaidOrders marks old, unpaid orders EXPIRED and releases their
// reservations atomically. Orders with a PENDING payment are deliberately left
// alone: a delayed provider callback could indicate a real charge and must be reconciled first.
func (s *OrderService) ExpireUnpaidOrders(ctx context.Context, olderThanMinutes, batchSize int) (int, error) {
	if olderThanMinutes < 1 || olderThanMinutes > 10080 || batchSize < 1 || batchSize > 1000 {
		return 0, ErrInvalidOrderQuery
	}
	return s.repo.ExpireUnpaidPendingOrders(ctx, olderThanMinutes, batchSize)
}
