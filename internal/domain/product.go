package domain

type Product struct {
	ID          string  `json:"id"`
	SKU         string  `json:"sku"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	PricePaise  int64   `json:"price_paise"`
	IsActive    bool    `json:"is_active"`
}

type ProductList struct {
	Items []Product `json:"items"`
	Page  int       `json:"page"`
	Limit int       `json:"limit"`
	Total int64     `json:"total"`
}

type CreateProductInput struct {
	SKU             string
	Name            string
	Description     *string
	PricePaise      int64
	InitialStock    int
	CreatedByUserID string
}

type UpdateProductInput struct {
	SKU         *string
	Name        *string
	Description *string
	PricePaise  *int64
	IsActive    *bool
}
