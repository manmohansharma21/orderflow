package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/manmohansharma21/orderflow/internal/domain"
)

func TestOrderServiceRejectsInvalidCreateRequests(t *testing.T) {
	svc := NewOrderService(nil)
	valid := domain.CreateOrderRequest{UserID: "user-1", ProductID: "product-1", Quantity: 1}
	tests := []struct {
		name string
		req  domain.CreateOrderRequest
	}{
		{name: "missing user", req: domain.CreateOrderRequest{ProductID: valid.ProductID, Quantity: 1}},
		{name: "missing product", req: domain.CreateOrderRequest{UserID: valid.UserID, Quantity: 1}},
		{name: "zero quantity", req: domain.CreateOrderRequest{UserID: valid.UserID, ProductID: valid.ProductID, Quantity: 0}},
		{name: "quantity above limit", req: domain.CreateOrderRequest{UserID: valid.UserID, ProductID: valid.ProductID, Quantity: 10001}},
		{name: "blank whitespace user", req: domain.CreateOrderRequest{UserID: "   ", ProductID: valid.ProductID, Quantity: 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateOrder(context.Background(), tt.req)
			if !errors.Is(err, ErrInvalidOrder) {
				t.Fatalf("expected ErrInvalidOrder, got %v", err)
			}
		})
	}
}

func TestOrderServiceRejectsInvalidQueries(t *testing.T) {
	svc := NewOrderService(nil)
	if _, err := svc.ListForUser(context.Background(), " ", 1, 20); !errors.Is(err, ErrInvalidOrderQuery) {
		t.Fatalf("expected ErrInvalidOrderQuery for blank user, got %v", err)
	}
	if _, err := svc.ListForUser(context.Background(), "user-1", 1_000_001, 20); !errors.Is(err, ErrInvalidOrderQuery) {
		t.Fatalf("expected ErrInvalidOrderQuery for excessive page, got %v", err)
	}
	if _, err := svc.GetForUser(context.Background(), "", "order-1"); !errors.Is(err, ErrInvalidOrderQuery) {
		t.Fatalf("expected ErrInvalidOrderQuery for blank user, got %v", err)
	}
	if _, err := svc.CancelForUser(context.Background(), "user-1", " "); !errors.Is(err, ErrInvalidOrderQuery) {
		t.Fatalf("expected ErrInvalidOrderQuery for blank order, got %v", err)
	}
}

func TestProductServiceRejectsInvalidCreateRequests(t *testing.T) {
	svc := NewProductService(nil)
	valid := domain.CreateProductInput{
		SKU: "SKU-1", Name: "Name", PricePaise: 100, InitialStock: 1, CreatedByUserID: "user-1",
	}
	tests := []struct {
		name  string
		input domain.CreateProductInput
	}{
		{name: "missing SKU", input: domain.CreateProductInput{Name: valid.Name, PricePaise: valid.PricePaise, InitialStock: valid.InitialStock, CreatedByUserID: valid.CreatedByUserID}},
		{name: "missing name", input: domain.CreateProductInput{SKU: valid.SKU, PricePaise: valid.PricePaise, InitialStock: valid.InitialStock, CreatedByUserID: valid.CreatedByUserID}},
		{name: "negative price", input: func() domain.CreateProductInput { x := valid; x.PricePaise = -1; return x }()},
		{name: "negative stock", input: func() domain.CreateProductInput { x := valid; x.InitialStock = -1; return x }()},
		{name: "stock too large", input: func() domain.CreateProductInput { x := valid; x.InitialStock = 1_000_001; return x }()},
		{name: "missing creator", input: func() domain.CreateProductInput { x := valid; x.CreatedByUserID = " "; return x }()},
		{name: "SKU too long", input: func() domain.CreateProductInput { x := valid; x.SKU = strings.Repeat("S", 101); return x }()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), tt.input)
			if !errors.Is(err, ErrInvalidProduct) {
				t.Fatalf("expected ErrInvalidProduct, got %v", err)
			}
		})
	}
}

func TestProductServiceRejectsInvalidUpdatesAndDeactivateIDs(t *testing.T) {
	svc := NewProductService(nil)
	if _, err := svc.Update(context.Background(), "product-1", domain.UpdateProductInput{}); !errors.Is(err, ErrInvalidProduct) {
		t.Fatalf("expected ErrInvalidProduct for empty update, got %v", err)
	}
	badPrice := int64(-1)
	if _, err := svc.Update(context.Background(), "product-1", domain.UpdateProductInput{PricePaise: &badPrice}); !errors.Is(err, ErrInvalidProduct) {
		t.Fatalf("expected ErrInvalidProduct for negative price, got %v", err)
	}
	badSKU := "  "
	if _, err := svc.Update(context.Background(), "product-1", domain.UpdateProductInput{SKU: &badSKU}); !errors.Is(err, ErrInvalidProduct) {
		t.Fatalf("expected ErrInvalidProduct for blank SKU, got %v", err)
	}
	if _, err := svc.Deactivate(context.Background(), " "); !errors.Is(err, ErrInvalidProduct) {
		t.Fatalf("expected ErrInvalidProduct for blank ID, got %v", err)
	}
}

func TestInventoryServiceRejectsInvalidStockChanges(t *testing.T) {
	svc := NewInventoryService(nil)
	valid := domain.InventoryMovementInput{ProductID: "product-1", CreatedByUserID: "user-1", Reason: "audit note"}
	tests := []struct {
		name    string
		receipt bool
		input   domain.InventoryMovementInput
	}{
		{name: "zero receipt", receipt: true, input: func() domain.InventoryMovementInput { x := valid; x.OnHandDelta = 0; return x }()},
		{name: "oversized receipt", receipt: true, input: func() domain.InventoryMovementInput { x := valid; x.OnHandDelta = 1_000_001; return x }()},
		{name: "missing product", receipt: true, input: func() domain.InventoryMovementInput { x := valid; x.OnHandDelta = 1; x.ProductID = " "; return x }()},
		{name: "missing actor", receipt: true, input: func() domain.InventoryMovementInput { x := valid; x.OnHandDelta = 1; x.CreatedByUserID = " "; return x }()},
		{name: "short reason", receipt: true, input: func() domain.InventoryMovementInput { x := valid; x.OnHandDelta = 1; x.Reason = "ok"; return x }()},
		{name: "zero adjustment", input: func() domain.InventoryMovementInput { x := valid; x.OnHandDelta = 0; return x }()},
		{name: "oversized adjustment", input: func() domain.InventoryMovementInput { x := valid; x.OnHandDelta = 1_000_001; return x }()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.receipt {
				_, err = svc.ReceiveStock(context.Background(), tt.input)
			} else {
				_, err = svc.AdjustStock(context.Background(), tt.input)
			}
			if !errors.Is(err, ErrInvalidInventoryRequest) {
				t.Fatalf("expected ErrInvalidInventoryRequest, got %v", err)
			}
		})
	}
}

func TestInventoryServiceRejectsInvalidMovementFilters(t *testing.T) {
	svc := NewInventoryService(nil)
	if _, err := svc.ListMovements(context.Background(), "", "UNKNOWN_TYPE", 1, 20); !errors.Is(err, ErrInvalidInventoryRequest) {
		t.Fatalf("expected ErrInvalidInventoryRequest for unsupported type, got %v", err)
	}
	if _, err := svc.ListMovements(context.Background(), "", "", 1_000_001, 20); !errors.Is(err, ErrInvalidInventoryRequest) {
		t.Fatalf("expected ErrInvalidInventoryRequest for excessive page, got %v", err)
	}
}
