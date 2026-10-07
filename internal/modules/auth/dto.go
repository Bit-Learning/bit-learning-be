package auth

import (
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/lcaohoanq/bit-learning-be-v2/internal/modules/user"
)

var validate = newValidator()

type RegisterRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,minbytes=8,maxbytes=72"`
	DisplayName string `json:"display_name" validate:"max=100"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,minbytes=8,maxbytes=72"`
}

type TokenResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
	User        user.User `json:"user"`
}

func (r *RegisterRequest) Normalize() {
	r.Email = strings.ToLower(strings.TrimSpace(r.Email))
	r.DisplayName = strings.TrimSpace(r.DisplayName)
}

func (r RegisterRequest) Validate() error { return validate.Struct(r) }

func (r *LoginRequest) Normalize() {
	r.Email = strings.ToLower(strings.TrimSpace(r.Email))
}

func (r LoginRequest) Validate() error { return validate.Struct(r) }

func newValidator() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		return strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
	})
	_ = v.RegisterValidation("minbytes", func(fl validator.FieldLevel) bool {
		return len([]byte(fl.Field().String())) >= 8
	})
	_ = v.RegisterValidation("maxbytes", func(fl validator.FieldLevel) bool {
		return len([]byte(fl.Field().String())) <= 72
	})
	return v
}
