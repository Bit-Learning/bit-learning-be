package user

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type fakeRepository struct {
	account     Account
	findErr     error
	updateErr   error
	updateInput UpdateProfileParams
}

func (f *fakeRepository) CreateAccount(_ context.Context, params CreateAccountParams) (Account, error) {
	f.account = Account{User: User{ID: uuid.New(), Email: params.Email, DisplayName: params.DisplayName}, PasswordHash: params.PasswordHash}
	return f.account, nil
}
func (f *fakeRepository) FindAccountByEmail(context.Context, string) (Account, error) {
	return f.account, f.findErr
}
func (f *fakeRepository) FindByID(context.Context, uuid.UUID) (Account, error) {
	return f.account, f.findErr
}
func (f *fakeRepository) UpdateProfile(_ context.Context, params UpdateProfileParams) (Account, error) {
	f.updateInput = params
	if f.updateErr != nil {
		return Account{}, f.updateErr
	}
	f.account.DisplayName, f.account.Bio, f.account.AvatarURL = params.DisplayName, params.Bio, params.AvatarURL
	return f.account, nil
}

func TestServiceGetProfile(t *testing.T) {
	id := uuid.New()
	repository := &fakeRepository{account: Account{User: User{ID: id, Email: "user@example.com"}, PasswordHash: "secret"}}
	profile, err := NewService(repository).GetProfile(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if profile.ID != id || profile.Email != "user@example.com" {
		t.Fatalf("profile = %#v", profile)
	}
}

func TestServiceGetProfileNotFound(t *testing.T) {
	_, err := NewService(&fakeRepository{findErr: ErrUserNotFound}).GetProfile(context.Background(), uuid.New())
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("error = %v", err)
	}
}

func TestServiceUpdateProfilePreservesOmittedFields(t *testing.T) {
	id := uuid.New()
	bio := "updated bio"
	repository := &fakeRepository{account: Account{User: User{ID: id, DisplayName: "Existing", Bio: "old", AvatarURL: "https://example.com/a.png"}}}
	profile, err := NewService(repository).UpdateProfile(context.Background(), id, ProfileChanges{Bio: &bio})
	if err != nil {
		t.Fatal(err)
	}
	if profile.DisplayName != "Existing" || profile.Bio != bio || profile.AvatarURL != "https://example.com/a.png" {
		t.Fatalf("profile = %#v", profile)
	}
}
