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
	ErrProductNotFound        = errors.New("product not found")
	ErrUserNotFound           = errors.New("user not found")
	ErrInsufficientStock      = errors.New("insufficient stock")
	ErrInvalidOrder           = errors.New("invalid order")
	ErrOrderNotFound          = errors.New("order not found")
	ErrOrderNotCancellable    = errors.New("order is not cancellable")
	ErrOrderInventoryConflict = errors.New("reserved inventory cannot be released")
	ErrInvalidOrderQuery      = errors.New("invalid order query")
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
	req.ProductID = strings.TrimSpace(req.ProductID)
	if req.UserID == "" || req.ProductID == "" || req.Quantity < 1 || req.Quantity > 10000 {
		return domain.CreateOrderResult{}, ErrInvalidOrder
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return domain.CreateOrderResult{}, err
	}
	defer tx.Rollback(ctx)

	productName, pricePaise, err := s.repo.GetActiveProduct(ctx, tx, req.ProductID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.CreateOrderResult{}, ErrProductNotFound
		}
		return domain.CreateOrderResult{}, err
	}
	if pricePaise > 0 && int64(req.Quantity) > math.MaxInt64/pricePaise {
		return domain.CreateOrderResult{}, ErrInvalidOrder
	}
	totalPaise := pricePaise * int64(req.Quantity)

	reserved, onHandAfter, reservedAfter, err := s.repo.ReserveInventory(ctx, tx, req.ProductID, req.Quantity)
	if err != nil {
		return domain.CreateOrderResult{}, err
	}
	if !reserved {
		return domain.CreateOrderResult{}, ErrInsufficientStock
	}

	orderID, err := s.repo.CreateOrder(ctx, tx, req.UserID, totalPaise)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.CreateOrderResult{}, ErrUserNotFound
		}
		return domain.CreateOrderResult{}, err
	}

	if err := s.repo.RecordInventoryMovement(ctx, tx, domain.InventoryMovementInput{
		ProductID: req.ProductID, MovementType: "ORDER_RESERVED",
		OnHandDelta: 0, ReservedDelta: req.Quantity,
		OnHandAfter: onHandAfter, ReservedAfter: reservedAfter,
		Reason: "Stock reserved for order creation", CreatedByUserID: req.UserID,
		ReferenceType: "ORDER", ReferenceID: orderID,
	}); err != nil {
		return domain.CreateOrderResult{}, err
	}

	if err := s.repo.CreateOrderItem(ctx, tx, orderID, req.ProductID, productName, req.Quantity, pricePaise, totalPaise); err != nil {
		return domain.CreateOrderResult{}, err
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
