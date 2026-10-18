package http

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/manmohansharma21/orderflow/internal/domain"
	"github.com/manmohansharma21/orderflow/internal/service"
)

type orderHandler struct{ service *service.OrderService }

type createOrderLineRequest struct {
	ProductID string `json:"product_id" binding:"required,uuid"`
	Quantity  int    `json:"quantity" binding:"required,min=1,max=10000"`
}

type createOrderRequest struct {
	Items     []createOrderLineRequest `json:"items" binding:"omitempty,min=1,max=50,dive"`
	ProductID string                   `json:"product_id"`
	Quantity  int                      `json:"quantity"`
}

func newOrderHandler(orderService *service.OrderService) *orderHandler {
	return &orderHandler{service: orderService}
}

func (h *orderHandler) createOrder(c *gin.Context) {
	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order request"})
		return
	}
	hasItems := len(req.Items) > 0
	hasLegacy := strings.TrimSpace(req.ProductID) != "" || req.Quantity != 0
	if hasItems == hasLegacy {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provide either items[] or the legacy product_id/quantity fields"})
		return
	}
	key := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	if len(key) < 8 || len(key) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Idempotency-Key header must contain 8 to 200 characters"})
		return
	}
	items := make([]domain.CreateOrderLineRequest, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, domain.CreateOrderLineRequest{ProductID: item.ProductID, Quantity: item.Quantity})
	}
	result, err := h.service.CreateOrder(c.Request.Context(), domain.CreateOrderRequest{
		UserID: c.GetString("user_id"), ProductID: req.ProductID, Quantity: req.Quantity,
		Items: items, IdempotencyKey: key,
	})
	if err != nil {
		writeOrderError(c, err)
		return
	}
	status := http.StatusCreated
	if result.Replayed {
		status = http.StatusOK
	}
	c.JSON(status, result)
}

func (h *orderHandler) listOrders(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page must be an integer"})
		return
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be an integer"})
		return
	}
	orders, err := h.service.ListForUser(c.Request.Context(), c.GetString("user_id"), page, limit)
	if err != nil {
		writeOrderError(c, err)
		return
	}
	c.JSON(http.StatusOK, orders)
}
func (h *orderHandler) getOrder(c *gin.Context) {
	orderID := c.Param("id")
	if !validUUID(orderID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}
	order, err := h.service.GetForUser(c.Request.Context(), c.GetString("user_id"), orderID)
	if err != nil {
		writeOrderError(c, err)
		return
	}
	c.JSON(http.StatusOK, order)
}
func (h *orderHandler) cancelOrder(c *gin.Context) {
	orderID := c.Param("id")
	if !validUUID(orderID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}
	order, err := h.service.CancelForUser(c.Request.Context(), c.GetString("user_id"), orderID)
	if err != nil {
		writeOrderError(c, err)
		return
	}
	c.JSON(http.StatusOK, order)
}
func writeOrderError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidOrder), errors.Is(err, service.ErrInvalidOrderQuery):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order request"})
	case errors.Is(err, service.ErrOrderIdempotencyConflict):
		c.JSON(http.StatusConflict, gin.H{"error": "Idempotency-Key was already used with a different order payload"})
	case errors.Is(err, service.ErrUserNotFound):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authenticated user not found; please log in again"})
	case errors.Is(err, service.ErrProductNotFound), errors.Is(err, service.ErrOrderNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "order or product not found"})
	case errors.Is(err, service.ErrInsufficientStock), errors.Is(err, service.ErrOrderNotCancellable), errors.Is(err, service.ErrOrderInventoryConflict):
		c.JSON(http.StatusConflict, gin.H{"error": "order operation conflicts with current state or inventory"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "order operation failed"})
	}
}
