package handler

import (
	"doc-api/config"
	"doc-api/internal/adapter/handler/request"
	"doc-api/internal/adapter/handler/response"
	"doc-api/internal/core/domain/entity"
	"doc-api/internal/core/service"
	"doc-api/internal/middleware"
	"doc-api/utils"
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type userHandler struct {
	userService service.UserServiceInterface
}

type UserHandlerInterface interface {
	SignIn(c fiber.Ctx) error
	CreateUserAccount(c fiber.Ctx) error
	UpdatePassword(c fiber.Ctx) error
	GetProfileUser(c fiber.Ctx) error
	UpdateUser(c fiber.Ctx) error
	UpdateProfileUser(c fiber.Ctx) error
	GetUserAll(c fiber.Ctx) error
	DeleteUser(c fiber.Ctx) error
}

func NewUserHandler(
	app *fiber.App,
	userService service.UserServiceInterface,
	cfg *config.Config,
	jwtService service.JwtServiceInterface,
	redis *redis.Client,
) UserHandlerInterface {
	userHandler := &userHandler{
		userService: userService,
	}

	mid := middleware.NewMiddlewareAdapter(cfg, jwtService, redis)

	// public route
	app.Post("/signin", userHandler.SignIn)
	app.Post("/signup", userHandler.CreateUserAccount)
	app.Put("/update-password", userHandler.UpdatePassword)

	// auth route
	authGroup := app.Group("/auth", mid.CheckToken())
	authGroup.Get("/users", userHandler.GetUserAll)
	authGroup.Put("/users/:id", userHandler.UpdateUser)
	authGroup.Put("/profile", userHandler.UpdateProfileUser)
	authGroup.Get("/profile", userHandler.GetProfileUser)
	authGroup.Delete("/users/:id", userHandler.DeleteUser)

	return userHandler
}

func (u *userHandler) SignIn(c fiber.Ctx) error {
	var req request.SignInRequest

	ctx := c.Context()

	if err := c.Bind().Body(&req); err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userHandler.SignIn").
			Msg("failed bind/validate request")

		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	reqEntity := entity.UserEntity{
		Email:    req.Email,
		Password: req.Password,
	}

	user, token, err := u.userService.SignIn(ctx, reqEntity)
	if err != nil {
		log.Error().
			Err(err).
			Str("email", req.Email).
			Str("source", "internal.adapter.userHandler.SignIn").
			Msg("failed sign in")

		if err.Error() == "404" {
			return fiber.NewError(fiber.StatusNotFound, "user not found")
		}

		return err
	}

	respSignIn := response.SignInResponse{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		Phone:       user.Phone,
		AccessToken: token,
	}

	return c.Status(fiber.StatusOK).JSON(response.DefaultResponse{
		Message: "success",
		Data:    respSignIn,
	})
}

func (u *userHandler) CreateUserAccount(c fiber.Ctx) error {
	var req request.SignUpRequest

	ctx := c.Context()

	if err := c.Bind().Body(&req); err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userHandler.CreateUserAccount").
			Msg("failed bind/validate request")

		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if req.Password != req.PasswordConfirmation {
		log.Error().
			Str("source", "internal.adapter.userHandler.CreateUserAccount").
			Msg("password confirmation mismatch")

		return fiber.NewError(fiber.StatusUnprocessableEntity, "passwords do not match")
	}

	reqEntity := entity.UserEntity{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	if err := u.userService.CreateUserAccount(ctx, reqEntity); err != nil {
		log.Error().
			Err(err).
			Str("email", req.Email).
			Str("source", "internal.adapter.userHandler.CreateUserAccount").
			Msg("failed create user account")

		return err
	}

	return c.Status(fiber.StatusCreated).JSON(response.DefaultResponse{
		Message: "success",
		Data:    nil,
	})
}

func (u *userHandler) GetProfileUser(c fiber.Ctx) error {
	var jwtUserData entity.JwtUserData

	ctx := c.Context()

	user, ok := c.Locals("user").(string)
	if !ok || user == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "data token not found")
	}

	if err := json.Unmarshal([]byte(user), &jwtUserData); err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userHandler.GetProfileUser").
			Msg("failed parse jwt user data")

		return fiber.NewError(fiber.StatusBadRequest, "invalid token data")
	}

	dataUser, err := u.userService.GetProfileUser(ctx, jwtUserData.UserID)
	if err != nil {
		log.Error().
			Err(err).
			Int64("user_id", jwtUserData.UserID).
			Str("source", "internal.adapter.userHandler.GetProfileUser").
			Msg("failed get profile user")

		if err.Error() == "404" {
			return fiber.NewError(fiber.StatusNotFound, "user not found")
		}

		return err
	}

	respProfile := response.ProfileResponse{
		ID:      dataUser.ID,
		Name:    dataUser.Name,
		Email:   dataUser.Email,
		Phone:   dataUser.Phone,
		Address: dataUser.Address,
		Photo:   dataUser.Photo,
	}

	return c.Status(fiber.StatusOK).JSON(response.DefaultResponse{
		Message: "success",
		Data:    respProfile,
	})
}

