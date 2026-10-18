package domain

import "time"

type InventorySummary struct {
	ProductID         string `json:"product_id"`
	SKU               string `json:"sku"`
	ProductName       string `json:"product_name"`
	OnHandQuantity    int    `json:"on_hand_quantity"`
	ReservedQuantity  int    `json:"reserved_quantity"`
	AvailableQuantity int    `json:"available_quantity"`
}

type InventoryMovement struct {
	ID              string    `json:"id"`
	ProductID       string    `json:"product_id"`
	SKU             string    `json:"sku"`
	ProductName     string    `json:"product_name"`
	MovementType    string    `json:"movement_type"`
	OnHandDelta     int       `json:"on_hand_delta"`
	ReservedDelta   int       `json:"reserved_delta"`
	OnHandAfter     int       `json:"on_hand_after"`
	ReservedAfter   int       `json:"reserved_after"`
	Reason          string    `json:"reason"`
	CreatedByUserID string    `json:"created_by_user_id,omitempty"`
	ReferenceType   string    `json:"reference_type,omitempty"`
	ReferenceID     string    `json:"reference_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type InventoryMovementList struct {
	Items []InventoryMovement `json:"items"`
	Page  int                 `json:"page"`
	Limit int                 `json:"limit"`
	Total int64               `json:"total"`
}

type InventoryMovementInput struct {
	ProductID       string
	MovementType    string
	OnHandDelta     int
	ReservedDelta   int
	OnHandAfter     int
	ReservedAfter   int
	Reason          string
	CreatedByUserID string
	ReferenceType   string
	ReferenceID     string
}
