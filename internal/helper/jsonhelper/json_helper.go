package jsonhelper

import (
	"briefcash-jwt/internal/dto"
	"encoding/json"
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
