package auth

import (
	"errors"
	"log/slog"
	"net/http"

	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/lcaohoanq/bit-learning-be-v2/internal/modules/user"
	"github.com/lcaohoanq/bit-learning-be-v2/internal/platform/httpx"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var input RegisterRequest
	if err := httpx.DecodeJSON(w, r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	result, err := h.service.Register(r.Context(), input)
	if details := httpx.ValidationDetails(err); len(details) > 0 {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "validation_error", "request validation failed", details...)
		return
	}
	if errors.Is(err, user.ErrEmailTaken) {
		httpx.WriteError(w, http.StatusConflict, "email_taken", "email is already registered")
		return
	}
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	httpx.WriteData(w, http.StatusCreated, result)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input LoginRequest
	if err := httpx.DecodeJSON(w, r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	result, err := h.service.Login(r.Context(), input)
	if details := httpx.ValidationDetails(err); len(details) > 0 {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "validation_error", "request validation failed", details...)
		return
	}
	if errors.Is(err, ErrInvalidCredentials) {
		httpx.WriteError(w, http.StatusUnauthorized, "invalid_credentials", "email or password is incorrect")
		return
	}
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	httpx.WriteData(w, http.StatusOK, result)
}

func (h *Handler) internalError(w http.ResponseWriter, r *http.Request, err error) {
	h.logger.ErrorContext(r.Context(), "request failed", "error", err, "request_id", chimiddleware.GetReqID(r.Context()))
	httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "an internal error occurred")
}
