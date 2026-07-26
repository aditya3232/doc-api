package service

import (
	"context"
	"doc-api/internal/core/domain/entity"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type healthCheckService struct {
	redis *redis.Client
	db    *gorm.DB
	minio *minio.Client
}

type HealthCheckInterface interface {
	HealthCheck(ctx context.Context) (*entity.HealthCheck, error)
}

func NewHealthCheckService(redis *redis.Client, db *gorm.DB, minio *minio.Client) HealthCheckInterface {
	return &healthCheckService{
		redis: redis,
		db:    db,
		minio: minio,
	}
}

var startedAt = time.Now()

func (u *healthCheckService) HealthCheck(ctx context.Context) (*entity.HealthCheck, error) {
	dbStatus := "UP"
	redisStatus := "UP"
	minioStatus := "UP"

	// db
	sqlDB, err := u.db.DB()
	if err != nil {
		dbStatus = "DOWN"
	} else if err := sqlDB.PingContext(ctx); err != nil {
		dbStatus = "DOWN"
	}

	// redis
	if err := u.redis.Ping(ctx).Err(); err != nil {
		redisStatus = "DOWN"
	}

	// MinIO
	if _, err := u.minio.ListBuckets(ctx); err != nil {
		minioStatus = "DOWN"
	}

	status := "UP"
	if dbStatus == "DOWN" || redisStatus == "DOWN" || minioStatus == "DOWN" {
		status = "DOWN"
	}

	return &entity.HealthCheck{
		Status:    status,
		Service:   "doc-api-service",
		Uptime:    time.Since(startedAt).Round(time.Second).String(),
		Timestamp: time.Now().UTC(),
		Dependencies: map[string]string{
			"database": dbStatus,
			"redis":    redisStatus,
			"minio":    minioStatus,
		},
	}, nil
}
