package router

import (
	"context"
	"net/http"
	"time"

	"github.com/Depado/ginprom"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/renanqts/external-dns-openwrt-webhook/pkg/logger"
	"go.uber.org/zap"
)

const (
	metrics_namespace     = "external_dns_openwrt_webhook"
	metrics_gin_subsystem = "gin"
	metrics_path          = "/metrics"
)

type Router struct {
	config *Config
	engine *gin.Engine
	srv    *http.Server
}

func New(config *Config) (*Router, error) {
	if config.Gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(ginzap.GinzapWithConfig(logger.Log, &ginzap.Config{
		TimeFormat: time.RFC3339,
		UTC:        true,
		SkipPaths:  []string{metrics_path, config.HealthCheckPath},
	}))
	r.Use(ginzap.RecoveryWithZap(logger.Log, true))

	r.GET(config.HealthCheckPath, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	p := ginprom.New(
		ginprom.Engine(r),
		ginprom.Namespace(metrics_namespace),
		ginprom.Subsystem(metrics_gin_subsystem),
	)
	r.Use(p.Instrument())

	return &Router{
		config: config,
		engine: r,
		srv: &http.Server{
			Addr:    ":" + config.Port,
			Handler: r,
		},
	}, nil
}

func (r *Router) Run() error {
	logger.Log.Info("starting http server", zap.String("port", r.config.Port))
	if err := r.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (r *Router) Shutdown(ctx context.Context) error {
	logger.Log.Debug("shutting down http server")
	if err := r.srv.Shutdown(ctx); err != nil {
		return err
	}
	logger.Log.Info("http server stopped")
	return nil
}

func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}
