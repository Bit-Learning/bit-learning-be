package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrEmailTaken   = errors.New("email is already registered")
	ErrUserNotFound = errors.New("user not found")
)

type CreateAccountParams struct {
	Email        string
	PasswordHash string
	DisplayName  string
}

type UpdateProfileParams struct {
	ID          uuid.UUID
	DisplayName string
	Bio         string
	AvatarURL   string
}

type Repository interface {
	CreateAccount(ctx context.Context, params CreateAccountParams) (Account, error)
	FindAccountByEmail(ctx context.Context, email string) (Account, error)
	FindByID(ctx context.Context, id uuid.UUID) (Account, error)
	UpdateProfile(ctx context.Context, params UpdateProfileParams) (Account, error)
}

type AccountService interface {
	CreateAccount(ctx context.Context, email, passwordHash, displayName string) (Account, error)
	FindAccountByEmail(ctx context.Context, email string) (Account, error)
}
