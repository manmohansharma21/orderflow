package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/manmohansharma21/orderflow/internal/cache"
	"github.com/manmohansharma21/orderflow/internal/config"
	"github.com/manmohansharma21/orderflow/internal/database"
	apphttp "github.com/manmohansharma21/orderflow/internal/http"
	"github.com/manmohansharma21/orderflow/internal/repository"
	"github.com/manmohansharma21/orderflow/internal/service"
	"github.com/manmohansharma21/orderflow/internal/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load configuration: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	cancel()
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer pool.Close()

	var productCache service.ProductCache
	var redisCache *cache.RedisProductCache
	if cfg.RedisEnabled {
		redisCache = cache.NewRedisProductCache(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
		pingCtx, pingCancel := context.WithTimeout(context.Background(), time.Second)
		if err := redisCache.Ping(pingCtx); err != nil {
			log.Printf("WARN: Redis unavailable; product caching disabled for this process: %v", err)
			_ = redisCache.Close()
			redisCache = nil
		} else {
			productCache = redisCache
		}
		pingCancel()
	}
	if redisCache != nil {
		defer redisCache.Close()
	}

	router := apphttp.NewRouterWithProductCache(pool, cfg.ReadinessTimeoutSeconds, productCache, time.Duration(cfg.ProductCacheTTLSeconds)*time.Second)
	workerCtx, stopWorker := context.WithCancel(context.Background())
	defer stopWorker()
	if cfg.ReconciliationIntervalSeconds > 0 {
		payments := service.NewPaymentService(repository.NewPaymentRepository(pool))
		go worker.RunPeriodicReconciliation(workerCtx, payments, time.Duration(cfg.ReconciliationIntervalSeconds)*time.Second, cfg.ReconciliationOlderThanMinutes, log.Default())
		log.Printf("payment reconciliation worker enabled: interval=%ds stale_threshold=%dm", cfg.ReconciliationIntervalSeconds, cfg.ReconciliationOlderThanMinutes)
	} else {
		log.Printf("payment reconciliation worker disabled")
	}

	server := &http.Server{Addr: ":" + cfg.Port, Handler: router, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		log.Printf("OrderFlow API listening on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	stopWorker()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
