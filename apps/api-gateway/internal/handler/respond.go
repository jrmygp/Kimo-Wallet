package handler

import (
	"encoding/json"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type errorResponseBody struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, statusCode int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, errorResponseBody{Error: message})
}

// writeGRPCError translates an error returned by an internal gRPC service
// into an HTTP response. This translation is the extent of the gateway's
// "business logic" per docs/CLAUDE.md §4 — it does not decide what is
// valid, it only relays what the owning service already decided.
//
// Messages are only passed through for codes whose message text is a
// deliberately safe, user-facing string in the owning service (validation
// errors, conflicts, not-found). Anything else gets a fixed generic
// message so an internal/infrastructure error detail is never echoed to
// an HTTP client.
func writeGRPCError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	switch st.Code() {
	case codes.InvalidArgument:
		writeError(w, http.StatusBadRequest, st.Message())
	case codes.AlreadyExists:
		writeError(w, http.StatusConflict, st.Message())
	case codes.NotFound:
		writeError(w, http.StatusNotFound, st.Message())
	case codes.Unavailable, codes.DeadlineExceeded:
		writeError(w, http.StatusServiceUnavailable, "service temporarily unavailable")
	default:
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}
