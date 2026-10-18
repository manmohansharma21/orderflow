package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/manmohansharma21/orderflow/internal/service"
)

type reportHandler struct{ service *service.ReportService }

func newReportHandler(s *service.ReportService) *reportHandler { return &reportHandler{service: s} }
func (h *reportHandler) orders(c *gin.Context) {
	report, err := h.service.Orders(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not build order report"})
		return
	}
	c.JSON(http.StatusOK, report)
}
func (h *reportHandler) sales(c *gin.Context) {
	report, err := h.service.Sales(c.Request.Context(), c.Query("from"), c.Query("to"))
	if errors.Is(err, service.ErrInvalidReportRange) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from/to must be valid YYYY-MM-DD dates, from must not exceed to, and range must be at most 367 days"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not build sales report"})
		return
	}
	c.JSON(http.StatusOK, report)
}
func (h *reportHandler) inventory(c *gin.Context) {
	threshold, err := strconv.ParseInt(c.DefaultQuery("low_stock_threshold", "5"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "low_stock_threshold must be an integer"})
		return
	}
	report, err := h.service.Inventory(c.Request.Context(), threshold)
	if errors.Is(err, service.ErrInvalidReportRange) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "low_stock_threshold must be between 0 and 1000000"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not build inventory report"})
		return
	}
	c.JSON(http.StatusOK, report)
}
