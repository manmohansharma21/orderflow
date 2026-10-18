package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/manmohansharma21/orderflow/internal/repository"
	"github.com/manmohansharma21/orderflow/internal/service"
)

type databasePinger interface {
	Ping(context.Context) error
}

func NewRouter(pool databasePinger, readinessTimeoutSeconds int) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/ready", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(
			c.Request.Context(),
			time.Duration(readinessTimeoutSeconds)*time.Second,
		)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not_ready",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	if dbPool, ok := pool.(*pgxpool.Pool); ok {
		orderRepo := repository.NewOrderRepository(dbPool)
		orderService := service.NewOrderService(orderRepo)
		router.POST("/orders", newOrderHandler(orderService).createOrder)

		productRepo := repository.NewProductRepository(dbPool)
		productService := service.NewProductService(productRepo)
		products := newProductHandler(productService)

		router.GET("/products", products.list)
		router.GET("/products/:id", products.getByID)
	}

	return router
}
