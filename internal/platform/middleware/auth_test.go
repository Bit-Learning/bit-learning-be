package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

type stubParser struct {
	id  uuid.UUID
	err error
}

func (p stubParser) Parse(string) (uuid.UUID, error) { return p.id, p.err }

func TestAuthenticateRejectsMissingAndInvalidTokens(t *testing.T) {
	tests := []struct {
		name   string
		header string
		parser stubParser
	}{
		{"missing", "", stubParser{}},
		{"invalid", "Bearer invalid", stubParser{err: errors.New("expired")}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := Authenticate(test.parser)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				t.Fatal("next handler was called")
			}))
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.Header.Set("Authorization", test.header)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d", response.Code)
			}
		})
	}
}

func TestAuthenticateAddsUserIDToContext(t *testing.T) {
	want := uuid.New()
	handler := Authenticate(stubParser{id: want})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, ok := UserID(r.Context())
		if !ok || got != want {
			t.Fatalf("user ID = %s, present = %v", got, ok)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer valid")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d", response.Code)
	}
}
