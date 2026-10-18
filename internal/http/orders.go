// package http

// import (
// 	"net/http"
// 	"strings"

// 	"github.com/gin-gonic/gin"
// 	"github.com/jackc/pgx/v5"
// 	"github.com/jackc/pgx/v5/pgxpool"
// )

// type orderHandler struct {
// 	pool *pgxpool.Pool
// }

// type createOrderRequest struct {
// 	CustomerEmail string `json:"customer_email" binding:"required,email"`
// 	ProductID     string `json:"product_id" binding:"required,uuid"`
// 	Quantity      int    `json:"quantity" binding:"required,min=1,max=10000"`
// }

// func (h *orderHandler) createOrder(c *gin.Context) {
// 	var req createOrderRequest

// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
// 		return
// 	}

// 	req.CustomerEmail = strings.TrimSpace(req.CustomerEmail)

// 	ctx := c.Request.Context()
// 	tx, err := h.pool.BeginTx(ctx, pgx.TxOptions{})
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not start transaction"})
// 		return
// 	}
// 	defer tx.Rollback(ctx)

// 	var productName string
// 	var pricePaise int64

// 	err = tx.QueryRow(ctx, `
// 		SELECT name, price_paise
// 		FROM products
// 		WHERE id = $1 AND is_active = TRUE
// 	`, req.ProductID).Scan(&productName, &pricePaise)

// 	if err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
// 		return
// 	}

// 	var available bool

// 	err = tx.QueryRow(ctx, `
// 		UPDATE inventory
// 		SET reserved_quantity = reserved_quantity + $2,
// 		    updated_at = NOW()
// 		WHERE product_id = $1
// 		  AND on_hand_quantity - reserved_quantity >= $2
// 		RETURNING TRUE
// 	`, req.ProductID, req.Quantity).Scan(&available)

// 	if err != nil {
// 		if err == pgx.ErrNoRows {
// 			c.JSON(http.StatusConflict, gin.H{"error": "insufficient stock"})
// 			return
// 		}
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "inventory reservation failed"})
// 		return
// 	}

// 	lineTotalPaise := pricePaise * int64(req.Quantity)

// 	var orderID string

// 	err = tx.QueryRow(ctx, `
// 		INSERT INTO orders (customer_email, total_amount_paise)
// 		VALUES ($1, $2)
// 		RETURNING id
// 	`, req.CustomerEmail, lineTotalPaise).Scan(&orderID)

// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "order creation failed"})
// 		return
// 	}

// 	_, err = tx.Exec(ctx, `
// 		INSERT INTO order_items (
// 			order_id, product_id, product_name,
// 			quantity, unit_price_paise, line_total_paise
// 		)
// 		VALUES ($1, $2, $3, $4, $5, $6)
// 	`, orderID, req.ProductID, productName,
// 		req.Quantity, pricePaise, lineTotalPaise)

// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "order item creation failed"})
// 		return
// 	}

// 	if err := tx.Commit(ctx); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "order transaction failed"})
// 		return
// 	}

//		c.JSON(http.StatusCreated, gin.H{
//			"order_id":           orderID,
//			"status":             "PENDING",
//			"total_amount_paise": lineTotalPaise,
//		})
//	}
package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"github.com/manmohansharma21/orderflow/internal/domain"
	"github.com/manmohansharma21/orderflow/internal/service"
)

type orderHandler struct {
	service *service.OrderService
}

type createOrderRequest struct {
	CustomerEmail string `json:"customer_email" binding:"required,email"`
	ProductID     string `json:"product_id" binding:"required,uuid"`
	Quantity      int    `json:"quantity" binding:"required,min=1,max=10000"`
}

func newOrderHandler(orderService *service.OrderService) *orderHandler {
	return &orderHandler{
		service: orderService,
	}
}

func (h *orderHandler) createOrder(c *gin.Context) {
	var req createOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	result, err := h.service.CreateOrder(
		c.Request.Context(),
		domain.CreateOrderRequest{
			CustomerEmail: strings.TrimSpace(req.CustomerEmail),
			ProductID:     req.ProductID,
			Quantity:      req.Quantity,
		},
	)

	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidOrder):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid order",
			})

		case errors.Is(err, service.ErrProductNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "product not found",
			})

		case errors.Is(err, service.ErrInsufficientStock):
			c.JSON(http.StatusConflict, gin.H{
				"error": "insufficient stock",
			})

		case errors.Is(err, pgx.ErrNoRows):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "product not found",
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "order creation failed",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, result)
}
