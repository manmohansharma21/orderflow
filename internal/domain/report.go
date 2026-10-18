package domain

import "time"

type OrdersOperationalReport struct {
	GeneratedAt           time.Time `json:"generated_at"`
	TotalOrders           int64     `json:"total_orders"`
	PendingOrders         int64     `json:"pending_orders"`
	ConfirmedOrders       int64     `json:"confirmed_orders"`
	FailedOrders          int64     `json:"failed_orders"`
	CancelledOrders       int64     `json:"cancelled_orders"`
	ExpiredOrders         int64     `json:"expired_orders"`
	ConfirmedRevenuePaise int64     `json:"confirmed_revenue_paise"`
}

type SalesReport struct {
	GeneratedAt            time.Time `json:"generated_at"`
	FromDate               string    `json:"from_date"`
	ToDate                 string    `json:"to_date"`
	ConfirmedOrders        int64     `json:"confirmed_orders"`
	ConfirmedRevenuePaise  int64     `json:"confirmed_revenue_paise"`
	SuccessfulPayments     int64     `json:"successful_payments"`
	SuccessfulPaymentPaise int64     `json:"successful_payment_paise"`
	FailedPayments         int64     `json:"failed_payments"`
	FailedPaymentPaise     int64     `json:"failed_payment_paise"`
}

type InventoryReportItem struct {
	ProductID         string `json:"product_id"`
	SKU               string `json:"sku"`
	Name              string `json:"name"`
	OnHandQuantity    int64  `json:"on_hand_quantity"`
	ReservedQuantity  int64  `json:"reserved_quantity"`
	AvailableQuantity int64  `json:"available_quantity"`
}

type InventoryReport struct {
	GeneratedAt       time.Time             `json:"generated_at"`
	ActiveProducts    int64                 `json:"active_products"`
	TotalOnHand       int64                 `json:"total_on_hand"`
	TotalReserved     int64                 `json:"total_reserved"`
	TotalAvailable    int64                 `json:"total_available"`
	LowStockProducts  int64                 `json:"low_stock_products"`
	LowStockThreshold int64                 `json:"low_stock_threshold"`
	LowStockItems     []InventoryReportItem `json:"low_stock_items"`
}
