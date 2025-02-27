package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/renanqts/external-dns-openwrt-webhook/internal/provider"
	"github.com/renanqts/external-dns-openwrt-webhook/pkg/config"
	"github.com/renanqts/external-dns-openwrt-webhook/pkg/logger"
	"github.com/renanqts/external-dns-openwrt-webhook/pkg/router"
	"github.com/renanqts/external-dns-openwrt-webhook/pkg/webhook"
	"go.uber.org/zap"
)

func main() {
	cfg := defaultConfig()
	if err := config.Read(cfg); err != nil {
		panic(err)
	}

	if err := logger.Init(cfg.Log); err != nil {
		panic(err)
	}

	provider, err := provider.New(cfg.Provider)
	if err != nil {
		logger.Log.Fatal("failed to setup provider", zap.Error(err))
	}

	router, err := router.New(cfg.Router)
	if err != nil {
		logger.Log.Fatal("failed to setup router", zap.Error(err))
	}

	webhook := webhook.New(provider)
	setupRoutes(router.GetEngine(), webhook)

	go func() {
		if err = router.Run(); err != nil {
			logger.Log.Fatal("failed to start server", zap.Error(err))
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	logger.Log.Info("termination signal received, shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ShutodwnTimeout)*time.Second)
	if err := router.Shutdown(ctx); err != nil {
		logger.Log.Error("failed to shutdown server", zap.Error(err))
	}

	cancel()
	<-ctx.Done()

	logger.Log.Info("service shutdown completed")
}

func setupRoutes(r *gin.Engine, webhook *webhook.Webhook) {
	apiGroup := r.Group("/")
	apiGroup.GET("/", webhook.Negotiate)
	apiGroup.GET("/records", webhook.Records)
	apiGroup.POST("/records", webhook.ApplyChanges)
	apiGroup.POST("/adjustendpoints", webhook.AdjustEndpoints)
}
