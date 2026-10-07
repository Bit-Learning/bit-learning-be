package user

import (
	"errors"
	"log/slog"
	"net/http"

	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/lcaohoanq/bit-learning-be-v2/internal/platform/httpx"
	platformmw "github.com/lcaohoanq/bit-learning-be-v2/internal/platform/middleware"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	id, ok := platformmw.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "authentication context is missing")
		return
	}
	profile, err := h.service.GetProfile(r.Context(), id)
	if errors.Is(err, ErrUserNotFound) {
		httpx.WriteError(w, http.StatusNotFound, "not_found", "user not found")
		return
	}
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	httpx.WriteData(w, http.StatusOK, profile)
}

func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	id, ok := platformmw.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "authentication context is missing")
		return
	}
	var input UpdateProfileRequest
	if err := httpx.DecodeJSON(w, r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	input.Normalize()
	if err := input.Validate(); err != nil {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "validation_error", "request validation failed", httpx.ValidationDetails(err)...)
		return
	}
	profile, err := h.service.UpdateProfile(r.Context(), id, ProfileChanges{
		DisplayName: input.DisplayName, Bio: input.Bio, AvatarURL: input.AvatarURL,
	})
	if errors.Is(err, ErrUserNotFound) {
		httpx.WriteError(w, http.StatusNotFound, "not_found", "user not found")
		return
	}
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	httpx.WriteData(w, http.StatusOK, profile)
}

func (h *Handler) internalError(w http.ResponseWriter, r *http.Request, err error) {
	h.logger.ErrorContext(r.Context(), "request failed", "error", err, "request_id", chimiddleware.GetReqID(r.Context()))
	httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "an internal error occurred")
}
