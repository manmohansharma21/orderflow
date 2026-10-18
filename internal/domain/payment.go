package domain

import "time"

type Payment struct {
	ID                string    `json:"id"`
	OrderID           string    `json:"order_id"`
	Provider          string    `json:"provider"`
	ProviderPaymentID string    `json:"provider_payment_id,omitempty"`
	Status            string    `json:"status"`
	AmountPaise       int64     `json:"amount_paise"`
	Currency          string    `json:"currency"`
	FailureReason     string    `json:"failure_reason,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type PaymentCallbackDetails struct {
	AmountPaise   int64
	Currency      string
	FailureReason string
}

type PaymentReconciliationIssue struct {
	IssueType     string    `json:"issue_type"`
	PaymentID     string    `json:"payment_id,omitempty"`
	OrderID       string    `json:"order_id"`
	PaymentStatus string    `json:"payment_status,omitempty"`
	OrderStatus   string    `json:"order_status"`
	AmountPaise   int64     `json:"amount_paise"`
	CreatedAt     time.Time `json:"created_at"`
	Details       string    `json:"details"`
}

type PaymentReconciliationReport struct {
	GeneratedAt      time.Time                    `json:"generated_at"`
	OlderThanMinutes int                          `json:"older_than_minutes"`
	IssueCount       int                          `json:"issue_count"`
	Issues           []PaymentReconciliationIssue `json:"issues"`
}
