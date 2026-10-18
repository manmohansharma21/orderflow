package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/manmohansharma21/orderflow/internal/domain"
)

type ReportRepository struct{ pool *pgxpool.Pool }

func NewReportRepository(pool *pgxpool.Pool) *ReportRepository { return &ReportRepository{pool: pool} }

func (r *ReportRepository) Orders(ctx context.Context) (domain.OrdersOperationalReport, error) {
	var report domain.OrdersOperationalReport
	report.GeneratedAt = time.Now().UTC()
	err := r.pool.QueryRow(ctx, `
        SELECT COUNT(*),
          COUNT(*) FILTER (WHERE status='PENDING'),
          COUNT(*) FILTER (WHERE status='CONFIRMED'),
          COUNT(*) FILTER (WHERE status='FAILED'),
          COUNT(*) FILTER (WHERE status='CANCELLED'),
          COUNT(*) FILTER (WHERE status='EXPIRED'),
          COALESCE(SUM(total_amount_paise) FILTER (WHERE status='CONFIRMED'), 0)
        FROM orders
    `).Scan(&report.TotalOrders, &report.PendingOrders, &report.ConfirmedOrders,
		&report.FailedOrders, &report.CancelledOrders, &report.ExpiredOrders, &report.ConfirmedRevenuePaise)
	return report, err
}

func (r *ReportRepository) Sales(ctx context.Context, from, to time.Time) (domain.SalesReport, error) {
	report := domain.SalesReport{GeneratedAt: time.Now().UTC(), FromDate: from.Format("2006-01-02"), ToDate: to.Add(-time.Nanosecond).Format("2006-01-02")}
	err := r.pool.QueryRow(ctx, `
        SELECT
          (SELECT COUNT(*) FROM orders WHERE status='CONFIRMED' AND updated_at >= $1 AND updated_at < $2),
          (SELECT COALESCE(SUM(total_amount_paise),0) FROM orders WHERE status='CONFIRMED' AND updated_at >= $1 AND updated_at < $2),
          (SELECT COUNT(*) FROM payments WHERE status='SUCCESS' AND updated_at >= $1 AND updated_at < $2),
          (SELECT COALESCE(SUM(amount_paise),0) FROM payments WHERE status='SUCCESS' AND updated_at >= $1 AND updated_at < $2),
          (SELECT COUNT(*) FROM payments WHERE status='FAILED' AND updated_at >= $1 AND updated_at < $2),
          (SELECT COALESCE(SUM(amount_paise),0) FROM payments WHERE status='FAILED' AND updated_at >= $1 AND updated_at < $2)
    `, from, to).Scan(&report.ConfirmedOrders, &report.ConfirmedRevenuePaise,
		&report.SuccessfulPayments, &report.SuccessfulPaymentPaise,
		&report.FailedPayments, &report.FailedPaymentPaise)
	return report, err
}

func (r *ReportRepository) Inventory(ctx context.Context, lowStockThreshold int64) (domain.InventoryReport, error) {
	report := domain.InventoryReport{GeneratedAt: time.Now().UTC(), LowStockThreshold: lowStockThreshold, LowStockItems: make([]domain.InventoryReportItem, 0)}
	err := r.pool.QueryRow(ctx, `
        SELECT COUNT(*), COALESCE(SUM(i.on_hand_quantity),0),
               COALESCE(SUM(i.reserved_quantity),0),
               COALESCE(SUM(i.on_hand_quantity-i.reserved_quantity),0),
               COUNT(*) FILTER (WHERE i.on_hand_quantity-i.reserved_quantity <= $1)
        FROM products p JOIN inventory i ON i.product_id=p.id
        WHERE p.is_active=TRUE
    `, lowStockThreshold).Scan(&report.ActiveProducts, &report.TotalOnHand,
		&report.TotalReserved, &report.TotalAvailable, &report.LowStockProducts)
	if err != nil {
		return domain.InventoryReport{}, err
	}
	rows, err := r.pool.Query(ctx, `
        SELECT p.id::text, p.sku, p.name, i.on_hand_quantity, i.reserved_quantity,
               i.on_hand_quantity-i.reserved_quantity
        FROM products p JOIN inventory i ON i.product_id=p.id
        WHERE p.is_active=TRUE AND i.on_hand_quantity-i.reserved_quantity <= $1
        ORDER BY (i.on_hand_quantity-i.reserved_quantity), p.sku
        LIMIT 100
    `, lowStockThreshold)
	if err != nil {
		return domain.InventoryReport{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var item domain.InventoryReportItem
		if err := rows.Scan(&item.ProductID, &item.SKU, &item.Name, &item.OnHandQuantity, &item.ReservedQuantity, &item.AvailableQuantity); err != nil {
			return domain.InventoryReport{}, err
		}
		report.LowStockItems = append(report.LowStockItems, item)
	}
	if err := rows.Err(); err != nil {
		return domain.InventoryReport{}, err
	}
	return report, nil
}
