package database

import (
	"context"
	"fmt"
	"log/slog"
	"subscriptions-service/internal/config"
	"subscriptions-service/internal/logger"
	"subscriptions-service/internal/trace"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type gormLoggerWithTrace struct {
	inner gormlogger.Interface
}

func (g *gormLoggerWithTrace) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return &gormLoggerWithTrace{inner: g.inner.LogMode(level)}
}

func (g *gormLoggerWithTrace) Info(ctx context.Context, msg string, data ...interface{}) {
	logger.Log.Info(msg, data...)
}

func (g *gormLoggerWithTrace) Warn(ctx context.Context, msg string, data ...interface{}) {
	logger.Log.Warn(msg, data...)
}

func (g *gormLoggerWithTrace) Error(ctx context.Context, msg string, data ...interface{}) {
	logger.Log.Error(msg, data...)
}

func (g *gormLoggerWithTrace) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, rows := fc()
	traceID := trace.TraceIDFromContext(ctx)

	logger.Log.Debug("SQL executed",
		"trace_id", traceID,
		"sql", sql,
		"rows", rows,
		"duration_ms", time.Since(begin).Milliseconds(),
		"error", err,
	)
}

func Connect(cfg config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)

	var gormLevel gormlogger.LogLevel
	switch cfg.LogLevel {
	case "debug":
		gormLevel = gormlogger.Info
	case "warn":
		gormLevel = gormlogger.Warn
	case "error":
		gormLevel = gormlogger.Error
	default:
		gormLevel = gormlogger.Silent
	}

	newLogger := &gormLoggerWithTrace{inner: gormlogger.New(
		slog.NewLogLogger(logger.Log.Handler(), slog.LevelDebug),
		gormlogger.Config{
			SlowThreshold: time.Second,
			LogLevel:      gormLevel,
			Colorful:      false,
		},
	)}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, err
	}

	logger.Log.Info("Подключение к базе данных успешно")
	return db, nil
}
