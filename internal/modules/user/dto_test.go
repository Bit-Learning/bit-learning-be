package user

import (
	"strings"
	"testing"
)

func pointer(value string) *string { return &value }

func TestUpdateProfileRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		request UpdateProfileRequest
		valid   bool
	}{
		{"valid", UpdateProfileRequest{DisplayName: pointer("Learner"), Bio: pointer("Learning Go"), AvatarURL: pointer("https://example.com/avatar.png")}, true},
		{"display name too long", UpdateProfileRequest{DisplayName: pointer(strings.Repeat("a", 101))}, false},
		{"bio too long", UpdateProfileRequest{Bio: pointer(strings.Repeat("a", 1001))}, false},
		{"avatar too long", UpdateProfileRequest{AvatarURL: pointer("https://example.com/" + strings.Repeat("a", 2049))}, false},
		{"invalid avatar URL", UpdateProfileRequest{AvatarURL: pointer("not a URL")}, false},
		{"empty avatar clears value", UpdateProfileRequest{AvatarURL: pointer("")}, true},
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