func (u *userHandler) DeleteUser(c fiber.Ctx) error {
	ctx := c.Context()

	user, ok := c.Locals("user").(string)
	if !ok || user == "" {
		log.Error().
			Str("source", "internal.adapter.userHandler.DeleteUser").
			Msg("data token not found")

		return fiber.NewError(fiber.StatusUnauthorized, "data token not valid")
	}

	idParamStr := c.Params("id")
	if idParamStr == "" {
		log.Info().
			Str("source", "internal.adapter.userHandler.DeleteUser").
			Msg("missing or invalid user ID")

		return fiber.NewError(fiber.StatusBadRequest, "missing or invalid user ID")
	}

	id, err := utils.StringToInt64(idParamStr)
	if err != nil {
		log.Info().
			Err(err).
			Str("id", idParamStr).
			Str("source", "internal.adapter.userHandler.DeleteUser").
			Msg("invalid user ID")

		return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID")
	}

	err = u.userService.DeleteUser(ctx, id)
	if err != nil {
		log.Error().
			Err(err).
			Int64("user_id", id).
			Str("source", "internal.adapter.userHandler.DeleteUser").
			Msg("failed delete user")

		if err.Error() == "404" {
			return fiber.NewError(fiber.StatusNotFound, "User not found")
		}

		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.DefaultResponse{
		Message: "user deleted successfully",
		Data:    nil,
	})
}

func (u *userHandler) GetUserAll(c fiber.Ctx) error {
	ctx := c.Context()

	user, ok := c.Locals("user").(string)
	if !ok || user == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "data token not valid")
	}

	search := c.Query("search")

	orderBy := c.Query("order_by", "created_at")

	orderType := c.Query("order_type", "desc")
	if orderType != "asc" && orderType != "desc" {
		orderType = "desc"
	}

	page, err := utils.StringToInt64(c.Query("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := utils.StringToInt64(c.Query("limit", "10"))
	if err != nil || limit <= 0 {
		limit = 10
	}

	reqEntity := entity.QueryStringUser{
		Search:    search,
		Page:      page,
		Limit:     limit,
		OrderBy:   orderBy,
		OrderType: orderType,
	}

	results, countData, totalPages, err := u.userService.GetUserAll(ctx, reqEntity)
	if err != nil {
		log.Error().
			Err(err).
			Str("search", search).
			Int64("page", page).
			Int64("limit", limit).
			Str("source", "internal.adapter.userHandler.GetUserAll").
			Msg("failed get user list")

		if err.Error() == "404" {
			return fiber.NewError(fiber.StatusNotFound, "data not found")
		}

		return err
	}

	respUser := make([]response.UserResponse, 0, len(results))

	for _, val := range results {
		respUser = append(respUser, response.UserResponse{
			ID:    val.ID,
			Name:  val.Name,
			Email: val.Email,
			Photo: val.Photo,
			Phone: val.Phone,
		})
	}

	return c.Status(fiber.StatusOK).JSON(
		response.DefaultResponseWithPaginations{
			Message: "data retrieved successfully",
			Data:    respUser,
			Pagination: &response.Pagination{
				Page:       page,
				TotalCount: countData,
				PerPage:    limit,
				TotalPage:  totalPages,
			},
		},
	)
}

func (u *userHandler) UpdatePassword(c fiber.Ctx) error {
	var req request.UpdatePasswordRequest

	ctx := c.Context()

	tokenString := c.Query("token")
	if tokenString == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "missing or invalid token")
	}

	if err := c.Bind().Body(&req); err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userHandler.UpdatePassword").
			Msg("failed bind/validate request")

		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if req.NewPassword != req.ConfirmPassword {
		log.Error().
			Str("source", "internal.adapter.userHandler.UpdatePassword").
			Msg("password confirmation mismatch")

		return fiber.NewError(fiber.StatusUnprocessableEntity, "new password and confirm password does not match")
	}

	reqEntity := entity.UserEntity{
		Password: req.NewPassword,
		Token:    tokenString,
	}

	if err := u.userService.UpdatePassword(ctx, reqEntity); err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userHandler.UpdatePassword").
			Msg("failed update password")

		if err.Error() == "404" {
			return fiber.NewError(fiber.StatusNotFound, "user not found")
		}

		if err.Error() == "401" {
			return fiber.NewError(fiber.StatusUnauthorized, "token expired or invalid")
		}

		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.DefaultResponse{
		Message: "password updated successfully",
		Data:    nil,
	})
}

