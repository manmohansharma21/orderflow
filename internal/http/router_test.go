package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakePinger struct {
	err error
}

func (f fakePinger) Ping(context.Context) error {
	return f.err
}

func TestHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(fakePinger{}, 3)

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d",
			http.StatusOK, response.Code)
	}

	if response.Body.String() != "{\"status\":\"ok\"}" {
		t.Fatalf("unexpected response body: %s", response.Body.String())
	}
}

func TestReadyEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		pingErr        error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "database reachable",
			pingErr:        nil,
			expectedStatus: http.StatusOK,
			expectedBody:   "{\"status\":\"ready\"}",
		},
		{
			name:           "database unavailable",
			pingErr:        errors.New("database connection failed"),
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   "{\"status\":\"not_ready\"}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter(fakePinger{err: tt.pingErr}, 3)

			request := httptest.NewRequest(
				http.MethodGet,
				"/ready",
				nil,
			)
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d",
					tt.expectedStatus, response.Code)
			}

			if response.Body.String() != tt.expectedBody {
				t.Fatalf("expected body %s, got %s",
					tt.expectedBody, response.Body.String())
			}
		})
	}
}
