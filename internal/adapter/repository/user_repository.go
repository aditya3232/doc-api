package repository

import (
	"context"
	"doc-api/internal/core/domain/entity"
	"doc-api/internal/core/domain/model"
	"errors"
	"fmt"
	"math"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

type UserRepositoryInterface interface {
	GetUserByEmail(ctx context.Context, email string) (*entity.UserEntity, error)
	CreateUserAccount(ctx context.Context, req entity.UserEntity) (int64, error)
	GetUserByID(ctx context.Context, userID int64) (*entity.UserEntity, error)
	UpdateDataUser(ctx context.Context, req entity.UserEntity) error
	GetUserAll(ctx context.Context, query entity.QueryStringUser) ([]entity.UserEntity, int64, int64, error)
	DeleteUser(ctx context.Context, userID int64) error
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
	updateData := map[string]any{
		"name":    req.Name,
		"email":   req.Email,
		"address": req.Address,
		"phone":   req.Phone,
		"photo":   req.Photo,
	}

	if req.Password != "" {
		updateData["password"] = req.Password
	}

	result := u.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", req.ID).Updates(updateData)
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

func (u *userRepository) GetUserAll(ctx context.Context, query entity.QueryStringUser) ([]entity.UserEntity, int64, int64, error) {
	users := []model.User{}
	var countData int64

	order := fmt.Sprintf("%s %s", query.OrderBy, query.OrderType)
	offset := (query.Page - 1) * query.Limit

	sqlMain := u.db.WithContext(ctx).
		Where("name ILIKE ? OR email ILIKE ? OR phone ILIKE ?", "%"+query.Search+"%", "%"+query.Search+"%", "%"+query.Search+"%")

	if err := sqlMain.Model(&users).Count(&countData).Error; err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userRepository.GetUserAll")
		return nil, 0, 0, err
	}

	totalPage := int(math.Ceil(float64(countData) / float64(query.Limit)))

	if err := sqlMain.Order(order).Limit(int(query.Limit)).Offset(int(offset)).Find(&users).Error; err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userRepository.GetUserAll")
		return nil, 0, 0, err
	}

	if len(users) < 1 {
		err := errors.New("404")
		log.Info().
			Str("source", "internal.adapter.userRepository.GetUserAll").
			Msg("No user found")
		return nil, 0, 0, err
	}

	resp := []entity.UserEntity{}
	for _, val := range users {
		resp = append(resp, entity.UserEntity{
			ID:      val.ID,
			Name:    val.Name,
			Email:   val.Email,
			Phone:   val.Phone,
			Photo:   val.Photo,
			Address: val.Address,
		})
	}

	return resp, countData, int64(totalPage), nil
}

func (u *userRepository) DeleteUser(ctx context.Context, userID int64) error {
	var user model.User

	err := u.db.Transaction(func(tx *gorm.DB) error {
		if err := u.db.WithContext(ctx).Where("id =?", userID).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = errors.New("404")
				log.Info().
					Str("source", "internal.adapter.userRepository.DeleteUser").
					Msg("No customer found")
				return err
			}
			log.Error().
				Err(err).
				Str("source", "internal.adapter.userRepository.DeleteUser")
			return err
		}

		if err := u.db.WithContext(ctx).Delete(&user).Error; err != nil {
			log.Error().
				Err(err).
				Str("source", "internal.adapter.userRepository.DeleteUser")
			return err
		}

		return nil
	})

	return err
}