func (u *userHandler) UpdateProfileUser(c fiber.Ctx) error {
	var (
		req         request.UpdateDataUserRequest
		jwtUserData entity.JwtUserData
	)

	ctx := c.Context()

	user, ok := c.Locals("user").(string)
	if !ok || user == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "data token not found")
	}

	if err := json.Unmarshal([]byte(user), &jwtUserData); err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userHandler.UpdateDataUser").
			Msg("failed parse jwt user data")

		return fiber.NewError(fiber.StatusBadRequest, "invalid token data")
	}

	if err := c.Bind().Body(&req); err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userHandler.UpdateDataUser").
			Msg("failed bind/validate request")

		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	reqEntity := entity.UserEntity{
		ID:      jwtUserData.UserID,
		Name:    req.Name,
		Email:   req.Email,
		Address: req.Address,
		Phone:   req.Phone,
		Photo:   req.Photo,
	}

	if err := u.userService.UpdateDataUser(ctx, reqEntity); err != nil {
		log.Error().
			Err(err).
			Int64("user_id", jwtUserData.UserID).
			Str("source", "internal.adapter.userHandler.UpdateDataUser").
			Msg("failed update user data")

		if err.Error() == "404" {
			return fiber.NewError(fiber.StatusNotFound, "user not found")
		}

		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.DefaultResponse{
		Message: "success",
		Data:    nil,
	})
}

func (u *userHandler) UpdateUser(c fiber.Ctx) error {
	var req request.UpdateDataUserRequest

	ctx := c.Context()

	user, ok := c.Locals("user").(string)
	if !ok || user == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "data token not valid")
	}

	if err := c.Bind().Body(&req); err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userHandler.UpdateCustomer").
			Msg("failed bind/validate request")

		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	idParamStr := c.Params("id")
	if idParamStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "missing or invalid customer ID")
	}

	id, err := utils.StringToInt64(idParamStr)
	if err != nil {
		log.Error().
			Err(err).
			Str("id", idParamStr).
			Str("source", "internal.adapter.userHandler.UpdateCustomer").
			Msg("invalid customer ID")

		return fiber.NewError(fiber.StatusBadRequest, "invalid customer ID")
	}

	phoneString := fmt.Sprintf("%v", req.Phone)

	reqEntity := entity.UserEntity{
		ID:      id,
		Name:    req.Name,
		Email:   req.Email,
		Phone:   phoneString,
		Address: req.Address,
		Photo:   req.Photo,
	}

	if err := u.userService.UpdateDataUser(ctx, reqEntity); err != nil {
		log.Error().
			Err(err).
			Int64("user_id", id).
			Str("source", "internal.adapter.userHandler.UpdateUser").
			Msg("failed update customer")

		if err.Error() == "404" {
			return fiber.NewError(fiber.StatusNotFound, "user not found")
		}

		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.DefaultResponse{
		Message: "success",
		Data:    nil,
	})
}
