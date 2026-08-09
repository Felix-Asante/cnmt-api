package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type errorBody struct {
	Error  string       `json:"error"`
	Fields []FieldError `json:"fields,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")

	var ve *ValidationError
	if errors.As(err, &ve) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(errorBody{
			Error:  "validation_failed",
			Fields: ve.Fields,
		})
		return
	}

	if status == 0 || status == http.StatusInternalServerError {
		status = StatusFromError(err)
	}

	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorBody{Error: PublicMessage(err)})
}

func StatusFromError(err error) int {
	switch {
	case errors.Is(err, NotFoundError):
		return http.StatusNotFound
	case errors.Is(err, ConflictError):
		return http.StatusConflict
	case errors.Is(err, UnauthorizedError):
		return http.StatusUnauthorized
	case errors.Is(err, ForbiddenError):
		return http.StatusForbidden
	case errors.Is(err, TooManyRequestsError):
		return http.StatusTooManyRequests
	case errors.Is(err, BadGatewayError):
		return http.StatusBadGateway
	case errors.Is(err, ServiceUnavailableError):
		return http.StatusServiceUnavailable
	case errors.Is(err, GatewayTimeoutError):
		return http.StatusGatewayTimeout
	case errors.Is(err, BadRequestError),
		errors.Is(err, InvalidJSONError),
		errors.Is(err, UnprocessableEntityError):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func PublicMessage(err error) string {
	if err == nil {
		return InternalServerError.Error()
	}

	switch {
	case errors.Is(err, BadRequestError):
		return trimSentinelPrefix(err, BadRequestError)
	case errors.Is(err, UnprocessableEntityError):
		return trimSentinelPrefix(err, UnprocessableEntityError)
	case errors.Is(err, InvalidJSONError):
		return InvalidJSONError.Error()
	case errors.Is(err, NotFoundError):
		return trimSentinelPrefix(err, NotFoundError)
	case errors.Is(err, ConflictError):
		return ConflictError.Error()
	case errors.Is(err, UnauthorizedError):
		return UnauthorizedError.Error()
	case errors.Is(err, ForbiddenError):
		return ForbiddenError.Error()
	case errors.Is(err, TooManyRequestsError):
		return TooManyRequestsError.Error()
	case errors.Is(err, BadGatewayError):
		return BadGatewayError.Error()
	case errors.Is(err, ServiceUnavailableError):
		return ServiceUnavailableError.Error()
	case errors.Is(err, GatewayTimeoutError):
		return GatewayTimeoutError.Error()
	case errors.Is(err, NetworkAuthenticationRequiredError):
		return NetworkAuthenticationRequiredError.Error()
	case errors.Is(err, InternalServerError):
		return InternalServerError.Error()
	default:
		// Unmapped errors: do not leak internals.
		return InternalServerError.Error()
	}
}

func trimSentinelPrefix(err, sentinel error) string {
	msg := err.Error()
	prefix := sentinel.Error() + ": "
	if strings.HasPrefix(msg, prefix) {
		return strings.TrimPrefix(msg, prefix)
	}
	if msg == sentinel.Error() {
		return sentinel.Error()
	}
	// fmt.Errorf("%w: detail") already handled; fallback to full message if it's the sentinel alone.
	return msg
}
