package api

import (
	"encoding/json"
	"net/http"
)

func JSONResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// The status line is already sent; an encoding failure cannot be reported to the client.
	_ = json.NewEncoder(w).Encode(data)
}

func OKResponse(w http.ResponseWriter, data any) {
	JSONResponse(w, http.StatusOK, data)
}

func ErrorResponse(w http.ResponseWriter, status int, message string) {
	JSONResponse(w, status, map[string]string{"error": message})
}
