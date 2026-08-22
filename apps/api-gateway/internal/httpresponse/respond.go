// Package httpresponse writes the one JSON envelope every response from
// this gateway uses — success or error, from any handler or middleware.
// Extracted out of the handler package once a second consumer
// (middleware, for 401 responses) needed the same shape — see
// docs/agent-logs for that decision.
package httpresponse

import (
	"encoding/json"
	"net/http"
)

// Response is the envelope. `Data` is always present (explicit `null` on
// error, never an omitted key) so a client can rely on the shape without
// checking which branch it got.
type Response struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Data       any    `json:"data"`
}

func WriteJSON(w http.ResponseWriter, statusCode int, message string, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(Response{
		StatusCode: statusCode,
		Message:    message,
		Data:       data,
	})
}

func WriteError(w http.ResponseWriter, statusCode int, message string) {
	WriteJSON(w, statusCode, message, nil)
}
