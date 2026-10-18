package domain

type CreateOrderRequest struct {
	CustomerEmail string
	ProductID     string
	Quantity      int
}

type CreateOrderResult struct {
	OrderID          string `json:"order_id"`
	Status           string `json:"status"`
	TotalAmountPaise int64  `json:"total_amount_paise"`
}
