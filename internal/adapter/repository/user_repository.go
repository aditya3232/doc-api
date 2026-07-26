package repository

import (
	"context"
	"doc-api/internal/core/domain/entity"
	"doc-api/internal/core/domain/model"
	"errors"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

type UserRepositoryInterface interface {
	GetUserByEmail(ctx context.Context, email string) (*entity.UserEntity, error)
	CreateUserAccount(ctx context.Context, req entity.UserEntity) (int64, error)
	UpdatePasswordByID(ctx context.Context, req entity.UserEntity) error
	GetUserByID(ctx context.Context, userID int64) (*entity.UserEntity, error)
	UpdateDataUser(ctx context.Context, req entity.UserEntity) error
}

func NewUserRepository(db *gorm.DB) UserRepositoryInterface {
	return &userRepository{db: db}
}

func (u *userRepository) GetUserByEmail(ctx context.Context, email string) (*entity.UserEntity, error) {
	modelUser := model.User{}

	result := u.db.WithContext(ctx).Where("email = ?", email).First(&modelUser)
	if result.Error != nil {
		log.Error().
			Err(result.Error).
			Str("source", "internal.adapter.userRepository.GetUserByEmail")
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		log.Info().
			Str("source", "internal.adapter.userRepository.GetUserByEmail").
			Msg("User not found")
		return nil, errors.New("404")
	}

	return &entity.UserEntity{
		ID:       modelUser.ID,
		Name:     modelUser.Name,
		Email:    email,
		Password: modelUser.Password,
		Address:  modelUser.Address,
		Phone:    modelUser.Phone,
		Photo:    modelUser.Photo,
	}, nil
}

func (u *userRepository) CreateUserAccount(ctx context.Context, req entity.UserEntity) (int64, error) {
	user := model.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	if err := u.db.WithContext(ctx).Create(&user).Error; err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userRepository.CreateUserAccount").
			Msg("failed create user")

		return 0, err
	}

	return user.ID, nil
}

func (u *userRepository) UpdatePasswordByID(ctx context.Context, req entity.UserEntity) error {
	user := model.User{}

	err := u.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Where("id = ?", req.ID).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = errors.New("404")
				log.Error().
					Err(err).
					Str("source", "internal.adapter.userRepository.UpdatePasswordByID")
				return err
			}

			log.Error().
				Err(err).
				Str("source", "internal.adapter.userRepository.UpdatePasswordByID")

			return err
		}

		if err := tx.Model(&user).Update("password", req.Password).Error; err != nil {
			log.Error().
				Err(err).
				Str("source", "internal.adapter.userRepository.UpdatePasswordByID").
				Msg("failed update password")

			return err
		}

		return nil
	})

	return err
}

func (u *userRepository) GetUserByID(ctx context.Context, userID int64) (*entity.UserEntity, error) {
	var user model.User

	if err := u.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")

			log.Error().
				Err(err).
				Str("source", "internal.adapter.userRepository.GetUserByID").
				Msg("user not found")

			return nil, err
		}

		log.Error().
			Err(err).
			Str("source", "internal.adapter.userRepository.GetUserByID").
			Msg("failed get user")

		return nil, err
	}

	return &entity.UserEntity{
		ID:      user.ID,
		Name:    user.Name,
		Email:   user.Email,
		Phone:   user.Phone,
		Photo:   user.Photo,
		Address: user.Address,
	}, nil
}

func (u *userRepository) UpdateDataUser(ctx context.Context, req entity.UserEntity) error {
	result := u.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", req.ID).Updates(map[string]any{
		"name":    req.Name,
		"email":   req.Email,
		"address": req.Address,
		"phone":   req.Phone,
		"photo":   req.Photo,
	})

	if result.Error != nil {
		log.Error().
			Err(result.Error).
			Str("source", "internal.adapter.userRepository.UpdateDataUser").
			Msg("failed update user")

		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}
