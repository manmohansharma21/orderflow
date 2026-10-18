package http

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"github.com/manmohansharma21/orderflow/internal/domain"
	"github.com/manmohansharma21/orderflow/internal/service"
)

type paymentHandler struct {
	service *service.PaymentService
}

type mockPaymentCallbackRequest struct {
	PaymentID     string `json:"payment_id"`
	Outcome       string `json:"outcome"`
	EventID       string `json:"event_id"`
	AmountPaise   *int64 `json:"amount_paise"`
	Currency      string `json:"currency"`
	FailureReason string `json:"failure_reason"`
}

func newPaymentHandler(s *service.PaymentService) *paymentHandler {
	return &paymentHandler{service: s}
}

func (h *paymentHandler) createForOrder(c *gin.Context) {
	orderID := c.Param("id")
	if !validUUID(orderID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}
	idempotencyKey := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	p, err := h.service.CreateForOrder(
		c.Request.Context(),
		c.GetString("user_id"),
		orderID,
		idempotencyKey,
	)
	if err != nil {
		writePaymentError(c, err)
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *paymentHandler) mockCallback(c *gin.Context) {
	secret := strings.TrimSpace(os.Getenv("MOCK_PAYMENT_WEBHOOK_SECRET"))
	if len(secret) < 32 {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "mock payment webhook secret is not configured"})
		return
	}

	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 64*1024+1))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not read callback"})
		return
	}
	if len(body) > 64*1024 {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "callback payload too large"})
		return
	}

	signature := strings.TrimSpace(c.GetHeader("X-Mock-Signature"))
	signature = strings.TrimPrefix(signature, "sha256=")
	provided, err := hex.DecodeString(signature)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid callback signature"})
		return
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	if !hmac.Equal(provided, mac.Sum(nil)) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid callback signature"})
		return
	}

	var req mockPaymentCallbackRequest
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid callback payload"})
		return
	}
	if !validUUID(req.PaymentID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment id"})
		return
	}
	if req.AmountPaise == nil || strings.TrimSpace(req.Currency) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount_paise and currency are required"})
		return
	}
	details := domain.PaymentCallbackDetails{AmountPaise: *req.AmountPaise, Currency: req.Currency, FailureReason: req.FailureReason}
	p, err := h.service.ProcessMockCallback(
		c.Request.Context(), req.PaymentID, req.Outcome, req.EventID, details,
	)
	if err != nil {
		writePaymentError(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *paymentHandler) reconciliation(c *gin.Context) {
	minutes, err := strconv.Atoi(c.DefaultQuery("older_than_minutes", "15"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "older_than_minutes must be an integer"})
		return
	}
	report, err := h.service.ReconciliationReport(c.Request.Context(), minutes)
	if err != nil {
		writePaymentError(c, err)
		return
	}
	c.JSON(http.StatusOK, report)
}

func writePaymentError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidPaymentRequest),
		errors.Is(err, service.ErrInvalidPaymentOutcome):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment request"})
	case errors.Is(err, service.ErrPaymentOrderNotFound),
		errors.Is(err, service.ErrPaymentNotFound),
		errors.Is(err, pgx.ErrNoRows):
		c.JSON(http.StatusNotFound, gin.H{"error": "order or payment not found"})
	case errors.Is(err, service.ErrPaymentConflict):
		c.JSON(http.StatusConflict, gin.H{"error": "payment or order state conflict"})
	case errors.Is(err, service.ErrPaymentAmountMismatch):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "callback amount or currency does not match the persisted payment"})
	case errors.Is(err, service.ErrIdempotencyConflict):
		c.JSON(http.StatusConflict, gin.H{"error": "idempotency key already used for another payment"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "payment operation failed"})
	}
}
