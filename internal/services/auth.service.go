package services

import (
	"context"
	"errors"
	"strings"

	"github.com/muhammadwahyupratamaa/assessment-backend/internal/lib"
	"github.com/muhammadwahyupratamaa/assessment-backend/internal/models"
	"github.com/muhammadwahyupratamaa/assessment-backend/internal/repo"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrWeakPassword       = errors.New("password does not meet security requirements")
)

type AuthService interface {
	Register(
		ctx context.Context,
		request models.RegisterRequest,
	) (models.User, error)

	Login(
		ctx context.Context,
		request models.LoginRequest,
	) (models.User, error)

	GetProfile(
		ctx context.Context,
		userID string,
	) (models.User, error)
}

type authService struct {
	userRepo repo.UserRepo
}

func NewAuthService(userRepo repo.UserRepo) AuthService {
	return &authService{
		userRepo: userRepo,
	}
}

func (s *authService) Register(
	ctx context.Context,
	request models.RegisterRequest,
) (models.User, error) {
	email := strings.ToLower(strings.TrimSpace(request.Email))

	if !lib.IsStrongPassword(request.Password) {
		return models.User{}, ErrWeakPassword
	}

	_, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil {
		return models.User{}, ErrEmailAlreadyExists
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, err
	}

	passwordHash, err := lib.HashPassword(request.Password)
	if err != nil {
		return models.User{}, err
	}

	user, err := s.userRepo.Create(ctx, models.User{
		Name:         strings.TrimSpace(request.Name),
		Email:        email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return models.User{}, ErrEmailAlreadyExists
		}

		return models.User{}, err
	}

	return user, nil
}

func (s *authService) Login(
	ctx context.Context,
	request models.LoginRequest,
) (models.User, error) {
	email := strings.ToLower(strings.TrimSpace(request.Email))

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrInvalidCredentials
		}

		return models.User{}, err
	}

	isValid, err := lib.VerifyPassword(
		request.Password,
		user.PasswordHash,
	)
	if err != nil || !isValid {
		return models.User{}, ErrInvalidCredentials
	}

	return user, nil
}

func (s *authService) GetProfile(
	ctx context.Context,
	userID string,
) (models.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func isUniqueViolation(err error) bool {
	var pgError *pgconn.PgError

	return errors.As(err, &pgError) && pgError.Code == "23505"
}
