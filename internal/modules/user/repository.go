package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	db "github.com/lcaohoanq/bit-learning-be-v2/internal/database/db"
)

type SQLCRepository struct {
	queries db.Querier
}

func NewRepository(queries db.Querier) *SQLCRepository {
	return &SQLCRepository{queries: queries}
}

func (r *SQLCRepository) CreateAccount(ctx context.Context, params CreateAccountParams) (Account, error) {
	result, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		Email: params.Email, PasswordHash: params.PasswordHash, DisplayName: params.DisplayName,
	})
	if isUniqueViolation(err) {
		return Account{}, ErrEmailTaken
	}
	if err != nil {
		return Account{}, err
	}
	return accountFromDB(result), nil
}

func (r *SQLCRepository) FindAccountByEmail(ctx context.Context, email string) (Account, error) {
	result, err := r.queries.GetUserByEmail(ctx, email)
	return mapResult(result, err)
}

func (r *SQLCRepository) FindByID(ctx context.Context, id uuid.UUID) (Account, error) {
	result, err := r.queries.GetUserByID(ctx, id)
	return mapResult(result, err)
}

func (r *SQLCRepository) UpdateProfile(ctx context.Context, params UpdateProfileParams) (Account, error) {
	result, err := r.queries.UpdateUserProfile(ctx, db.UpdateUserProfileParams{
		ID: params.ID, DisplayName: params.DisplayName, Bio: params.Bio, AvatarUrl: params.AvatarURL,
	})
	return mapResult(result, err)
}

func mapResult(result db.User, err error) (Account, error) {
	if errors.Is(err, pgx.ErrNoRows) {
		return Account{}, ErrUserNotFound
	}
	if err != nil {
		return Account{}, err
	}
	return accountFromDB(result), nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func accountFromDB(value db.User) Account {
	return Account{
		User: User{ID: value.ID, Email: value.Email, DisplayName: value.DisplayName, Bio: value.Bio,
			AvatarURL: value.AvatarUrl, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt},
		PasswordHash: value.PasswordHash,
	}
}
