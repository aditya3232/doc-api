package handler

import (
	"doc-api/config"
	"doc-api/internal/adapter/handler/response"
	"doc-api/internal/adapter/storage"
	"doc-api/internal/core/service"
	"doc-api/internal/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type uploadImage struct {
	storageHandler storage.MinioStorageInterface
}

type UploadImageInterface interface {
	UploadImage(c fiber.Ctx) error
}

func NewUploadImage(
	app *fiber.App,
	cfg *config.Config,
	storageHandler storage.MinioStorageInterface,
	jwtService service.JwtServiceInterface,
	redis *redis.Client,
) UploadImageInterface {
	res := &uploadImage{
		storageHandler: storageHandler,
	}

	mid := middleware.NewMiddlewareAdapter(cfg, jwtService, redis)

	// auth route
	authGroup := app.Group("/auth", mid.CheckToken())
	authGroup.Post("/profile/image-upload", res.UploadImage)

	return res
}

func (u *uploadImage) UploadImage(c fiber.Ctx) error {
	fileHeader, err := c.FormFile("photo")
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.uploadImage.UploadImage").
			Msg("failed get uploaded file")

		return fiber.NewError(
			fiber.StatusUnprocessableEntity,
			err.Error(),
		)
	}

	ctx := c.Context()

	url, err := u.storageHandler.ProcessAndUploadImage(
		ctx,
		fileHeader,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("filename", fileHeader.Filename).
			Str("source", "internal.adapter.uploadImage.UploadImage").
			Msg("failed upload image")

		return fiber.NewError(
			fiber.StatusInternalServerError,
			err.Error(),
		)
	}

	return c.Status(fiber.StatusOK).JSON(response.DefaultResponse{
		Message: "success",
		Data: map[string]string{
			"image_url": url,
		},
	})
}
