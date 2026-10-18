package http

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/manmohansharma21/orderflow/internal/service"
)

type mockProviderHandler struct{ service *service.PaymentService }

type mockProviderOutcomeRequest struct {
	Outcome       string `json:"outcome" binding:"required,oneof=SUCCESS FAILED"`
	AmountPaise   *int64 `json:"amount_paise" binding:"required,gte=0"`
	Currency      string `json:"currency" binding:"required,len=3"`
	FailureReason string `json:"failure_reason"`
}

func newMockProviderHandler(s *service.PaymentService) *mockProviderHandler {
	return &mockProviderHandler{service: s}
}

// setOutcome is a local teaching/demo hook. It deliberately refuses to run
// when APP_ENV=production and requires ADMIN authorization at the router.
func (h *mockProviderHandler) setOutcome(c *gin.Context) {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production") {
		c.JSON(http.StatusNotFound, gin.H{"error": "mock provider simulator is disabled in production"})
		return
	}
	paymentID := c.Param("id")
	if !validUUID(paymentID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment id"})
		return
	}
	var req mockProviderOutcomeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "outcome, amount_paise and three-letter currency are required"})
		return
	}
	result, err := h.service.SetMockProviderOutcome(
		c.Request.Context(), paymentID, req.Outcome, *req.AmountPaise, req.Currency, req.FailureReason,
	)
	if err != nil {
		switch {
		case err == service.ErrInvalidPaymentRequest:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid simulated provider outcome"})
		case err == service.ErrPaymentNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "mock provider payment not found"})
		case err == service.ErrMockProviderOutcomeFinalized:
			c.JSON(http.StatusConflict, gin.H{"error": "provider outcome is already final"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update mock provider state"})
		}
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *mockProviderHandler) recoverKnownOutcomes(c *gin.Context) {
	age, err := strconv.Atoi(c.DefaultQuery("older_than_minutes", "0"))
	if err != nil || age < 0 || age > 10080 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "older_than_minutes must be an integer from 0 to 10080"})
		return
	}
	batch, err := strconv.Atoi(c.DefaultQuery("limit", "100"))
	if err != nil || batch < 1 || batch > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be an integer from 1 to 1000"})
		return
	}
	count, err := h.service.ReconcileKnownMockProviderOutcomes(c.Request.Context(), age, batch)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "mock-provider reconciliation failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"recovered_count": count, "older_than_minutes": age, "limit": batch})
}
