package main

import (
	"subscriptions-service/internal/config"
	"subscriptions-service/internal/database"
	"subscriptions-service/internal/handlers"
	"subscriptions-service/internal/logger"
	"subscriptions-service/internal/repository"

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
	}

	repo := repository.NewSubscriptionRepository(db)
	h := handlers.NewHandler(repo)
	r := gin.Default()
	h.RegisterRoutes(r)

	docs.SwaggerInfo.BasePath = "/"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	if err := r.Run(cfg.ServerAddress); err != nil {
		logger.Log.Error("Ошибка запуска сервера: %v", "error", err)
	} else {
		logger.Log.Info("Сервис запущен", "port", cfg.ServerAddress)
	}
}
