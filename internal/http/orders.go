package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/manmohansharma21/orderflow/internal/domain"
	"github.com/manmohansharma21/orderflow/internal/service"
)

type orderHandler struct {
	service *service.OrderService
}

type createOrderRequest struct {
	ProductID string `json:"product_id" binding:"required,uuid"`
	Quantity  int    `json:"quantity" binding:"required,min=1,max=10000"`
}

func newOrderHandler(orderService *service.OrderService) *orderHandler {
	return &orderHandler{service: orderService}
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
			UserID:    c.GetString("user_id"),
			ProductID: req.ProductID,
			Quantity:  req.Quantity,
		},
	)

	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidOrder):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid order",
			})

		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "authenticated user not found; please log in again",
			})

		case errors.Is(err, service.ErrProductNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "product not found",
			})

		case errors.Is(err, service.ErrInsufficientStock):
			c.JSON(http.StatusConflict, gin.H{
				"error": "insufficient stock",
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
