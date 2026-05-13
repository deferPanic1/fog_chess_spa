package service

import (
	"context"
	"errors"
	"time"

	"github.com/Gilf4/fog_chess/internal/domain/models"
	"github.com/Gilf4/fog_chess/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	users     repository.UserRepository
	secret    string
	accessTTL time.Duration
}

func NewAuthService(users repository.UserRepository, secret string, accessTTL time.Duration) *AuthService {
	return &AuthService{
		users:     users,
		secret:    secret,
		accessTTL: accessTTL,
	}
}

func (s *AuthService) Register(ctx context.Context, email, username, password string) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return s.users.Create(ctx, &models.User{
		Email:    email,
		Username: username,
		PassHash: hash,
		Rating:   1000,
		Role:     models.RoleUser,
	})
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, *models.User, error) {
	user, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", nil, ErrInvalidCredentials
		}
		return "", nil, err
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	token, err := s.generateJWT(user.ID, string(user.Role))
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

func (s *AuthService) generateJWT(userID int64, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secret))
}
