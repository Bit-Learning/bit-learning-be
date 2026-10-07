package app

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/lcaohoanq/bit-learning-be-v2/internal/modules/auth"
	"github.com/lcaohoanq/bit-learning-be-v2/internal/modules/user"
)

type memoryUsers struct {
	byID    map[uuid.UUID]user.Account
	byEmail map[string]uuid.UUID
}

func newMemoryUsers() *memoryUsers {
	return &memoryUsers{byID: make(map[uuid.UUID]user.Account), byEmail: make(map[string]uuid.UUID)}
}

func (m *memoryUsers) CreateAccount(_ context.Context, params user.CreateAccountParams) (user.Account, error) {
	if _, exists := m.byEmail[params.Email]; exists {
		return user.Account{}, user.ErrEmailTaken
	}
	now := time.Now().UTC()
	account := user.Account{User: user.User{ID: uuid.New(), Email: params.Email, DisplayName: params.DisplayName, CreatedAt: now, UpdatedAt: now}, PasswordHash: params.PasswordHash}
	m.byID[account.ID], m.byEmail[account.Email] = account, account.ID
	return account, nil
}

func (m *memoryUsers) FindAccountByEmail(_ context.Context, email string) (user.Account, error) {
	id, ok := m.byEmail[email]
	if !ok {
		return user.Account{}, user.ErrUserNotFound
	}
	return m.byID[id], nil
}

func (m *memoryUsers) FindByID(_ context.Context, id uuid.UUID) (user.Account, error) {
	account, ok := m.byID[id]
	if !ok {
		return user.Account{}, user.ErrUserNotFound
	}
	return account, nil
}

func (m *memoryUsers) UpdateProfile(_ context.Context, params user.UpdateProfileParams) (user.Account, error) {
	account, ok := m.byID[params.ID]
	if !ok {
		return user.Account{}, user.ErrUserNotFound
	}
	account.DisplayName, account.Bio, account.AvatarURL = params.DisplayName, params.Bio, params.AvatarURL
	account.UpdatedAt = time.Now().UTC()
	m.byID[params.ID] = account
	return account, nil
}

func testRouter(repository user.Repository) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	users := user.NewService(repository)
	tokens := auth.NewTokenManager("01234567890123456789012345678901", time.Minute)
	return NewRouter(auth.NewHandler(auth.NewService(users, tokens), logger), user.NewHandler(users, logger), tokens, logger)
}

func perform(handler http.Handler, method, path, body, token string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func responseData(t *testing.T, response *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("response has no data envelope: %s", response.Body.String())
	}
	return data
}

func TestAuthAndProfileRoutes(t *testing.T) {
	handler := testRouter(newMemoryUsers())
	registered := perform(handler, http.MethodPost, "/v1/auth/register", `{"email":"USER@example.com","password":"password123","display_name":"Learner"}`, "")
	if registered.Code != http.StatusCreated {
		t.Fatalf("register status = %d, body = %s", registered.Code, registered.Body)
	}
	registerData := responseData(t, registered)
	if _, exposed := registerData["password_hash"]; exposed {
		t.Fatal("password hash was exposed")
	}
	token, ok := registerData["access_token"].(string)
	if !ok || token == "" {
		t.Fatalf("token = %#v", registerData["access_token"])
	}

	login := perform(handler, http.MethodPost, "/v1/auth/login", `{"email":"user@example.com","password":"password123"}`, "")
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", login.Code, login.Body)
	}
	responseData(t, login)

	profile := perform(handler, http.MethodGet, "/v1/users/me", "", token)
	if profile.Code != http.StatusOK {
		t.Fatalf("profile status = %d, body = %s", profile.Code, profile.Body)
	}
	if responseData(t, profile)["email"] != "user@example.com" {
		t.Fatalf("profile body = %s", profile.Body)
	}

	updated := perform(handler, http.MethodPatch, "/v1/users/me", `{"bio":"Learning Go"}`, token)
	if updated.Code != http.StatusOK {
		t.Fatalf("update status = %d, body = %s", updated.Code, updated.Body)
	}
	updatedData := responseData(t, updated)
	if updatedData["bio"] != "Learning Go" || updatedData["display_name"] != "Learner" {
		t.Fatalf("updated body = %s", updated.Body)
	}

	oldRoute := perform(handler, http.MethodGet, "/v1/me", "", token)
	if oldRoute.Code != http.StatusNotFound {
		t.Fatalf("old route status = %d", oldRoute.Code)
	}
}

func TestHandlerErrorMappings(t *testing.T) {
	handler := testRouter(newMemoryUsers())
	validBody := `{"email":"user@example.com","password":"password123"}`
	first := perform(handler, http.MethodPost, "/v1/auth/register", validBody, "")
	if first.Code != http.StatusCreated {
		t.Fatalf("first register status = %d", first.Code)
	}
	token := responseData(t, first)["access_token"].(string)

	tests := []struct {
		name, method, path, body, token string
		status                          int
		code                            string
	}{
		{"duplicate", http.MethodPost, "/v1/auth/register", validBody, "", http.StatusConflict, "email_taken"},
		{"wrong password", http.MethodPost, "/v1/auth/login", `{"email":"user@example.com","password":"wrongpass"}`, "", http.StatusUnauthorized, "invalid_credentials"},
		{"validation", http.MethodPost, "/v1/auth/register", `{"email":"invalid","password":"short"}`, "", http.StatusUnprocessableEntity, "validation_error"},
		{"missing token", http.MethodGet, "/v1/users/me", "", "", http.StatusUnauthorized, "unauthorized"},
		{"bad json", http.MethodPatch, "/v1/users/me", `{"unknown":true}`, token, http.StatusBadRequest, "invalid_request"},
		{"multiple json objects", http.MethodPatch, "/v1/users/me", `{"bio":"one"} {"bio":"two"}`, token, http.StatusBadRequest, "invalid_request"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := perform(handler, test.method, test.path, test.body, test.token)
			if response.Code != test.status || !strings.Contains(response.Body.String(), `"code":"`+test.code+`"`) {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body)
			}
		})
	}
	validation := perform(handler, http.MethodPost, "/v1/auth/register", `{"email":"invalid","password":"password123"}`, "")
	if !strings.Contains(validation.Body.String(), `"details":[{"field":"email","rule":"email"}]`) {
		t.Fatalf("validation details = %s", validation.Body)
	}
}

func TestProfileNotFound(t *testing.T) {
	repository := newMemoryUsers()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	users := user.NewService(repository)
	tokens := auth.NewTokenManager("01234567890123456789012345678901", time.Minute)
	handler := NewRouter(auth.NewHandler(auth.NewService(users, tokens), logger), user.NewHandler(users, logger), tokens, logger)
	token, _, err := tokens.Issue(uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	response := perform(handler, http.MethodGet, "/v1/users/me", "", token)
	if response.Code != http.StatusNotFound || !strings.Contains(response.Body.String(), `"code":"not_found"`) {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body)
	}
}
