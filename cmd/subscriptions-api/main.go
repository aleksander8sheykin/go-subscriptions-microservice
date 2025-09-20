package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"subscriptions-service/internal/config"
	"subscriptions-service/internal/database"
	"subscriptions-service/internal/handlers"
	"subscriptions-service/internal/logger"
	"subscriptions-service/internal/repository"
	"subscriptions-service/internal/trace"

	docs "subscriptions-service/internal/swagger"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	cfg := config.LoadConfig()

	db, err := database.Connect(cfg)
	if err != nil {
		logger.Log.Error("Ошибка подключения к БД", "error", err)
		return
	}

	repo := repository.NewSubscriptionRepository(db)
	h := handlers.NewHandler(repo)

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		traceID := fmt.Sprintf("%d", rand.Int63())
		c.Set("trace_id", traceID)

		ctx := trace.WithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)

		c.Writer.Header().Set("X-Trace-ID", traceID)

		start := time.Now()
		c.Next()
		latency := time.Since(start)

		logger.Log.Info("HTTP request",
			"trace_id", traceID,
			"method", c.Request.Method,
			"path", c.Request.URL.RequestURI(),
			"status", c.Writer.Status(),
			"latency_ms", latency.Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	})

	h.RegisterRoutes(r)

	docs.SwaggerInfo.BasePath = "/"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	srv := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Log.Info("Сервис запущен", "port", cfg.ServerAddress)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Error("Ошибка запуска сервера", "error", err)
		}
	}()

	<-quit
	logger.Log.Info("Получен сигнал завершения работы, отключаем сервис...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Error("Ошибка при завершении сервера", "error", err)
	}

	sqlDB, err := db.DB()
	if err == nil {
		if err := sqlDB.Close(); err != nil {
			logger.Log.Error("Ошибка закрытия соединения с БД", "error", err)
		}
	}

	logger.Log.Info("Сервис завершил работу корректно")
}
