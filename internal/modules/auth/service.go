package auth

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/lcaohoanq/bit-learning-be-v2/internal/modules/user"
)

type Service struct {
	users  Users
	tokens TokenIssuer
}

func NewService(users Users, tokens TokenIssuer) *Service {
	return &Service{users: users, tokens: tokens}
}

func (s *Service) Register(ctx context.Context, input RegisterRequest) (TokenResponse, error) {
	input.Normalize()
	if err := input.Validate(); err != nil {
		return TokenResponse{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return TokenResponse{}, err
	}
	account, err := s.users.CreateAccount(ctx, input.Email, string(hash), input.DisplayName)
	if err != nil {
		return TokenResponse{}, err
	}
	return s.issue(account)
}

func (s *Service) Login(ctx context.Context, input LoginRequest) (TokenResponse, error) {
	input.Normalize()
	if err := input.Validate(); err != nil {
		return TokenResponse{}, err
	}
	account, err := s.users.FindAccountByEmail(ctx, input.Email)
	if errors.Is(err, user.ErrUserNotFound) {
		return TokenResponse{}, ErrInvalidCredentials
	}
	if err != nil {
		return TokenResponse{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(input.Password)) != nil {
		return TokenResponse{}, ErrInvalidCredentials
	}
	return s.issue(account)
}

func (s *Service) issue(account user.Account) (TokenResponse, error) {
	token, expiresAt, err := s.tokens.Issue(account.ID)
	if err != nil {
		return TokenResponse{}, err
	}
	return TokenResponse{AccessToken: token, TokenType: "Bearer", ExpiresAt: expiresAt, User: account.User}, nil
}
