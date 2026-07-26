package service

import (
	"context"
	"doc-api/config"
	"doc-api/internal/adapter/repository"
	"doc-api/internal/core/domain/entity"
	"doc-api/utils"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type userService struct {
	repo       repository.UserRepositoryInterface
	cfg        *config.Config
	jwtService JwtServiceInterface
	redis      *redis.Client
}

type UserServiceInterface interface {
	SignIn(ctx context.Context, req entity.UserEntity) (*entity.UserEntity, string, error)
	CreateUserAccount(ctx context.Context, req entity.UserEntity) error
	UpdatePassword(ctx context.Context, req entity.UserEntity) error
	GetProfileUser(ctx context.Context, userID int64) (*entity.UserEntity, error)
	UpdateDataUser(ctx context.Context, req entity.UserEntity) error
	GetUserAll(ctx context.Context, query entity.QueryStringUser) ([]entity.UserEntity, int64, int64, error)
	DeleteUser(ctx context.Context, userID int64) error
}

func NewUserService(repo repository.UserRepositoryInterface, cfg *config.Config,
	jwtService JwtServiceInterface, redis *redis.Client) UserServiceInterface {
	return &userService{
		repo:       repo,
		cfg:        cfg,
		jwtService: jwtService,
		redis:      redis,
	}
}

func (u *userService) DeleteUser(ctx context.Context, userID int64) error {
	return u.repo.DeleteUser(ctx, userID)
}

func (u *userService) GetUserAll(ctx context.Context, query entity.QueryStringUser) ([]entity.UserEntity, int64, int64, error) {
	return u.repo.GetUserAll(ctx, query)
}

func (u *userService) UpdateDataUser(ctx context.Context, req entity.UserEntity) error {
	return u.repo.UpdateDataUser(ctx, req)
}

func (u *userService) GetProfileUser(ctx context.Context, userID int64) (*entity.UserEntity, error) {
	return u.repo.GetUserByID(ctx, userID)
}

func (u *userService) UpdatePassword(ctx context.Context, req entity.UserEntity) error {
	password, err := utils.HashPassword(req.Password)
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.core.userService.UpdatePassword")
		return err
	}

	req.Password = password

	err = u.repo.UpdatePasswordByID(ctx, req)
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.core.userService.UpdatePassword")
		return err
	}

	return nil
}

func (u *userService) CreateUserAccount(ctx context.Context, req entity.UserEntity) error {
	password, err := utils.HashPassword(req.Password)
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.core.userService.CreateUserAccount")
		return err
	}

	req.Password = password

	_, err = u.repo.CreateUserAccount(ctx, req)
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.core.userService.CreateUserAccount")
		return err
	}

	return nil
}

func (u *userService) SignIn(ctx context.Context, req entity.UserEntity) (*entity.UserEntity, string, error) {
	user, err := u.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.core.userService.SignIn")
		return nil, "", err
	}

	if checkPass := utils.CheckPasswordHash(req.Password, user.Password); !checkPass {
		err = errors.New("password is incorrect")
		log.Error().
			Err(err).
			Str("source", "internal.core.userService.SignIn")
		return nil, "", err
	}

	token, err := u.jwtService.GenerateToken(user.ID)
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.core.userService.SignIn")
		return nil, "", err
	}

	sessionData := map[string]any{
		"user_id":    user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"logged_in":  true,
		"created_at": time.Now().String(),
		"token":      token,
	}

	jsonData, err := json.Marshal(sessionData)
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.core.userService.SignIn").
			Msg("Error encoding JSON")
		return nil, "", err
	}

	err = u.redis.Set(ctx, token, jsonData, time.Hour*23).Err()
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.core.userService.SignIn")
		return nil, "", err
	}

	return user, token, nil
}
