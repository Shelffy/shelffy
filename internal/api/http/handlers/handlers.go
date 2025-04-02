package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type R = map[string]any

func getRequestData[T any](r *http.Request) (T, error) {
	var entity T
	err := json.NewDecoder(r.Body).Decode(&entity)
	return entity, err
}

func response(payload any, code int, w http.ResponseWriter) error {
	w.WriteHeader(code)
	messageBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = w.Write(messageBytes)
	return err
}

func errorResponse(errorMessage string, code int, w http.ResponseWriter) error {
	return response(R{"error": errorMessage}, code, w)
}

func successResponse(payload any, code int, w http.ResponseWriter) error {
	return response(R{"success": payload}, code, w)
}

func logResponseWriteError(err error, logger *slog.Logger) {
	if err != nil {
		logger.Error("failed to write response", "error", err)
	}
}
