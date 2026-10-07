package auth

import (
	"strings"
	"testing"
)

func TestRegisterRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		request RegisterRequest
		valid   bool
	}{
		{"valid", RegisterRequest{Email: "USER@example.com ", Password: "password123", DisplayName: " Learner "}, true},
		{"invalid email", RegisterRequest{Email: "not-an-email", Password: "password123"}, false},
		{"short password", RegisterRequest{Email: "user@example.com", Password: "short"}, false},
		{"long password", RegisterRequest{Email: "user@example.com", Password: strings.Repeat("a", 73)}, false},
		{"multibyte password over bcrypt limit", RegisterRequest{Email: "user@example.com", Password: strings.Repeat("ê", 37)}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.request.Normalize()
			err := test.request.Validate()
			if (err == nil) != test.valid {
				t.Fatalf("Validate() error = %v, valid = %v", err, test.valid)
			}
		})
	}
}

func TestRegisterRequestNormalize(t *testing.T) {
	request := RegisterRequest{Email: " USER@Example.COM ", DisplayName: " Learner "}
	request.Normalize()
	if request.Email != "user@example.com" || request.DisplayName != "Learner" {
		t.Fatalf("normalized request = %#v", request)
	}
}
