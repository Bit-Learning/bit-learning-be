package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/lcaohoanq/bit-learning-be-v2/internal/modules/user"
)

type fakeUsers struct {
	account   user.Account
	createErr error
	findErr   error
}

func (f *fakeUsers) CreateAccount(_ context.Context, email, passwordHash, displayName string) (user.Account, error) {
	if f.createErr != nil {
		return user.Account{}, f.createErr
	}
	f.account = user.Account{User: user.User{ID: uuid.New(), Email: email, DisplayName: displayName}, PasswordHash: passwordHash}
	return f.account, nil
}

func (f *fakeUsers) FindAccountByEmail(context.Context, string) (user.Account, error) {
	return f.account, f.findErr
}

type fakeIssuer struct{}

func (fakeIssuer) Issue(uuid.UUID) (string, time.Time, error) {
	return "signed-token", time.Unix(123, 0).UTC(), nil
}

func TestServiceRegister(t *testing.T) {
	users := &fakeUsers{}
	result, err := NewService(users, fakeIssuer{}).Register(context.Background(), RegisterRequest{
		Email: " USER@example.com ", Password: "password123", DisplayName: " Learner ",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.AccessToken != "signed-token" || result.User.Email != "user@example.com" || result.User.DisplayName != "Learner" {
		t.Fatalf("result = %#v", result)
	}
	if bcrypt.CompareHashAndPassword([]byte(users.account.PasswordHash), []byte("password123")) != nil {
		t.Fatal("password was not hashed")
	}
}

func TestServiceRegisterDuplicateEmail(t *testing.T) {
	_, err := NewService(&fakeUsers{createErr: user.ErrEmailTaken}, fakeIssuer{}).Register(context.Background(), RegisterRequest{
		Email: "user@example.com", Password: "password123",
	})
	if !errors.Is(err, user.ErrEmailTaken) {
		t.Fatalf("error = %v", err)
	}
}

func TestServiceLoginWrongPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("right-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	users := &fakeUsers{account: user.Account{User: user.User{ID: uuid.New()}, PasswordHash: string(hash)}}
	_, err = NewService(users, fakeIssuer{}).Login(context.Background(), LoginRequest{Email: "user@example.com", Password: "wrong-password"})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("error = %v", err)
	}
}

func TestServiceLoginMissingUserHidesLookupResult(t *testing.T) {
	_, err := NewService(&fakeUsers{findErr: user.ErrUserNotFound}, fakeIssuer{}).Login(context.Background(), LoginRequest{
		Email: "user@example.com", Password: "password123",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("error = %v", err)
	}
}
