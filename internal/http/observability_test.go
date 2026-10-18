package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestIDAndMetricsEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(fakePinger{}, 3)

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	request.Header.Set("X-Request-ID", "orderflow-test-request-123")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected /health 200, got %d", response.Code)
	}
	if got := response.Header().Get("X-Request-ID"); got != "orderflow-test-request-123" {
		t.Fatalf("expected request id echoed, got %q", got)
	}

	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsResp := httptest.NewRecorder()
	router.ServeHTTP(metricsResp, metricsReq)
	if metricsResp.Code != http.StatusOK {
		t.Fatalf("expected /metrics 200, got %d", metricsResp.Code)
	}
	body := metricsResp.Body.String()
	if !strings.Contains(body, "orderflow_http_requests_total") || !strings.Contains(body, `route="/health"`) || !strings.Contains(body, "orderflow_http_request_duration_seconds_bucket") {
		t.Fatalf("expected request counter and duration histogram in metrics output, got %q", body)
	}
}
