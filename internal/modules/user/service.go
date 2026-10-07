package user

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) CreateAccount(ctx context.Context, email, passwordHash, displayName string) (Account, error) {
	return s.repository.CreateAccount(ctx, CreateAccountParams{
		Email: strings.ToLower(strings.TrimSpace(email)), PasswordHash: passwordHash, DisplayName: strings.TrimSpace(displayName),
	})
}

func (s *Service) FindAccountByEmail(ctx context.Context, email string) (Account, error) {
	return s.repository.FindAccountByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
}

func (s *Service) GetProfile(ctx context.Context, id uuid.UUID) (User, error) {
	account, err := s.repository.FindByID(ctx, id)
	return account.User, err
}

func (s *Service) UpdateProfile(ctx context.Context, id uuid.UUID, changes ProfileChanges) (User, error) {
	current, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return User{}, err
	}
	displayName, bio, avatarURL := current.DisplayName, current.Bio, current.AvatarURL
	if changes.DisplayName != nil {
		displayName = *changes.DisplayName
	}
	if changes.Bio != nil {
		bio = *changes.Bio
	}
	if changes.AvatarURL != nil {
		avatarURL = *changes.AvatarURL
	}
	updated, err := s.repository.UpdateProfile(ctx, UpdateProfileParams{
		ID: id, DisplayName: displayName, Bio: bio, AvatarURL: avatarURL,
	})
	return updated.User, err
}
