package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/manmohansharma21/orderflow/internal/observability"
	"github.com/manmohansharma21/orderflow/internal/repository"
	"github.com/manmohansharma21/orderflow/internal/service"
)

type databasePinger interface{ Ping(context.Context) error }

func NewRouter(pool databasePinger, readinessTimeoutSeconds int) *gin.Engine {
	return newRouter(pool, readinessTimeoutSeconds, nil, time.Minute)
}

func NewRouterWithProductCache(pool databasePinger, readinessTimeoutSeconds int, productCache service.ProductCache, productCacheTTL time.Duration) *gin.Engine {
	return newRouter(pool, readinessTimeoutSeconds, productCache, productCacheTTL)
}

func newRouter(pool databasePinger, readinessTimeoutSeconds int, productCache service.ProductCache, productCacheTTL time.Duration) *gin.Engine {
	router := gin.New()
	metrics := observability.NewHTTPMetrics()
	router.Use(gin.Logger(), gin.Recovery(), metrics.Middleware())
	router.GET("/metrics", metrics.Handler())
	router.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	router.GET("/ready", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(readinessTimeoutSeconds)*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	if dbPool, ok := pool.(*pgxpool.Pool); ok {
		auth := newAuthHandler(dbPool)
		router.POST("/auth/register", auth.register)
		router.POST("/auth/login", auth.login)
		router.GET("/profile", AuthRequired(), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"user_id": c.GetString("user_id"), "role": c.GetString("user_role")})
		})
		router.POST("/admin/example", AuthRequired(), RequireRole("ADMIN"), func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "admin access granted"}) })
		orderService := service.NewOrderService(repository.NewOrderRepository(dbPool))
		orders := newOrderHandler(orderService)
		router.POST("/orders", AuthRequired(), orders.createOrder)
		router.GET("/orders", AuthRequired(), orders.listOrders)
		router.GET("/orders/:id", AuthRequired(), orders.getOrder)
		router.POST("/orders/:id/cancel", AuthRequired(), orders.cancelOrder)
		paymentService := service.NewPaymentService(repository.NewPaymentRepository(dbPool))
		payments := newPaymentHandler(paymentService)
		router.POST("/orders/:id/payments", AuthRequired(), payments.createForOrder)
		router.POST("/payments/mock-callback", payments.mockCallback)
		router.GET("/admin/payments/reconciliation", AuthRequired(), RequireRole("ADMIN"), payments.reconciliation)
		mockProvider := newMockProviderHandler(paymentService)
		router.POST("/admin/mock-provider/payments/:id/outcome", AuthRequired(), RequireRole("ADMIN"), mockProvider.setOutcome)
		router.POST("/admin/payments/reconciliation/run", AuthRequired(), RequireRole("ADMIN"), mockProvider.recoverKnownOutcomes)
		reportService := service.NewReportService(repository.NewReportRepository(dbPool))
		reports := newReportHandler(reportService)
		router.GET("/admin/reports/orders", AuthRequired(), RequireRole("ADMIN"), reports.orders)
		router.GET("/admin/reports/sales", AuthRequired(), RequireRole("ADMIN"), reports.sales)
		router.GET("/admin/reports/inventory", AuthRequired(), RequireRole("ADMIN"), reports.inventory)
		productService := service.NewProductServiceWithCache(repository.NewProductRepository(dbPool), productCache, productCacheTTL)
		products := newProductHandler(productService)
		router.GET("/products", products.list)
		router.GET("/products/:id", products.getByID)
		router.POST("/products", AuthRequired(), RequireRole("ADMIN"), products.create)
		router.PATCH("/products/:id", AuthRequired(), RequireRole("ADMIN"), products.update)
		router.DELETE("/products/:id", AuthRequired(), RequireRole("ADMIN"), products.deactivate)
		inventoryService := service.NewInventoryService(repository.NewInventoryRepository(dbPool))
		inventory := newInventoryHandler(inventoryService)
		router.GET("/inventory/:product_id", inventory.getAvailability)
		router.POST("/admin/inventory/receipts", AuthRequired(), RequireRole("ADMIN"), inventory.receiveStock)
		router.POST("/admin/inventory/adjustments", AuthRequired(), RequireRole("ADMIN"), inventory.adjustStock)
		router.GET("/admin/inventory/movements", AuthRequired(), RequireRole("ADMIN"), inventory.listMovements)
	}
	return router
}
