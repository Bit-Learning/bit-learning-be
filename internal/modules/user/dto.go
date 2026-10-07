package user

import (
	"net/url"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	_ = v.RegisterValidation("optional_url", func(field validator.FieldLevel) bool {
		value := field.Field().String()
		if value == "" {
			return true
		}
		parsed, err := url.ParseRequestURI(value)
		return err == nil && parsed.Scheme != "" && parsed.Host != ""
	})
	return v
}

type UpdateProfileRequest struct {
	DisplayName *string `json:"display_name" validate:"omitempty,max=100"`
	Bio         *string `json:"bio" validate:"omitempty,max=1000"`
	AvatarURL   *string `json:"avatar_url" validate:"omitempty,max=2048,optional_url"`
}

func (r *UpdateProfileRequest) Normalize() {
	trim(r.DisplayName)
	trim(r.Bio)
	trim(r.AvatarURL)
}

func (r UpdateProfileRequest) Validate() error {
	return validate.Struct(r)
}

func trim(value *string) {
	if value != nil {
		*value = strings.TrimSpace(*value)
	}
}
