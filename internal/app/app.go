package app

import (
	"context"
	"doc-api/config"
	"doc-api/internal/adapter/handler"
	"doc-api/internal/adapter/repository"
	"doc-api/internal/adapter/storage"
	"doc-api/internal/core/service"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	fiberCors "github.com/gofiber/fiber/v3/middleware/cors"
	fiberRecover "github.com/gofiber/fiber/v3/middleware/recover"

	"github.com/rs/zerolog/log"
)

func RunServer() {
	cfg := config.NewConfig()

	config.NewLogger(cfg.App.AppEnv, cfg.App.LogLevel, cfg.App.AppName)

	db, err := cfg.ConnectionPostgres()
	if err != nil {
		log.Fatal().
			Err(err).
			Str("source", "internal.app.RunServer").
			Msg("failed connect postgres")
	}

	minio, err := cfg.NewMinio()
	if err != nil {
		log.Fatal().
			Err(err).
			Str("source", "internal.app.RunServer").
			Msg("failed connect to minio")
	}

	redis, err := cfg.NewRedisClient()
	if err != nil {
		log.Fatal().
			Err(err).
			Str("source", "internal.app.RunServer").
			Msg("failed connect to redis")
	}

	storageHandler := storage.NewMinioStorage(cfg, minio)
	userRepo := repository.NewUserRepository(db.DB)
	jwtService := service.NewJwtService(cfg)
	healthCheckService := service.NewHealthCheckService(redis, db.DB, minio, cfg)
	userService := service.NewUserService(userRepo, cfg, jwtService, redis)

	app := cfg.NewFiber()
	app.Use(fiberRecover.New())
	app.Use(fiberCors.New())

	handler.NewHealthCheckHandler(app, healthCheckService)
	handler.NewUserHandler(app, userService, cfg, jwtService, redis)
	handler.NewUploadImage(app, cfg, storageHandler, jwtService, redis)

	go func() {
		if cfg.App.AppPort == "" {
			cfg.App.AppPort = os.Getenv("APP_PORT")
		}

		port := ":" + cfg.App.AppPort

		log.Info().
			Str("port", port).
			Str("source", "internal.app.RunServer").
			Msg("server started")

		err = app.Listen(
			port,
			fiber.ListenConfig{
				EnablePrefork: cfg.App.WebPrefork,
			},
		)

		if err != nil {
			log.Fatal().
				Err(err).
				Str("source", "internal.app.RunServer").
				Msg("failed start server")
		}
	}()

	// =========================
	// Graceful Shutdown
	// =========================
	quit := make(chan os.Signal, 1)

	signal.Notify(
		quit,
		os.Interrupt,
		syscall.SIGTERM,
	)

	<-quit // nunggu sampai ada sinyal cancel baru jalankan kode dibawah

	log.Info().
		Str("source", "internal.app.RunServer").
		Msg("shutting down server in 5 seconds")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.app.RunServer").
			Msg("failed shutdown server")
	}

	log.Info().
		Str("source", "internal.app.RunServer").
		Msg("server stopped gracefully")

}
