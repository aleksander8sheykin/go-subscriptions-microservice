package database

import (
	"fmt"
	"log/slog"
	"subscriptions-service/internal/config"
	"subscriptions-service/internal/logger"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func Connect(cfg config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)

	var gormLevel gormlogger.LogLevel
	switch cfg.LogLevel {
	case "debug":
		gormLevel = gormlogger.Info // GORM не умеет debug, Info → самый подробный
	case "warn":
		gormLevel = gormlogger.Warn
	case "error":
		gormLevel = gormlogger.Error
	default:
		gormLevel = gormlogger.Silent
	}

	newLogger := gormlogger.New(
		slog.NewLogLogger(logger.Log.Handler(), slog.LevelDebug), // slog → gorm
		gormlogger.Config{
			SlowThreshold: time.Second,
			LogLevel:      gormLevel,
			Colorful:      false,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		return nil, err
	}
	logger.Log.Info("Подключение к базе данных успешно")
	return db, nil
}
