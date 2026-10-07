package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/lcaohoanq/bit-learning-be-v2/internal/modules/user"
)

var ErrInvalidCredentials = errors.New("email or password is incorrect")

type Users interface {
	CreateAccount(ctx context.Context, email, passwordHash, displayName string) (user.Account, error)
	FindAccountByEmail(ctx context.Context, email string) (user.Account, error)
}

type TokenIssuer interface {
	Issue(userID uuid.UUID) (string, time.Time, error)
}
