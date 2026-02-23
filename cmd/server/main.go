package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"imagesv2/internal/config"
	"imagesv2/internal/handler"
	"imagesv2/internal/service"
	"imagesv2/internal/storage"
)

func main() {
	cfg := config.Load()

	store, err := storage.NewFileSystem(cfg.StorageDir)
	if err != nil {
		log.Fatalf("init storage: %v", err)
	}

	svc := service.New(cfg, store)
	h := handler.New(svc)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), handler.LoggingMiddleware())
	h.Register(r)

	srv := &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: r,
	}

	go func() {
		slog.Info("starting server", "addr", cfg.ListenAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown: %v", err)
	}
	slog.Info("server stopped")
}
