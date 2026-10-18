package domain

import "time"

type CreateOrderRequest struct {
	UserID    string
	ProductID string
	Quantity  int
}

type CreateOrderResult struct {
	OrderID          string `json:"order_id"`
	Status           string `json:"status"`
	TotalAmountPaise int64  `json:"total_amount_paise"`
}

type OrderItem struct {
	ProductID      string `json:"product_id"`
	ProductName    string `json:"product_name"`
	Quantity       int    `json:"quantity"`
	UnitPricePaise int64  `json:"unit_price_paise"`
	LineTotalPaise int64  `json:"line_total_paise"`
}

type Order struct {
	ID               string      `json:"id"`
	Status           string      `json:"status"`
	TotalAmountPaise int64       `json:"total_amount_paise"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
	Items            []OrderItem `json:"items"`
}

type OrderList struct {
	Items []Order `json:"items"`
	Page  int     `json:"page"`
	Limit int     `json:"limit"`
	Total int64   `json:"total"`
}
