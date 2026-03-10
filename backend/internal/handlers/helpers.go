package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/kshitijdhara/blog/internal/domain"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, domain.ErrorResponse{
		Error: domain.ErrorBody{Code: code, Message: message},
	})
}

func decodeJSON(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return errors.New("request body is nil")
	}
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(v)
	if err != nil {
		slog.Error("json decode error", "error", err.Error())
	}
	return err
}

func queryInt(r *http.Request, key string, defaultVal int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}
