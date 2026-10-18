package http

import (
	"errors"
	stdhttp "net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"github.com/manmohansharma21/orderflow/internal/service"
)

type productHandler struct {
	service *service.ProductService
}

func newProductHandler(s *service.ProductService) *productHandler {
	return &productHandler{service: s}
}

func (h *productHandler) list(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	result, err := h.service.List(
		c.Request.Context(),
		c.Query("search"),
		page,
		limit,
	)
	if err != nil {
		c.JSON(stdhttp.StatusInternalServerError, gin.H{
			"error": "could not list products",
		})
		return
	}

	c.JSON(stdhttp.StatusOK, result)
}

func (h *productHandler) getByID(c *gin.Context) {
	product, err := h.service.GetByID(
		c.Request.Context(),
		c.Param("id"),
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(stdhttp.StatusNotFound, gin.H{
				"error": "product not found",
			})
			return
		}

		c.JSON(stdhttp.StatusInternalServerError, gin.H{
			"error": "could not get product",
		})
		return
	}

	c.JSON(stdhttp.StatusOK, product)
}
