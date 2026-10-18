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
	ErrProductNotFound   = errors.New("product not found")
	ErrUserNotFound      = errors.New("user not found")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrInvalidOrder      = errors.New("invalid order")
)

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

	if req.UserID == "" ||
		req.ProductID == "" ||
		req.Quantity < 1 ||
		req.Quantity > 10000 {
		return domain.CreateOrderResult{}, ErrInvalidOrder
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return domain.CreateOrderResult{}, err
	}
	defer tx.Rollback(ctx)

	productName, pricePaise, err := s.repo.GetActiveProduct(
		ctx,
		tx,
		req.ProductID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.CreateOrderResult{}, ErrProductNotFound
		}
		return domain.CreateOrderResult{}, err
	}

	if pricePaise > 0 &&
		int64(req.Quantity) > math.MaxInt64/pricePaise {
		return domain.CreateOrderResult{}, ErrInvalidOrder
	}

	totalPaise := pricePaise * int64(req.Quantity)

	reserved, err := s.repo.ReserveInventory(
		ctx,
		tx,
		req.ProductID,
		req.Quantity,
	)
	if err != nil {
		return domain.CreateOrderResult{}, err
	}
	if !reserved {
		return domain.CreateOrderResult{}, ErrInsufficientStock
	}

	orderID, err := s.repo.CreateOrder(
		ctx,
		tx,
		req.UserID,
		totalPaise,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.CreateOrderResult{}, ErrUserNotFound
		}
		return domain.CreateOrderResult{}, err
	}

	if err := s.repo.CreateOrderItem(
		ctx,
		tx,
		orderID,
		req.ProductID,
		productName,
		req.Quantity,
		pricePaise,
		totalPaise,
	); err != nil {
		return domain.CreateOrderResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.CreateOrderResult{}, err
	}

	return domain.CreateOrderResult{
		OrderID:          orderID,
		Status:           "PENDING",
		TotalAmountPaise: totalPaise,
	}, nil
}
