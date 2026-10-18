package domain

import "time"

// MockProviderOutcome models the independent simulated provider-side result.
// It is intentionally separate from Payment.Status so reconciliation can recover
// a terminal provider result when a callback was not delivered to OrderFlow.
type MockProviderOutcome struct {
	PaymentID         string    `json:"payment_id"`
	ProviderPaymentID string    `json:"provider_payment_id"`
	Outcome           string    `json:"outcome"`
	AmountPaise       int64     `json:"amount_paise"`
	Currency          string    `json:"currency"`
	FailureReason     string    `json:"failure_reason,omitempty"`
	UpdatedAt         time.Time `json:"updated_at"`
}
