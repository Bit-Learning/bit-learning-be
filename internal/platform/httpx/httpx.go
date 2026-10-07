package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type ErrorDetail struct {
	Field string `json:"field"`
	Rule  string `json:"rule"`
}

type errorBody struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return errors.New("body must be valid JSON: " + err.Error())
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("body must contain a single JSON object")
	}
	return nil
}

func WriteData(w http.ResponseWriter, status int, data any) {
	writeJSON(w, status, map[string]any{"data": data})
}

func WriteError(w http.ResponseWriter, status int, code, message string, details ...ErrorDetail) {
	writeJSON(w, status, map[string]any{"error": errorBody{Code: code, Message: message, Details: details}})
}

func ValidationDetails(err error) []ErrorDetail {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return nil
	}
	details := make([]ErrorDetail, 0, len(validationErrors))
	for _, fieldError := range validationErrors {
		rule := fieldError.Tag()
		if rule == "minbytes" {
			rule = "min"
		} else if rule == "maxbytes" {
			rule = "max"
		} else if rule == "optional_url" {
			rule = "url"
		}
		details = append(details, ErrorDetail{Field: fieldError.Field(), Rule: rule})
	}
	return details
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
