package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
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

	msg := InternalServerError.Error()
	if err != nil {
		msg = err.Error()
	}

	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorBody{Error: msg})
}
