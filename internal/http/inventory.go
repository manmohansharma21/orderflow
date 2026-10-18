package http

import (
	"errors"
	stdhttp "net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"github.com/manmohansharma21/orderflow/internal/domain"
	"github.com/manmohansharma21/orderflow/internal/service"
)

type inventoryHandler struct {
	service *service.InventoryService
}

type receiveStockRequest struct {
	ProductID string `json:"product_id" binding:"required,uuid"`
	Quantity  int    `json:"quantity" binding:"required,min=1,max=1000000"`
	Reason    string `json:"reason" binding:"required,min=3,max=500"`
}

type adjustStockRequest struct {
	ProductID     string `json:"product_id" binding:"required,uuid"`
	QuantityDelta int    `json:"quantity_delta" binding:"required,min=-1000000,max=1000000,ne=0"`
	Reason        string `json:"reason" binding:"required,min=3,max=500"`
}

func newInventoryHandler(s *service.InventoryService) *inventoryHandler {
	return &inventoryHandler{service: s}
}

func (h *inventoryHandler) getAvailability(c *gin.Context) {
	productID := c.Param("product_id")
	if !validUUID(productID) {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}
	item, err := h.service.GetAvailability(c.Request.Context(), productID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, service.ErrProductNotFound) {
			c.JSON(stdhttp.StatusNotFound, gin.H{"error": "product inventory not found"})
			return
		}
		c.JSON(stdhttp.StatusInternalServerError, gin.H{"error": "could not get inventory"})
		return
	}
	c.JSON(stdhttp.StatusOK, item)
}

func (h *inventoryHandler) receiveStock(c *gin.Context) {
	var req receiveStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": "invalid stock receipt request"})
		return
	}
	item, err := h.service.ReceiveStock(c.Request.Context(), domain.InventoryMovementInput{
		ProductID:       req.ProductID,
		OnHandDelta:     req.Quantity,
		Reason:          req.Reason,
		CreatedByUserID: c.GetString("user_id"),
	})
	if err != nil {
		writeInventoryError(c, err)
		return
	}
	c.JSON(stdhttp.StatusOK, item)
}

func (h *inventoryHandler) adjustStock(c *gin.Context) {
	var req adjustStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": "invalid stock adjustment request"})
		return
	}
	item, err := h.service.AdjustStock(c.Request.Context(), domain.InventoryMovementInput{
		ProductID:       req.ProductID,
		OnHandDelta:     req.QuantityDelta,
		Reason:          req.Reason,
		CreatedByUserID: c.GetString("user_id"),
	})
	if err != nil {
		writeInventoryError(c, err)
		return
	}
	c.JSON(stdhttp.StatusOK, item)
}

func (h *inventoryHandler) listMovements(c *gin.Context) {
	productID := c.Query("product_id")
	if productID != "" && !validUUID(productID) {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": "invalid product_id"})
		return
	}
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": "page must be an integer"})
		return
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": "limit must be an integer"})
		return
	}
	result, err := h.service.ListMovements(
		c.Request.Context(), productID, c.Query("type"), page, limit,
	)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInventoryRequest) {
			c.JSON(stdhttp.StatusBadRequest, gin.H{"error": "invalid movement filter or pagination"})
			return
		}
		c.JSON(stdhttp.StatusInternalServerError, gin.H{"error": "could not list inventory movements"})
		return
	}
	c.JSON(stdhttp.StatusOK, result)
}

func writeInventoryError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInventoryRequest):
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": "invalid inventory request"})
	case errors.Is(err, service.ErrProductNotFound), errors.Is(err, pgx.ErrNoRows):
		c.JSON(stdhttp.StatusNotFound, gin.H{"error": "active product inventory not found"})
	case errors.Is(err, service.ErrInventoryConflict):
		c.JSON(stdhttp.StatusConflict, gin.H{"error": "adjustment would violate available or reserved stock constraints"})
	default:
		c.JSON(stdhttp.StatusInternalServerError, gin.H{"error": "inventory operation failed"})
	}
}
