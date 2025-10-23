package jsonhelper

import (
	"briefcash-jwt/internal/dto"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func WriteJson(w http.ResponseWriter, httpStatus int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(data)
}

func WriteJsonError(w http.ResponseWriter, httpStatus int, message string) {
	WriteJson(w, httpStatus, dto.JwtDataResponse{
		Status:  false,
		Message: message,
	})
}

func ParseJsonBody(r *http.Request, destination interface{}) ([]byte, error) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body request: %w", err)
	}
	r.Body.Close()
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	if err := json.Unmarshal(bodyBytes, destination); err != nil {
		return nil, fmt.Errorf("invalid json format: %w", err)
	}
	return bodyBytes, nil
}
