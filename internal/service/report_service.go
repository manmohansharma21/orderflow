package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/manmohansharma21/orderflow/internal/domain"
	"github.com/manmohansharma21/orderflow/internal/repository"
)

var ErrInvalidReportRange = errors.New("invalid report date range")

type ReportService struct{ repo *repository.ReportRepository }

func NewReportService(repo *repository.ReportRepository) *ReportService {
	return &ReportService{repo: repo}
}
func (s *ReportService) Orders(ctx context.Context) (domain.OrdersOperationalReport, error) {
	return s.repo.Orders(ctx)
}
func (s *ReportService) Sales(ctx context.Context, fromText, toText string) (domain.SalesReport, error) {
	fromText, toText = strings.TrimSpace(fromText), strings.TrimSpace(toText)
	now := time.Now().UTC()
	if toText == "" {
		toText = now.Format("2006-01-02")
	}
	if fromText == "" {
		fromText = now.AddDate(0, 0, -29).Format("2006-01-02")
	}
	fromDate, err := time.Parse("2006-01-02", fromText)
	if err != nil {
		return domain.SalesReport{}, ErrInvalidReportRange
	}
	toDate, err := time.Parse("2006-01-02", toText)
	if err != nil {
		return domain.SalesReport{}, ErrInvalidReportRange
	}
	if toDate.Before(fromDate) || toDate.Sub(fromDate) > 366*24*time.Hour {
		return domain.SalesReport{}, ErrInvalidReportRange
	}
	toExclusive := toDate.AddDate(0, 0, 1)
	return s.repo.Sales(ctx, fromDate.UTC(), toExclusive.UTC())
}
func (s *ReportService) Inventory(ctx context.Context, lowStockThreshold int64) (domain.InventoryReport, error) {
	if lowStockThreshold < 0 || lowStockThreshold > 1_000_000 {
		return domain.InventoryReport{}, ErrInvalidReportRange
	}
	return s.repo.Inventory(ctx, lowStockThreshold)
}
