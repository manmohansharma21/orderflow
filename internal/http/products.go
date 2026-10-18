package http

import (
	"errors"
	stdhttp "net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"github.com/manmohansharma21/orderflow/internal/domain"
	"github.com/manmohansharma21/orderflow/internal/repository"
	"github.com/manmohansharma21/orderflow/internal/service"
)

var uuidPattern = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type productHandler struct {
	service *service.ProductService
}

type createProductRequest struct {
	SKU          string  `json:"sku" binding:"required,max=100"`
	Name         string  `json:"name" binding:"required,max=255"`
	Description  *string `json:"description"`
	PricePaise   *int64  `json:"price_paise" binding:"required,gte=0"`
	InitialStock int     `json:"initial_stock" binding:"gte=0"`
}

type updateProductRequest struct {
	SKU         *string `json:"sku"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	PricePaise  *int64  `json:"price_paise"`
	IsActive    *bool   `json:"is_active"`
}

func newProductHandler(s *service.ProductService) *productHandler {
	return &productHandler{service: s}
}

func (h *productHandler) list(c *gin.Context) {
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

	result, err := h.service.List(c.Request.Context(), c.Query("search"), page, limit)
	if err != nil {
		c.JSON(stdhttp.StatusInternalServerError, gin.H{"error": "could not list products"})
		return
	}
	c.JSON(stdhttp.StatusOK, result)
}

func (h *productHandler) getByID(c *gin.Context) {
	if !validUUID(c.Param("id")) {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}
	product, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeProductError(c, err)
		return
	}
	c.JSON(stdhttp.StatusOK, product)
}

func (h *productHandler) create(c *gin.Context) {
	var req createProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": "invalid product request"})
		return
	}
	product, err := h.service.Create(c.Request.Context(), domain.CreateProductInput{
		SKU: req.SKU, Name: req.Name, Description: req.Description,
		PricePaise: *req.PricePaise, InitialStock: req.InitialStock,
	})
	if err != nil {
		writeProductError(c, err)
		return
	}
	c.JSON(stdhttp.StatusCreated, product)
}

func (h *productHandler) update(c *gin.Context) {
	id := c.Param("id")
	if !validUUID(id) {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}
	var req updateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": "invalid product request"})
		return
	}
	product, err := h.service.Update(c.Request.Context(), id, domain.UpdateProductInput{
		SKU: req.SKU, Name: req.Name, Description: req.Description,
		PricePaise: req.PricePaise, IsActive: req.IsActive,
	})
	if err != nil {
		writeProductError(c, err)
		return
	}
	c.JSON(stdhttp.StatusOK, product)
}

func (h *productHandler) deactivate(c *gin.Context) {
	id := c.Param("id")
	if !validUUID(id) {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}
	product, err := h.service.Deactivate(c.Request.Context(), id)
	if err != nil {
		writeProductError(c, err)
		return
	}
	c.JSON(stdhttp.StatusOK, product)
}

func writeProductError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidProduct):
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": "invalid product"})
	case errors.Is(err, repository.ErrDuplicateSKU):
		c.JSON(stdhttp.StatusConflict, gin.H{"error": "product SKU already exists"})
	case errors.Is(err, pgx.ErrNoRows):
		c.JSON(stdhttp.StatusNotFound, gin.H{"error": "product not found"})
	default:
		c.JSON(stdhttp.StatusInternalServerError, gin.H{"error": "product operation failed"})
	}
}

func validUUID(value string) bool {
	return uuidPattern.MatchString(value)
}
