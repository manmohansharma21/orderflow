package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const testJWTSecret = "orderflow-test-jwt-secret-that-is-long-enough-2026"

func signedTestJWT(t *testing.T, method jwt.SigningMethod, role string) string {
	t.Helper()
	claims := authClaims{
		UserID: "test-user-id",
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "test-user-id",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(method, claims)
	value, err := token.SignedString([]byte(testJWTSecret))
	if err != nil {
		t.Fatalf("sign test JWT: %v", err)
	}
	return value
}

func middlewareTestRouter(middleware ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/protected", append(middleware, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"user_id": c.GetString("user_id"),
			"role":    c.GetString("user_role"),
		})
	})...)
	return router
}

func TestAuthRequiredRejectsMissingOrMalformedBearer(t *testing.T) {
	t.Setenv("JWT_SECRET", testJWTSecret)
	tests := []struct {
		name   string
		header string
	}{
		{name: "missing header", header: ""},
		{name: "wrong scheme", header: "Token abc"},
		{name: "missing token", header: "Bearer   "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.header != "" {
				request.Header.Set("Authorization", tt.header)
			}
			middlewareTestRouter(AuthRequired()).ServeHTTP(recorder, request)
			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("expected %d, got %d: %s", http.StatusUnauthorized, recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestAuthRequiredRejectsWeakJWTSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "too-short")
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer arbitrary-token")
	middlewareTestRouter(AuthRequired()).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected %d, got %d: %s", http.StatusInternalServerError, recorder.Code, recorder.Body.String())
	}
}

func TestAuthRequiredRejectsInvalidOrWrongAlgorithmToken(t *testing.T) {
	t.Setenv("JWT_SECRET", testJWTSecret)
	tests := []struct {
		name  string
		token string
	}{
		{name: "malformed token", token: "not-a-jwt"},
		{name: "wrong algorithm", token: signedTestJWT(t, jwt.SigningMethodHS384, "CUSTOMER")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/protected", nil)
			request.Header.Set("Authorization", "Bearer "+tt.token)
			middlewareTestRouter(AuthRequired()).ServeHTTP(recorder, request)
			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("expected %d, got %d: %s", http.StatusUnauthorized, recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestAuthRequiredAcceptsValidJWTAndSetsClaims(t *testing.T) {
	t.Setenv("JWT_SECRET", testJWTSecret)
	token := signedTestJWT(t, jwt.SigningMethodHS256, "CUSTOMER")
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	middlewareTestRouter(AuthRequired()).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `"user_id":"test-user-id"`) || !strings.Contains(body, `"role":"CUSTOMER"`) {
		t.Fatalf("expected user claims in response, got %s", body)
	}
}

func TestRequireRoleAllowsOnlyMatchingRole(t *testing.T) {
	tests := []struct {
		name       string
		role       string
		wantStatus int
	}{
		{name: "admin allowed", role: "ADMIN", wantStatus: http.StatusOK},
		{name: "customer denied", role: "CUSTOMER", wantStatus: http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.GET("/admin", func(c *gin.Context) {
				c.Set("user_role", tt.role)
				c.Next()
			}, RequireRole("ADMIN"), func(c *gin.Context) {
				c.Status(http.StatusOK)
			})
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/admin", nil))
			if recorder.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d: %s", tt.wantStatus, recorder.Code, recorder.Body.String())
			}
		})
	}
}
