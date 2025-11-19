package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"gh-ts/internal/models"
	"gh-ts/internal/repository"
	"gh-ts/internal/utils"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	users         repository.UserRepository
	sessionSecret string
}

func NewAuthService(users repository.UserRepository, sessionSecret string) *AuthService {
	return &AuthService{users: users, sessionSecret: sessionSecret}
}

func (a *AuthService) Register(ctx context.Context, email, name, password string) (*models.User, error) {
	email = strings.TrimSpace(email)
	name = strings.TrimSpace(name)
	
	if email == "" || name == "" {
		return nil, errors.New("invalid input")
	}
	
	// Validate email format
	if !utils.ValidateEmail(email) {
		return nil, errors.New("invalid email format")
	}
	
	// Validate password strength
	if err := utils.ValidatePasswordStrength(password); err != nil {
		return nil, err
	}

	// Self-registration is only allowed for end users.
	role := "end_user"

	hash, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}
	return a.users.Create(ctx, email, name, role, hash)
}

func (a *AuthService) Login(ctx context.Context, email, password string) (token string, user *models.User, err error) {
	email = strings.TrimSpace(email)
	u, hash, err := a.users.GetByEmail(ctx, email)
	if err != nil {
		return "", nil, err
	}
	if u == nil {
		return "", nil, ErrInvalidCredentials
	}
	if !u.Active {
		return "", nil, ErrInvalidCredentials
	}
	if !utils.CheckPassword(hash, password) {
		return "", nil, ErrInvalidCredentials
	}
	tok, err := utils.SignJWT(a.sessionSecret, u.ID, u.Role, 24*time.Hour)
	if err != nil {
		return "", nil, err
	}
	return tok, u, nil
}
