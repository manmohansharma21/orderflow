package http

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/manmohansharma21/orderflow/internal/service"
)

const testWebhookSecret = "orderflow-test-webhook-secret-0123456789abcdef"

func paymentTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := newPaymentHandler(service.NewPaymentService(nil))
	router := gin.New()
	// Supply an authenticated identity so tests of the Idempotency-Key header
	// reach the intended service validation rather than failing on missing user ID.
	router.POST("/orders/:id/payments", func(c *gin.Context) {
		c.Set("user_id", "test-user-id")
		c.Next()
	}, handler.createForOrder)
	router.POST("/payments/mock-callback", handler.mockCallback)
	router.GET("/admin/payments/reconciliation", handler.reconciliation)
	return router
}

func signTestPayload(secret, body string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(body))
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestMockCallbackRejectsMissingSecret(t *testing.T) {
	t.Setenv("MOCK_PAYMENT_WEBHOOK_SECRET", "")
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/payments/mock-callback", strings.NewReader(`{}`))
	paymentTestRouter().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected %d, got %d: %s", http.StatusServiceUnavailable, recorder.Code, recorder.Body.String())
	}
}

func TestMockCallbackRejectsInvalidSignature(t *testing.T) {
	t.Setenv("MOCK_PAYMENT_WEBHOOK_SECRET", testWebhookSecret)
	body := `{"payment_id":"28e3f2b2-cfc8-4df9-8b6c-d0f7104c9b1d","outcome":"SUCCESS","event_id":"evt-test"}`
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/payments/mock-callback", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Mock-Signature", "sha256=invalid")
	paymentTestRouter().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d: %s", http.StatusUnauthorized, recorder.Code, recorder.Body.String())
	}
}

func TestMockCallbackRejectsOversizedPayload(t *testing.T) {
	t.Setenv("MOCK_PAYMENT_WEBHOOK_SECRET", testWebhookSecret)
	body := strings.Repeat("x", 64*1024+1)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/payments/mock-callback", strings.NewReader(body))
	paymentTestRouter().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected %d, got %d: %s", http.StatusRequestEntityTooLarge, recorder.Code, recorder.Body.String())
	}
}

func TestMockCallbackRejectsMalformedSignedPayload(t *testing.T) {
	t.Setenv("MOCK_PAYMENT_WEBHOOK_SECRET", testWebhookSecret)
	body := `{`
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/payments/mock-callback", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Mock-Signature", signTestPayload(testWebhookSecret, body))
	paymentTestRouter().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d: %s", http.StatusBadRequest, recorder.Code, recorder.Body.String())
	}
}

func TestMockCallbackRejectsInvalidPaymentIDBeforeRepository(t *testing.T) {
	t.Setenv("MOCK_PAYMENT_WEBHOOK_SECRET", testWebhookSecret)
	body := `{"payment_id":"not-a-uuid","outcome":"SUCCESS","event_id":"evt-test"}`
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/payments/mock-callback", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Mock-Signature", signTestPayload(testWebhookSecret, body))
	paymentTestRouter().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d: %s", http.StatusBadRequest, recorder.Code, recorder.Body.String())
	}
}

func TestMockCallbackRejectsUnsupportedOutcomeBeforeRepository(t *testing.T) {
	t.Setenv("MOCK_PAYMENT_WEBHOOK_SECRET", testWebhookSecret)
	body := `{"payment_id":"28e3f2b2-cfc8-4df9-8b6c-d0f7104c9b1d","outcome":"REFUNDED","event_id":"evt-test","amount_paise":49900,"currency":"INR"}`
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/payments/mock-callback", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Mock-Signature", signTestPayload(testWebhookSecret, body))
	paymentTestRouter().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d: %s", http.StatusBadRequest, recorder.Code, recorder.Body.String())
	}
}

func TestMockCallbackRequiresAmountAndCurrency(t *testing.T) {
	t.Setenv("MOCK_PAYMENT_WEBHOOK_SECRET", testWebhookSecret)
	body := `{"payment_id":"28e3f2b2-cfc8-4df9-8b6c-d0f7104c9b1d","outcome":"SUCCESS","event_id":"evt-test"}`
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/payments/mock-callback", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Mock-Signature", signTestPayload(testWebhookSecret, body))
	paymentTestRouter().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected %d for missing amount/currency, got %d: %s", http.StatusBadRequest, recorder.Code, recorder.Body.String())
	}
}

func TestCreatePaymentRejectsInvalidOrderID(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/orders/not-a-uuid/payments", nil)
	paymentTestRouter().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d: %s", http.StatusBadRequest, recorder.Code, recorder.Body.String())
	}
}

func TestCreatePaymentRejectsMissingOrShortIdempotencyKeyBeforeRepository(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{name: "missing key", key: ""},
		{name: "short key", key: "short"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/orders/8e840ec0-9315-4574-847f-15e8a17a578a/payments", nil)
			if tt.key != "" {
				request.Header.Set("Idempotency-Key", tt.key)
			}
			paymentTestRouter().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected %d, got %d: %s", http.StatusBadRequest, recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestReconciliationRejectsInvalidThreshold(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{name: "not integer", path: "/admin/payments/reconciliation?older_than_minutes=abc"},
		{name: "below minimum", path: "/admin/payments/reconciliation?older_than_minutes=0"},
		{name: "above maximum", path: "/admin/payments/reconciliation?older_than_minutes=10081"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			paymentTestRouter().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected %d, got %d: %s", http.StatusBadRequest, recorder.Code, recorder.Body.String())
			}
		})
	}
}
